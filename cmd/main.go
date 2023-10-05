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
	{"define", sxbuiltins.DefineS}, {"set!", sxbuiltins.SetXS},
	{"if", sxbuiltins.IfS},
	{"begin", sxbuiltins.BeginS},
	{"and", sxbuiltins.AndS}, {"or", sxbuiltins.OrS},
	{"lambda", sxbuiltins.LambdaS},
	{"let", sxbuiltins.LetS},
	{"timeit", sxbuiltins.TimeitS},
	{"defmacro", sxbuiltins.DefMacroS},
}

var builtinsA = []struct {
	name string
	fn   sxeval.BuiltinA
}{
	{"==", sxbuiltins.Identical}, {"=", sxbuiltins.Equal},
	{"boolean", sxbuiltins.Boolean}, {"not", sxbuiltins.Not},
	{"number?", sxbuiltins.NumberP},
	{"+", sxbuiltins.Add}, {"-", sxbuiltins.Sub}, {"*", sxbuiltins.Mul},
	{"div", sxbuiltins.Div}, {"mod", sxbuiltins.Mod},
	{"<", sxbuiltins.NumLess}, {"<=", sxbuiltins.NumLessEqual},
	{">=", sxbuiltins.NumGreaterEqual}, {">", sxbuiltins.NumGreater},
	{"min", sxbuiltins.Min}, {"max", sxbuiltins.Max},
	{"cons", sxbuiltins.Cons}, {"pair?", sxbuiltins.PairP},
	{"null?", sxbuiltins.NullP}, {"list?", sxbuiltins.ListP},
	{"car", sxbuiltins.Car}, {"cdr", sxbuiltins.Cdr},
	{"caar", sxbuiltins.Caar}, {"cadr", sxbuiltins.Cadr}, {"cdar", sxbuiltins.Cdar}, {"cddr", sxbuiltins.Cddr},
	{"caaar", sxbuiltins.Caaar}, {"caadr", sxbuiltins.Caadr}, {"cadar", sxbuiltins.Cadar}, {"caddr", sxbuiltins.Caddr},
	{"cdaar", sxbuiltins.Cdaar}, {"cdadr", sxbuiltins.Cdadr}, {"cddar", sxbuiltins.Cddar}, {"cdddr", sxbuiltins.Cdddr},
	{"caaaar", sxbuiltins.Caaaar}, {"caaadr", sxbuiltins.Caaadr}, {"caadar", sxbuiltins.Caadar}, {"caaddr", sxbuiltins.Caaddr},
	{"cadaar", sxbuiltins.Cadaar}, {"cadadr", sxbuiltins.Cadadr}, {"caddar", sxbuiltins.Caddar}, {"cadddr", sxbuiltins.Cadddr},
	{"cdaaar", sxbuiltins.Cdaaar}, {"cdaadr", sxbuiltins.Cdaadr}, {"cdadar", sxbuiltins.Cdadar}, {"cdaddr", sxbuiltins.Cdaddr},
	{"cddaar", sxbuiltins.Cddaar}, {"cddadr", sxbuiltins.Cddadr}, {"cdddar", sxbuiltins.Cdddar}, {"cddddr", sxbuiltins.Cddddr},
	{"last", sxbuiltins.Last},
	{"list", sxbuiltins.List}, {"list*", sxbuiltins.ListStar},
	{"append", sxbuiltins.Append}, {"reverse", sxbuiltins.Reverse},
	{"length", sxbuiltins.Length}, {"assoc", sxbuiltins.Assoc},
	{"->string", sxbuiltins.ToString}, {"string-append", sxbuiltins.StringAppend},
	{"callable?", sxbuiltins.CallableP},
	{"parent-environment", sxbuiltins.ParentEnv},
	{"environment-bindings", sxbuiltins.EnvBindings},
	{"undefined?", sxbuiltins.UndefinedP}, {"defined?", sxbuiltins.DefinedP},
}
var builtinsFA = []struct {
	name string
	fn   sxeval.BuiltinFA
}{
	{"map", sxbuiltins.Map}, {"apply", sxbuiltins.Apply},
	{"fold", sxbuiltins.Fold}, {"fold-reverse", sxbuiltins.FoldReverse},
	{"current-environment", sxbuiltins.CurrentEnv},
	{"bound?", sxbuiltins.BoundP},
	{"environment-lookup", sxbuiltins.EnvLookup}, {"environment-resolve", sxbuiltins.EnvResolve},
	{"macroexpand-0", sxbuiltins.MacroExpand0},
	{"pp", sxbuiltins.Pretty},
}

