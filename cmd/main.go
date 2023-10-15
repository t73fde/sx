//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
	"zettelstore.de/sx.fossil/sxreader"
)

type mainParserExecutor struct {
	origParser   sxeval.Parser
	origExecutor sxeval.Executor
	logReader    bool
	logParser    bool
	logExpr      bool
	logExecutor  bool
}

func (mpe *mainParserExecutor) Parse(pf *sxeval.ParseFrame, form sx.Object) (sxeval.Expr, error) {
	if !mpe.logParser {
		return mpe.origParser.Parse(pf, form)
	}
	env := pf.Environment()
	fmt.Printf(";P %v<-%v %T %v\n", env, env.Parent(), form, form)
	expr, err := mpe.origParser.Parse(pf, form)
	if err != nil {
		return nil, err
	}
	fmt.Printf(";Q ")
	expr.Print(os.Stdout)
	fmt.Println()
	return expr, nil
}

func (mpe *mainParserExecutor) Execute(frame *sxeval.Frame, expr sxeval.Expr) (sx.Object, error) {
	if !mpe.logExecutor {
		return mpe.origExecutor.Execute(frame, expr)
	}
	env := frame.Environment()
	fmt.Printf(";X %v<-%v ", env, env.Parent())
	expr.Print(os.Stdout)
	fmt.Println()
	obj, err := mpe.origExecutor.Execute(frame, expr)
	if err != nil {
		return nil, err
	}
	fmt.Printf(";O %T %v\n", obj, obj)
	return obj, nil
}

var syntaxes = []struct {
	name string
	fn   sxeval.SyntaxFn
}{
	{"define", sxbuiltins.DefineS},
	{"defvar", sxbuiltins.DefVarS},
	{"defconst", sxbuiltins.DefConstS},
	{"set!", sxbuiltins.SetXS},
	{"if", sxbuiltins.IfS},
	{"defun", sxbuiltins.DefunS}, {"lambda", sxbuiltins.LambdaS},
	{"defmacro", sxbuiltins.DefMacroS},
}

var builtinsA = []struct {
	name string
	fn   sxeval.BuiltinAold
}{
	{"==", sxbuiltins.IdenticalOld}, {"=", sxbuiltins.EqualOld},
	{"number?", sxbuiltins.NumberPold},
	{"+", sxbuiltins.AddOld}, {"-", sxbuiltins.SubOld}, {"*", sxbuiltins.MulOld},
	{"div", sxbuiltins.DivOld}, {"mod", sxbuiltins.ModOld},
	{"<", sxbuiltins.NumLessOld}, {"<=", sxbuiltins.NumLessEqualOld},
	{">=", sxbuiltins.NumGreaterEqualOld}, {">", sxbuiltins.NumGreaterOld},
	{"min", sxbuiltins.MinOld}, {"max", sxbuiltins.MaxOld},
	{"cons", sxbuiltins.ConsOld}, {"pair?", sxbuiltins.PairPold},
	{"null?", sxbuiltins.NullPold}, {"list?", sxbuiltins.ListPold},
	{"car", sxbuiltins.CarOld}, {"cdr", sxbuiltins.CdrOld},
	{"caar", sxbuiltins.CaarOld}, {"cadr", sxbuiltins.CadrOld}, {"cdar", sxbuiltins.CdarOld}, {"cddr", sxbuiltins.CddrOld},
	{"caaar", sxbuiltins.CaaarOld}, {"caadr", sxbuiltins.CaadrOld}, {"cadar", sxbuiltins.CadarOld}, {"caddr", sxbuiltins.CaddrOld},
	{"cdaar", sxbuiltins.CdaarOld}, {"cdadr", sxbuiltins.CdadrOld}, {"cddar", sxbuiltins.CddarOld}, {"cdddr", sxbuiltins.CdddrOld},
	{"caaaar", sxbuiltins.CaaaarOld}, {"caaadr", sxbuiltins.CaaadrOld}, {"caadar", sxbuiltins.CaadarOld}, {"caaddr", sxbuiltins.CaaddrOld},
	{"cadaar", sxbuiltins.CadaarOld}, {"cadadr", sxbuiltins.CadadrOld}, {"caddar", sxbuiltins.CaddarOld}, {"cadddr", sxbuiltins.CadddrOld},
	{"cdaaar", sxbuiltins.CdaaarOld}, {"cdaadr", sxbuiltins.CdaadrOld}, {"cdadar", sxbuiltins.CdadarOld}, {"cdaddr", sxbuiltins.CdaddrOld},
	{"cddaar", sxbuiltins.CddaarOld}, {"cddadr", sxbuiltins.CddadrOld}, {"cdddar", sxbuiltins.CdddarOld}, {"cddddr", sxbuiltins.CddddrOld},
	{"last", sxbuiltins.LastOld},
	{"list", sxbuiltins.ListOld}, {"list*", sxbuiltins.ListStarOld},
	{"append", sxbuiltins.AppendOld}, {"reverse", sxbuiltins.ReverseOld},
	{"length", sxbuiltins.LengthOld}, {"assoc", sxbuiltins.AssocOld},
	{"->string", sxbuiltins.ToStringOld}, {"string-append", sxbuiltins.StringAppendOld},
	{"callable?", sxbuiltins.CallablePold},
	{"parent-environment", sxbuiltins.ParentEnvOld},
	{"environment-bindings", sxbuiltins.EnvBindingsOld},
	{"undefined?", sxbuiltins.UndefinedPold}, {"defined?", sxbuiltins.DefinedPold},
}
var builtinsFA = []struct {
	name string
	fn   sxeval.BuiltinFAold
}{
	{"map", sxbuiltins.MapOld}, {"apply", sxbuiltins.ApplyOld},
	{"fold", sxbuiltins.FoldOld}, {"fold-reverse", sxbuiltins.FoldReverseOld},
	{"current-environment", sxbuiltins.CurrentEnvOld},
	{"bound?", sxbuiltins.BoundPold},
	{"environment-lookup", sxbuiltins.EnvLookupOld}, {"environment-resolve", sxbuiltins.EnvResolveOld},
	{"macroexpand-0", sxbuiltins.MacroExpand0old},
	{"pp", sxbuiltins.PrettyOld},
}

