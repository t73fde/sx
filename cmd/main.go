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
	baseExecutor sxeval.Executor
	origParser   sxeval.Parser
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

func (mpe *mainParserExecutor) Reset() { mpe.baseExecutor.Reset() }

func (mpe *mainParserExecutor) Execute(frame *sxeval.Frame, expr sxeval.Expr) (sx.Object, error) {
	if !mpe.logExecutor {
		return mpe.baseExecutor.Execute(frame, expr)
	}
	env := frame.Environment()
	fmt.Printf(";X %v<-%v ", env, env.Parent())
	expr.Print(os.Stdout)
	fmt.Println()
	obj, err := mpe.baseExecutor.Execute(frame, expr)
	if err != nil {
		return nil, err
	}
	fmt.Printf(";O %T %v\n", obj, obj)
	return obj, nil
}

var specials = []*sxeval.Special{
	&sxbuiltins.QuoteS, &sxbuiltins.QuasiquoteS, // quote, quasiquote
	&sxbuiltins.UnquoteS, &sxbuiltins.UnquoteSplicingS, // unquote, unquote-splicing
	&sxbuiltins.DefVarS, &sxbuiltins.DefConstS, // defvar, defconst
	&sxbuiltins.DefunS, &sxbuiltins.LambdaS, // defun, lambda
	&sxbuiltins.SetXS,     // set!
	&sxbuiltins.CondS,     // cond
	&sxbuiltins.IfS,       // if
	&sxbuiltins.BeginS,    // begin
	&sxbuiltins.DefMacroS, // defmacro
}

var builtins = []*sxeval.Builtin{
	&sxbuiltins.Equal,                    // =
	&sxbuiltins.Identical,                // ==
	&sxbuiltins.NullP,                    // null?
	&sxbuiltins.Cons,                     // cons
	&sxbuiltins.PairP, &sxbuiltins.ListP, // pair?, list?
	&sxbuiltins.Car, &sxbuiltins.Cdr, // car, cdr
	&sxbuiltins.Caar, &sxbuiltins.Cadr, &sxbuiltins.Cdar, &sxbuiltins.Cddr,
	&sxbuiltins.Caaar, &sxbuiltins.Caadr, &sxbuiltins.Cadar, &sxbuiltins.Caddr,
	&sxbuiltins.Cdaar, &sxbuiltins.Cdadr, &sxbuiltins.Cddar, &sxbuiltins.Cdddr,
	&sxbuiltins.Caaaar, &sxbuiltins.Caaadr, &sxbuiltins.Caadar, &sxbuiltins.Caaddr,
	&sxbuiltins.Cadaar, &sxbuiltins.Cadadr, &sxbuiltins.Caddar, &sxbuiltins.Cadddr,
	&sxbuiltins.Cdaaar, &sxbuiltins.Cdaadr, &sxbuiltins.Cdadar, &sxbuiltins.Cdaddr,
	&sxbuiltins.Cddaar, &sxbuiltins.Cddadr, &sxbuiltins.Cdddar, &sxbuiltins.Cddddr,
	&sxbuiltins.Last,                       // last
	&sxbuiltins.List, &sxbuiltins.ListStar, // list, list*
	&sxbuiltins.Append,                        // append
	&sxbuiltins.Reverse,                       // reverse
	&sxbuiltins.Length,                        // length
	&sxbuiltins.Assoc,                         // assoc
	&sxbuiltins.Map,                           // map
	&sxbuiltins.Apply,                         // apply
	&sxbuiltins.Fold, &sxbuiltins.FoldReverse, // fold, fold-reverse
	&sxbuiltins.NumberP,                               // number?
	&sxbuiltins.Add, &sxbuiltins.Sub, &sxbuiltins.Mul, // +, -, *
	&sxbuiltins.Div, &sxbuiltins.Mod, // div, mod
	&sxbuiltins.NumLess, &sxbuiltins.NumLessEqual, // <, <=
	&sxbuiltins.NumGreater, &sxbuiltins.NumGreaterEqual, // >, >=
	&sxbuiltins.ToString, &sxbuiltins.StringAppend, // ->string, string-append
	&sxbuiltins.CallableP,    //callable?
	&sxbuiltins.Macroexpand0, // macroexpand-0
	&sxbuiltins.Defined,      // defined?
	&sxbuiltins.CurrentEnv,   // current-environment
	&sxbuiltins.ParentEnv,    // parent-environment
	&sxbuiltins.EnvBindings,  //environment-bindings
	&sxbuiltins.BoundP,       // bound?
	&sxbuiltins.EnvLookup,    // environment-lookup
	&sxbuiltins.EnvResolve,   // environment-resolve
	&sxbuiltins.Pretty,       // pp
}