func main() {
	rd := sxreader.MakeReader(os.Stdin)
	sf := rd.SymbolFactory()
	symQQ, symUQ, symUQS := sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")
	sxbuiltins.InstallQuasiQuoteReader(rd, symQQ, '`', symUQ, ',', symUQS, '@')

	mpe := mainParserExecutor{
		origParser:   nil,
		origExecutor: nil,
		logReader:    true,
		logParser:    true,
		logExpr:      false,
		logExecutor:  true,
	}
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
		engine.BindBuiltinA(bDef.name, bDef.fn)
	}
	for _, bDef := range builtinsFA {
		engine.BindBuiltinFA(bDef.name, bDef.fn)
	}
	engine.Bind("UNDEFINED", sx.MakeUndefined())
	engine.BindBuiltinA("log-reader", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logReader
		mpe.logReader = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-parser", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logParser
		mpe.logParser = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-expr", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logExpr
		mpe.logExpr = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-executor", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logExecutor
		mpe.logExecutor = !res
		return sx.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-off", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		mpe.logReader = false
		mpe.logParser = false
		mpe.logExecutor = false
		return sx.Nil(), nil
	})
	engine.BindBuiltinA("panic", func(args []sx.Object) (sx.Object, error) {
		err := sxbuiltins.CheckArgs(args, 0, 1)
		if err != nil {
			panic(err)
		}
		if len(args) == 0 {
			panic("common panic")
		}
		panic(args[0])
	})
	root.Freeze()
	env := sxeval.MakeChildEnvironment(engine.GetToplevelEnv(), "repl", 1024)
	env.Bind(sf.MustMake("root-env"), root)
	env.Bind(sf.MustMake("repl-env"), env)
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
		expr = eng.Rework(expr)
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
	case *sxeval.BuiltinCallExpr:
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
	case *sxbuiltins.LetExpr:
		fmt.Println("LET")
		printLet(eng, e, level+1)
	case *sxbuiltins.AndExpr:
		fmt.Println("AND")
		printExprSeq(eng, &e.ExprSeq, level+1)
	case *sxbuiltins.OrExpr:
		fmt.Println("OR")
		printExprSeq(eng, &e.ExprSeq, level+1)
	case *sxbuiltins.LambdaExpr:
		fmt.Printf("LAMBDA %q", e.Name)
		for _, sym := range e.Params {
			fmt.Printf(" %v", sym)
		}
		if e.Rest != nil {
			fmt.Printf(" . %v", e.Rest)
		}
		fmt.Println()
		printExprSeq(eng, &e.ExprSeq, level+1)
	case *sxbuiltins.BeginExpr:
		fmt.Println("BEGIN")
		printExprSeq(eng, &e.ExprSeq, level+1)
	case *sxbuiltins.If2Expr:
		fmt.Println("IF2")
		printExpr(eng, e.Test, level+1)
		printExpr(eng, e.True, level+1)
	case *sxbuiltins.If3Expr:
		fmt.Println("IF3")
		printExpr(eng, e.Test, level+1)
		printExpr(eng, e.True, level+1)
		printExpr(eng, e.False, level+1)
	case *sxbuiltins.DefineExpr:
		fmt.Println("DEFINE", e.Sym)
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
			fmt.Printf("%T\n", expr)
		}
	}
}
func printLet(eng *sxeval.Engine, le *sxbuiltins.LetExpr, level int) {
	for i, sym := range le.Symbols {
		fmt.Print(strings.Repeat(" ", level*2))
		fmt.Print(sym, ":")
		printExpr(eng, le.Exprs[i], -level)
	}
	printExprSeq(eng, &le.ExprSeq, level)

}
func printExprSeq(eng *sxeval.Engine, exs *sxbuiltins.ExprSeq, level int) {
	for _, ex := range exs.Front {
		printExpr(eng, ex, level)
	}
	printExpr(eng, exs.Last, level)
}
