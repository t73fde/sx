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
	"zettelstore.de/sx.fossil/sxbuiltins/binding"
	"zettelstore.de/sx.fossil/sxbuiltins/callable"
	"zettelstore.de/sx.fossil/sxbuiltins/cond"
	"zettelstore.de/sx.fossil/sxbuiltins/define"
	"zettelstore.de/sx.fossil/sxbuiltins/env"
	"zettelstore.de/sx.fossil/sxbuiltins/equiv"
	"zettelstore.de/sx.fossil/sxbuiltins/list"
	"zettelstore.de/sx.fossil/sxbuiltins/macro"
	"zettelstore.de/sx.fossil/sxbuiltins/number"
	"zettelstore.de/sx.fossil/sxbuiltins/pprint"
	"zettelstore.de/sx.fossil/sxbuiltins/quote"
	"zettelstore.de/sx.fossil/sxbuiltins/timeit"
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
	{"define", define.DefineS}, {"set!", define.SetXS},
	{"if", cond.IfS},
	{"begin", cond.BeginS},
	{"and", sxbuiltins.AndS}, {"or", sxbuiltins.OrS},
	{"lambda", callable.LambdaS},
	{"let", binding.LetS},
	{"timeit", timeit.TimeitS},
	{"defmacro", macro.DefMacroS}, {"macro", macro.MacroS},
}

var builtinsA = []struct {
	name string
	fn   sxeval.BuiltinA
}{
	{"eq?", equiv.EqP}, {"eql?", equiv.EqlP}, {"equal?", equiv.EqualP},
	{"boolean?", sxbuiltins.BooleanP}, {"boolean", sxbuiltins.Boolean}, {"not", sxbuiltins.Not},
	{"number?", number.NumberP},
	{"+", number.Add}, {"-", number.Sub}, {"*", number.Mul},
	{"div", number.Div}, {"mod", number.Mod},
	{"=", number.Equal},
	{"<", number.Less}, {"<=", number.LessEqual},
	{">=", number.GreaterEqual}, {">", number.Greater},
	{"min", number.Min}, {"max", number.Max},
	{"cons", list.Cons}, {"pair?", list.PairP},
	{"null?", list.NullP}, {"list?", list.ListP},
	{"car", list.Car}, {"cdr", list.Cdr}, {"last", list.Last},
	{"list", list.List}, {"list*", list.ListStar}, {"append", list.Append}, {"reverse", list.Reverse},
	{"length", list.Length},
	{"callable?", callable.CallableP},
	{"parent-env", env.ParentEnv}, {"bindings", env.Bindings}, {"all-bindings", env.AllBindings},
}
var builtinsFA = []struct {
	name string
	fn   sxeval.BuiltinFA
}{
	{"map", callable.Map}, {"apply", callable.Apply},
	{"fold", callable.Fold}, {"fold-reverse", callable.FoldReverse},
	{"env", env.Env},
	{"bound?", env.BoundP},
	{"macroexpand-0", macro.MacroExpand0},
	{"pp", pprint.Pretty},
}

func main() {
	rd := sxreader.MakeReader(os.Stdin)
	sf := rd.SymbolFactory()
	symQuote := sf.MustMake("quote")
	quote.InstallQuoteReader(rd, symQuote, '\'')
	symQQ, symUQ, symUQS := sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")
	quote.InstallQuasiQuoteReader(rd, symQQ, '`', symUQ, ',', symUQS, '@')

	mpe := mainParserExecutor{
		origParser:   nil,
		origExecutor: nil,
		logReader:    true,
		logParser:    true,
		logExpr:      false,
		logExecutor:  true,
	}
	engine := sxeval.MakeEngine(sf, sxeval.MakeRootEnvironment())
	mpe.origParser = engine.SetParser(&mpe)
	mpe.origExecutor = engine.SetExecutor(&mpe)
	root := engine.RootEnvironment()
	root, _ = quote.InstallQuoteSyntax(root, symQuote)
	root, _ = quote.InstallQuasiQuoteSyntax(root, symQQ, symUQ, symUQS)
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
	case *binding.LetExpr:
		fmt.Println("LET")
		for i, sym := range e.Symbols {
			fmt.Print(strings.Repeat(" ", (level+1)*2))
			fmt.Print(sym, ":")
			printExpr(eng, e.Expr[i], -(level + 1))
		}
		for _, ex := range e.Front {
			printExpr(eng, ex, level+1)
		}
		printExpr(eng, e.Last, level+1)
	case *sxbuiltins.AndExpr:
		fmt.Println("AND")
		for _, ex := range e.Front {
			printExpr(eng, ex, level+1)
		}
		printExpr(eng, e.Last, level+1)
	case *sxbuiltins.OrExpr:
		fmt.Println("OR")
		for _, ex := range e.Front {
			printExpr(eng, ex, level+1)
		}
		printExpr(eng, e.Last, level+1)
	case *callable.LambdaExpr:
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
	case *cond.BeginExpr:
		fmt.Println("BEGIN")
		for _, ex := range e.Front {
			printExpr(eng, ex, level+1)
		}
		printExpr(eng, e.Last, level+1)
	case *cond.If2Expr:
		fmt.Println("IF2")
		printExpr(eng, e.Test, level+1)
		printExpr(eng, e.True, level+1)
	case *cond.If3Expr:
		fmt.Println("IF3")
		printExpr(eng, e.Test, level+1)
		printExpr(eng, e.True, level+1)
		printExpr(eng, e.False, level+1)
	case *define.DefineExpr:
		fmt.Println("DEFINE", e.Sym)
		printExpr(eng, e.Val, level+1)
	case *define.SetXExpr:
		fmt.Println("SET!", e.Sym)
		printExpr(eng, e.Val, level+1)
	case quote.MakeListExpr:
		fmt.Println("MAKELIST")
		printExpr(eng, e.Elem, level+1)
	default:
		switch e {
		case sxeval.NilExpr:
			fmt.Println("NIL")
		case sxeval.TrueExpr:
			fmt.Println("TRUE")
		case sxeval.FalseExpr:
			fmt.Println("FALSE")
		default:
			fmt.Printf("%T\n", expr)
		}
	}
}