func main() {
	sf := sx.MakeMappedFactory(1024)
	rd := sxreader.MakeReader(os.Stdin, sxreader.WithSymbolFactory(sf))
	symQQ, symUQ, symUQS := installQQ(rd)

	mpe := mainParserExecutor{}
	engine := sxeval.MakeEngine(sf, sxeval.MakeRootEnvironment(len(syntaxes)+len(builtinsA)+len(builtinsFA)+16))
	engine.SetQuote(nil)
	mpe.origParser = engine.SetParser(&mpe)
	mpe.origExecutor = engine.SetExecutor(&mpe)
	root := engine.RootEnvironment()
	sxbuiltins.InstallQuasiQuoteSyntax(root, symQQ, symUQ, symUQS)
	for _, synDef := range syntaxes {
		engine.BindSyntax(synDef.name, synDef.fn)
	}
	for _, bDef := range builtinsA {
		engine.BindBuiltinAold(bDef.name, bDef.fn)
	}
	for _, bDef := range builtinsFA {
		engine.BindBuiltinFAold(bDef.name, bDef.fn)
	}
	engine.Bind("UNDEFINED", sx.MakeUndefined())
	engine.BindBuiltinAold("log-reader", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logReader
		mpe.logReader = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinAold("log-parser", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logParser
		mpe.logParser = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinAold("log-expr", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logExpr
		mpe.logExpr = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinAold("log-executor", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logExecutor
		mpe.logExecutor = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinAold("log-off", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		mpe.logReader = false
		mpe.logParser = false
		mpe.logExecutor = false
		return sx.Nil(), nil
	})
	engine.BindBuiltinAold("panic", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 1)
		if err != nil {
			panic(err)
		}
		if len(args) == 0 {
			panic("common panic")
		}
		panic(args[0])
	})
	err := readPrelude(engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read prelude: %v\n", err)
		os.Exit(17)
	}
	root.Freeze()
	env := sxeval.MakeChildEnvironment(engine.GetToplevelEnv(), "repl", 1024)
	env.Bind(sf.MustMake("root-env"), root)
	env.Bind(sf.MustMake("repl-env"), env)

	mpe.logReader = true
	mpe.logParser = true
	mpe.logExpr = false
	mpe.logExecutor = true

	var wg sync.WaitGroup
	wg.Add(1)
	go repl(rd, &mpe, engine, env, &wg)
	wg.Wait()
}

func repl(rd *sxreader.Reader, mpe *mainParserExecutor, eng *sxeval.Engine, env sxeval.Environment, wg *sync.WaitGroup) {
	defer func() {
		if val := recover(); val != nil {
			stack := debug.Stack()
			fmt.Printf("RECOVER PANIC: %v\n\n%s\n", val, string(stack))
			go repl(rd, mpe, eng, env, wg)
			return
		}
		wg.Done()
	}()

	for {
		fmt.Print("> ")
		obj, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(";r", err)
			continue
		}
		if mpe.logReader {
			fmt.Println(";<", obj)
		}
		expr, err := eng.Parse(env, obj)
		if err != nil {
			fmt.Println(";p", err)
			continue
		}
		expr = eng.Rework(env, expr)
		if mpe.logExpr {
			printExpr(eng, expr, 0)
			continue
		} else if mpe.logReader {
			fmt.Printf(";= ")
			expr.Print(os.Stdout)
			fmt.Println()
		}
		res, err := eng.Execute(env, expr)
		if err != nil {
			fmt.Println(";e", err)
			continue
		}
		fmt.Println(sx.Repr(res))
	}
}