func main() {
	sf := sx.MakeMappedFactory(1024)
	rd := sxreader.MakeReader(os.Stdin, sxreader.WithSymbolFactory(sf))

	mpe := mainParserExecutor{baseExecutor: &sxeval.SimpleExecutor{}}
	engine := sxeval.MakeEngine(sf, sxeval.MakeRootEnvironment(len(specials)+len(builtins)+16))
	mpe.origParser = engine.SetParser(&mpe)
	root := engine.RootEnvironment()
	for _, synDef := range specials {
		engine.BindSpecial(synDef)
	}
	for _, b := range builtins {
		engine.BindBuiltin(b)
	}
	engine.Bind("UNDEFINED", sx.MakeUndefined())
	engine.BindBuiltin(&sxeval.Builtin{
		Name:     "log-reader",
		MinArity: 0,
		MaxArity: 0,
		IsPure:   false,
		Fn: func(*sxeval.Frame, []sx.Object) (sx.Object, error) {
			res := mpe.logReader
			mpe.logReader = !res
			return sx.MakeBoolean(res), nil
		},
	})
	engine.BindBuiltin(&sxeval.Builtin{
		Name:     "log-parser",
		MinArity: 0,
		MaxArity: 0,
		IsPure:   false,
		Fn: func(*sxeval.Frame, []sx.Object) (sx.Object, error) {
			res := mpe.logParser
			mpe.logParser = !res
			return sx.MakeBoolean(res), nil
		},
	})
	engine.BindBuiltin(&sxeval.Builtin{
		Name:     "log-expr",
		MinArity: 0,
		MaxArity: 0,
		IsPure:   false,
		Fn: func(*sxeval.Frame, []sx.Object) (sx.Object, error) {
			res := mpe.logExpr
			mpe.logExpr = !res
			return sx.MakeBoolean(res), nil
		},
	})
	engine.BindBuiltin(&sxeval.Builtin{
		Name:     "log-executor",
		MinArity: 0,
		MaxArity: 0,
		IsPure:   false,
		Fn: func(*sxeval.Frame, []sx.Object) (sx.Object, error) {
			res := mpe.logExecutor
			mpe.logExecutor = !res
			return sx.MakeBoolean(res), nil
		},
	})
	engine.BindBuiltin(&sxeval.Builtin{
		Name:     "log-off",
		MinArity: 0,
		MaxArity: 0,
		IsPure:   false,
		Fn: func(*sxeval.Frame, []sx.Object) (sx.Object, error) {
			mpe.logReader = false
			mpe.logParser = false
			mpe.logExecutor = false
			return sx.Nil(), nil
		},
	})
	engine.BindBuiltin(&sxeval.Builtin{
		Name:     "panic",
		MinArity: 0,
		MaxArity: 1,
		IsPure:   false,
		Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
			if len(args) == 0 {
				panic("common panic")
			}
			panic(args[0])
		},
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
		expr, err := eng.Parse(obj, env)
		if err != nil {
			fmt.Println(";p", err)
			continue
		}
		expr = eng.Rework(expr, env, mpe)
		if mpe.logExpr {
			printExpr(expr, 0)
			continue
		} else if mpe.logReader {
			fmt.Printf(";= ")
			expr.Print(os.Stdout)
			fmt.Println()
		}
		res, err := eng.Execute(expr, env, mpe)
		if err != nil {
			fmt.Println(";e", err)
			continue
		}
		fmt.Println(sx.Repr(res))
	}
}

func printExpr(expr sxeval.Expr, level int) {
	if level <= 0 {
		level = -level
	} else {
		fmt.Print(strings.Repeat(" ", level*2))
	}

	switch e := expr.(type) {
	case *sxeval.BuiltinCallExpr:
		fmt.Printf("B-CALL %v\n", e.Proc.Name)
		for _, arg := range e.Args {
			printExpr(arg, level+1)
		}
	case *sxeval.CallExpr:
		fmt.Println("CALL")
		printExpr(e.Proc, level+1)
		for _, arg := range e.Args {
			printExpr(arg, level+1)
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
		printExpr(e.Expr, level+1)
	case *sxbuiltins.CondExpr:
		fmt.Println("COND")
		for _, clause := range e.Cases {
			printExpr(clause.Test, level+1)
			printExpr(clause.Expr, level+1)
		}
	case *sxbuiltins.IfExpr:
		fmt.Println("IF")
		printExpr(e.Test, level+1)
		printExpr(e.True, level+1)
		printExpr(e.False, level+1)
	case *sxbuiltins.DefineExpr:
		if e.Const {
			fmt.Println("DEFCONST", e.Sym)
		} else {
			fmt.Println("DEFVAR", e.Sym)
		}
		printExpr(e.Val, level+1)
	case *sxbuiltins.SetXExpr:
		fmt.Println("SET!", e.Sym)
		printExpr(e.Val, level+1)
	case sxbuiltins.MakeListExpr:
		fmt.Println("MAKELIST")
		printExpr(e.Elem, level+1)
	default:
		switch e {
		case sxeval.NilExpr:
			fmt.Println("NIL")
		default:
			fmt.Printf("%T/%v\n", expr, expr)
		}
	}
}

//go:embed prelude.sxn
var prelude string

func readPrelude(engine *sxeval.Engine) error {
	rd := sxreader.MakeReader(strings.NewReader(prelude), sxreader.WithSymbolFactory(engine.SymbolFactory()))
	for {
		form, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err = engine.Eval(form, engine.RootEnvironment(), nil)
		if err != nil {
			return err
		}
	}
}
