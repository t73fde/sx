//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
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

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
	"zettelstore.de/sx.fossil/sxpf/builtins/binding"
	"zettelstore.de/sx.fossil/sxpf/builtins/boolean"
	"zettelstore.de/sx.fossil/sxpf/builtins/callable"
	"zettelstore.de/sx.fossil/sxpf/builtins/cond"
	"zettelstore.de/sx.fossil/sxpf/builtins/define"
	"zettelstore.de/sx.fossil/sxpf/builtins/env"
	"zettelstore.de/sx.fossil/sxpf/builtins/equiv"
	"zettelstore.de/sx.fossil/sxpf/builtins/list"
	"zettelstore.de/sx.fossil/sxpf/builtins/macro"
	"zettelstore.de/sx.fossil/sxpf/builtins/number"
	"zettelstore.de/sx.fossil/sxpf/builtins/pprint"
	"zettelstore.de/sx.fossil/sxpf/builtins/quote"
	"zettelstore.de/sx.fossil/sxpf/builtins/timeit"
	"zettelstore.de/sx.fossil/sxpf/eval"
	"zettelstore.de/sx.fossil/sxpf/reader"
)

type mainParserExecutor struct {
	origParser   eval.Parser
	origExecutor eval.Executor
	logReader    bool
	logParser    bool
	logExpr      bool
	logExecutor  bool
}

func (mpe *mainParserExecutor) Parse(eng *eval.Engine, env sxpf.Environment, form sxpf.Object) (eval.Expr, error) {
	if !mpe.logParser {
		return mpe.origParser.Parse(eng, env, form)
	}
	fmt.Printf(";P %v<-%v %T %v\n", env, env.Parent(), form, form)
	expr, err := mpe.origParser.Parse(eng, env, form)
	if err != nil {
		return nil, err
	}
	fmt.Printf(";Q ")
	expr.Print(os.Stdout)
	fmt.Println()
	return expr, nil
}

func (mpe *mainParserExecutor) Execute(eng *eval.Engine, env sxpf.Environment, expr eval.Expr) (sxpf.Object, error) {
	if !mpe.logExecutor {
		return mpe.origExecutor.Execute(eng, env, expr)
	}
	fmt.Printf(";X %v<-%v ", env, env.Parent())
	expr.Print(os.Stdout)
	fmt.Println()
	obj, err := mpe.origExecutor.Execute(eng, env, expr)
	if err != nil {
		return nil, err
	}
	fmt.Printf(";O %T %v\n", obj, obj)
	return obj, nil
}

var syntaxes = []struct {
	name string
	fn   eval.SyntaxFn
}{
	{"define", define.DefineS}, {"set!", define.SetXS},
	{"if", cond.IfS},
	{"begin", cond.BeginS},
	{"and", boolean.AndS}, {"or", boolean.OrS},
	{"lambda", callable.LambdaS},
	{"let", binding.LetS},
	{"timeit", timeit.TimeitS},
	{"defmacro", macro.DefMacroS}, {"macro", macro.MacroS},
}

