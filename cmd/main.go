//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package main

import (
	_ "embed"
	"errors"
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

type mainEngine struct {
	logReader   bool
	logParse    bool
	logExpr     bool
	logExecutor bool
	parseLevel  int
	execLevel   int
	execCount   int
}

// ----- Executor methods

// Reset the executor.
func (me *mainEngine) Reset() { me.execCount = 0 }

func (me *mainEngine) Execute(env *sxeval.Environment, expr sxeval.Expr) (sx.Object, error) {
	if !me.logExecutor {
		return expr.Compute(env)
	}
	spaces := strings.Repeat(" ", me.execLevel)
	me.execLevel++
	bind := env.Binding()
	fmt.Printf("%v;X%d %v<-%v ", spaces, me.execLevel, bind, bind.Parent())
	expr.Print(os.Stdout)
	fmt.Println()
	obj, err := expr.Compute(env)
	me.execCount++
	if err == nil {
		fmt.Printf("%v;O%d %T %v\n", spaces, me.execLevel, obj, obj)
	}
	me.execLevel--
	return obj, err
}

// ----- ParseObserver methods

// BeforeParse logs everythinf before the parsing happens.
func (me *mainEngine) BeforeParse(pe *sxeval.ParseEnvironment, form sx.Object) sx.Object {
	if me.logParse {
		spaces := strings.Repeat(" ", me.parseLevel)
		me.parseLevel++
		bind := pe.Binding()
		fmt.Printf("%v;P%v %v<-%v %T %v\n", spaces, me.parseLevel, bind, bind.Parent(), form, form)
	}
	return form
}

func (me *mainEngine) AfterParse(pe *sxeval.ParseEnvironment, form sx.Object, expr sxeval.Expr, err error) {
	if me.logParse {
		spaces := strings.Repeat(" ", me.parseLevel-1)
		bind := pe.Binding()
		fmt.Printf("%v;Q%v %v<-%v %v %v\n", spaces, me.parseLevel, bind, bind.Parent(), err, expr)
		me.parseLevel--
	}
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
	&sxbuiltins.Append,               // append
	&sxbuiltins.Reverse,              // reverse
	&sxbuiltins.Length,               // length
	&sxbuiltins.Assoc,                // assoc
	&sxbuiltins.All, &sxbuiltins.Any, // all, any
	&sxbuiltins.Map,                           // map
	&sxbuiltins.Apply,                         // apply
	&sxbuiltins.Fold, &sxbuiltins.FoldReverse, // fold, fold-reverse
	&sxbuiltins.NumberP,                               // number?
	&sxbuiltins.Add, &sxbuiltins.Sub, &sxbuiltins.Mul, // +, -, *
	&sxbuiltins.Div, &sxbuiltins.Mod, // div, mod
	&sxbuiltins.NumLess, &sxbuiltins.NumLessEqual, // <, <=
	&sxbuiltins.NumGreater, &sxbuiltins.NumGreaterEqual, // >, >=
	&sxbuiltins.ToString, &sxbuiltins.Concat, // ->string, concat
	&sxbuiltins.Vector, &sxbuiltins.VectorP, // vector, vector?
	&sxbuiltins.VectorLength,                         // vector-length
	&sxbuiltins.VectorGet, &sxbuiltins.VectorSetBang, // vget, vset
	&sxbuiltins.Vector2List, &sxbuiltins.List2Vector, // vector->list, list->vector
	&sxbuiltins.CallableP,      // callable?
	&sxbuiltins.Macroexpand0,   // macroexpand-0
	&sxbuiltins.Defined,        // defined?
	&sxbuiltins.CurrentBinding, // current-environment
	&sxbuiltins.ParentBinding,  // parent-environment
	&sxbuiltins.Bindings,       // environment-bindings
	&sxbuiltins.BoundP,         // bound?
	&sxbuiltins.BindingLookup,  // environment-lookup
	&sxbuiltins.BindingResolve, // environment-resolve
	&sxbuiltins.Pretty,         // pp
	&sxbuiltins.Error,          // error
	&sxbuiltins.NotBoundError,  // not-bound-error
	{
		Name:     "panic",
		MinArity: 0,
		MaxArity: 1,
		TestPure: nil,
		Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
			if len(args) == 0 {
				panic("common panic")
			}
			panic(args[0])
		},
	},
}