func printExpr(eng *sxeval.Engine, expr sxeval.Expr, level int) {
	if level <= 0 {
		level = -level
	} else {
		fmt.Print(strings.Repeat(" ", level*2))
	}

	switch e := expr.(type) {
	case *sxeval.BuiltinCallExprOld:
		fmt.Printf("B-CALL %v\n", e.Proc.Name(eng))
		for _, arg := range e.Args {
			printExpr(eng, arg, level+1)
		}
	case *sxeval.CallExpr:
		fmt.Println("CALL")
		printExpr(eng, e.Proc, level+1)
		for _, arg := range e.Args {
			printExpr(eng, arg, level+1)
		}
	case sxeval.ResolveExpr:
		fmt.Printf("RESOLVE %v\n", e.Symbol)
	case sxeval.ObjExpr:
		fmt.Printf("OBJ %T/%v\n", e.Obj, e.Obj)
	case *sxbuiltins.LambdaExpr:
		fmt.Printf("LAMBDA %q", e.Name)
		for _, sym := range e.Params {
			fmt.Printf(" %v", sym)
		}
		if e.Rest != nil {
			fmt.Printf(" . %v", e.Rest)
		}
		fmt.Println()
		for _, ex := range e.Front {
			printExpr(eng, ex, level+1)
		}
		printExpr(eng, e.Last, level+1)
	case *sxbuiltins.IfExpr:
		fmt.Println("IF")
		printExpr(eng, e.Test, level+1)
		printExpr(eng, e.True, level+1)
		printExpr(eng, e.False, level+1)
	case *sxbuiltins.DefineExpr:
		if e.Const {
			fmt.Println("DEFCONST", e.Sym)
		} else {
			fmt.Println("DEFVAR", e.Sym)
		}
		printExpr(eng, e.Val, level+1)
	case *sxbuiltins.SetXExpr:
		fmt.Println("SET!", e.Sym)
		printExpr(eng, e.Val, level+1)
	case sxbuiltins.MakeListExpr:
		fmt.Println("MAKELIST")
		printExpr(eng, e.Elem, level+1)
	default:
		switch e {
		case sxeval.NilExpr:
			fmt.Println("NIL")
		default:
			fmt.Printf("%T/%v\n", expr, expr)
		}
	}
}

func installQQ(rd *sxreader.Reader) (*sx.Symbol, *sx.Symbol, *sx.Symbol) {
	sf := rd.SymbolFactory()
	symQQ, symUQ, symUQS := sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")
	sxbuiltins.InstallQuasiQuoteReader(rd, symQQ, '`', symUQ, ',', symUQS, '@')
	return symQQ, symUQ, symUQS
}

//go:embed prelude.sxn
var prelude string

func readPrelude(engine *sxeval.Engine) error {
	rd := sxreader.MakeReader(strings.NewReader(prelude), sxreader.WithSymbolFactory(engine.SymbolFactory()))
	installQQ(rd)
	for {
		form, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err = engine.Eval(engine.RootEnvironment(), form)
		if err != nil {
			return err
		}
	}
}