var builtinsA = []struct {
	name string
	fn   eval.BuiltinA
}{
	{"eq?", equiv.EqP}, {"eql?", equiv.EqlP}, {"equal?", equiv.EqualP},
	{"boolean?", boolean.BooleanP}, {"boolean", boolean.Boolean}, {"not", boolean.Not},
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
var builtinsEEA = []struct {
	name string
	fn   eval.BuiltinEEA
}{
	{"map", callable.Map}, {"apply", callable.Apply},
	{"fold", callable.Fold}, {"fold-reverse", callable.FoldReverse},
	{"env", env.Env},
	{"bound?", env.BoundP},
	{"macroexpand-0", macro.MacroExpand0},
	{"pp", pprint.Pretty},
}

func main() {
	rd := reader.MakeReader(os.Stdin)
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
	engine := eval.MakeEngine(sf, sxpf.MakeRootEnvironment())
	mpe.origParser = engine.SetParser(&mpe)
	mpe.origExecutor = engine.SetExecutor(&mpe)
	root := engine.RootEnvironment()
	quote.InstallQuoteSyntax(root, symQuote)
	quote.InstallQuasiQuoteSyntax(root, symQQ, symUQ, symUQS)
	for _, synDef := range syntaxes {
		engine.BindSyntax(synDef.name, synDef.fn)
	}
	for _, bDef := range builtinsA {
		engine.BindBuiltinA(bDef.name, bDef.fn)
	}
	for _, bDef := range builtinsEEA {
		engine.BindBuiltinEEA(bDef.name, bDef.fn)
	}
	engine.Bind("UNDEFINED", sxpf.MakeUndefined())
	engine.BindBuiltinA("log-reader", func(args []sxpf.Object) (sxpf.Object, error) {
		err := builtins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logReader
		mpe.logReader = !res
		return sxpf.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-parser", func(args []sxpf.Object) (sxpf.Object, error) {
		err := builtins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logParser
		mpe.logParser = !res
		return sxpf.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-expr", func(args []sxpf.Object) (sxpf.Object, error) {
		err := builtins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logExpr
		mpe.logExpr = !res
		return sxpf.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-executor", func(args []sxpf.Object) (sxpf.Object, error) {
		err := builtins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		res := mpe.logExecutor
		mpe.logExecutor = !res
		return sxpf.MakeBoolean(res), nil
	})
	engine.BindBuiltinA("log-off", func(args []sxpf.Object) (sxpf.Object, error) {
		err := builtins.CheckArgs(args, 0, 0)
		if err != nil {
			return nil, err
		}
		mpe.logReader = false
		mpe.logParser = false
		mpe.logExecutor = false
		return sxpf.Nil(), nil
	})
	engine.BindBuiltinA("panic", func(args []sxpf.Object) (sxpf.Object, error) {
		err := builtins.CheckArgs(args, 0, 1)
		if err != nil {
			panic(err)
		}
		if len(args) == 0 {
			panic("common panic")
		}
		panic(args[0])
	})
	root.Freeze()
	env := sxpf.MakeChildEnvironment(engine.GetToplevelEnv(), "repl", 1024)
	env.Bind(sf.MustMake("root-env"), root)
	env.Bind(sf.MustMake("repl-env"), env)
	var wg sync.WaitGroup
	wg.Add(1)
	go repl(rd, &mpe, engine, env, &wg)
	wg.Wait()
}

func repl(rd *reader.Reader, mpe *mainParserExecutor, eng *eval.Engine, env sxpf.Environment, wg *sync.WaitGroup) {
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
		fmt.Println(sxpf.Repr(res))
	}
}

func printExpr(eng *eval.Engine, expr eval.Expr, level int) {
	if level <= 0 {
		level = -level
	} else {
		fmt.Print(strings.Repeat(" ", level*2))
	}

	switch e := expr.(type) {
	case *eval.BuiltinCallExpr:
		fmt.Printf("B-CALL %v\n", e.Proc.Name(eng))
		for _, arg := range e.Args {
			printExpr(eng, arg, level+1)
		}
	case *eval.CallExpr:
		fmt.Println("CALL")
		printExpr(eng, e.Proc, level+1)
		for _, arg := range e.Args {
			printExpr(eng, arg, level+1)
		}
	case eval.ResolveExpr:
		fmt.Printf("RESOLVE %v\n", e.Symbol)
	case eval.ObjExpr:
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
	case *boolean.AndExpr:
		fmt.Println("AND")
		for _, ex := range e.Front {
			printExpr(eng, ex, level+1)
		}
		printExpr(eng, e.Last, level+1)
	case *boolean.OrExpr:
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
		case eval.NilExpr:
			fmt.Println("NIL")
		case eval.TrueExpr:
			fmt.Println("TRUE")
		case eval.FalseExpr:
			fmt.Println("FALSE")
		default:
			fmt.Printf("%T\n", expr)
		}
	}
}