func (me *mainEngine) bindOwn(root *sxeval.Binding) {
	root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-reader",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn: func(*sxeval.Environment, sx.Vector) (sx.Object, error) {
			res := me.logReader
			me.logReader = !res
			return sx.MakeBoolean(res), nil
		},
	})
	root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-parse",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn: func(*sxeval.Environment, sx.Vector) (sx.Object, error) {
			res := me.logParse
			me.logParse = !res
			return sx.MakeBoolean(res), nil
		},
	})
	root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-expr",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn: func(*sxeval.Environment, sx.Vector) (sx.Object, error) {
			res := me.logExpr
			me.logExpr = !res
			return sx.MakeBoolean(res), nil
		},
	})
	root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-executor",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn: func(*sxeval.Environment, sx.Vector) (sx.Object, error) {
			res := me.logExecutor
			me.logExecutor = !res
			return sx.MakeBoolean(res), nil
		},
	})
	root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-off",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn: func(*sxeval.Environment, sx.Vector) (sx.Object, error) {
			me.logReader = false
			me.logParse = false
			me.logExecutor = false
			return sx.Nil(), nil
		},
	})

}

func main() {
	rd := sxreader.MakeReader(os.Stdin)

	root := sxeval.MakeRootBinding(len(specials) + len(builtins) + 16)
	for _, synDef := range specials {
		root.BindSpecial(synDef)
	}
	for _, b := range builtins {
		root.BindBuiltin(b)
	}
	root.Bind("UNDEFINED", sx.MakeUndefined())
	me := mainEngine{}
	me.bindOwn(root)
	err := readPrelude(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read prelude: %v\n", err)
		os.Exit(17)
	}
	root.Freeze()
	bind := sxeval.MakeChildBinding(root, "repl", 1024)
	bind.Bind(sx.Symbol("root-binding"), root)
	bind.Bind(sx.Symbol("repl-binding"), bind)

	me.logReader = true
	me.logParse = true
	me.logExpr = false
	me.logExecutor = true

	var wg sync.WaitGroup
	wg.Add(1)
	go repl(rd, &me, bind, &wg)
	wg.Wait()
}

func repl(rd *sxreader.Reader, me *mainEngine, bind *sxeval.Binding, wg *sync.WaitGroup) {
	defer func() {
		if val := recover(); val != nil {
			stack := debug.Stack()
			fmt.Printf("RECOVER PANIC: %v\n\n%s\n", val, string(stack))
			go repl(rd, me, bind, wg)
			return
		}
		wg.Done()
	}()

	for {
		env := sxeval.MakeExecutionEnvironment(bind, sxeval.WithExecutor(me), sxeval.WithParseObserver(me))
		fmt.Print("> ")
		obj, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(";r", err)
			continue
		}
		if me.logReader {
			fmt.Println(";<", obj)
		}
		expr, err := env.Parse(obj)
		if err != nil {
			fmt.Println(";p", err)
			continue
		}
		expr = env.Rework(expr)
		if me.logExpr {
			printExpr(expr, 0)
			continue
		} else if me.logReader {
			fmt.Printf(";= ")
			expr.Print(os.Stdout)
			fmt.Println()
		}
		res, err := env.Run(expr)
		if me.logExecutor {
			fmt.Println(";#", me.execCount)
		}
		if err != nil {
			fmt.Println(";e", err)
			var execErr sxeval.ExecuteError
			if errors.As(err, &execErr) {
				for i, elem := range execErr.Stack {
					val := elem.Expr.Unparse()
					fmt.Printf(";n%d env: %v, expr: %T/%v\n", i, elem.Env, val, val)
				}
			}
			continue
		}
		fmt.Println(res)
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
	case sxeval.ResolveSymbolExpr:
		fmt.Printf("RESOLVE %v\n", e.Symbol)
	case sxeval.ObjExpr:
		fmt.Printf("OBJ %T/%v\n", e.Obj, e.Obj)
	case *sxbuiltins.LambdaExpr:
		fmt.Printf("LAMBDA %q", e.Name)
		for _, sym := range e.Params {
			fmt.Printf(" %v", sym)
		}
		if e.Rest != "" {
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

func readPrelude(root *sxeval.Binding) error {
	rd := sxreader.MakeReader(strings.NewReader(prelude))
	env := sxeval.MakeExecutionEnvironment(root)
	for {
		form, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err = env.Eval(form)
		if err != nil {
			return err
		}
	}
}
