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

// Package main provides a simple interpreter for s-expressions.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"strings"
	"sync"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/sx/sxeval"
	"t73f.de/r/sx/sxreader"
)

type mainEngine struct {
	logReader    bool
	logParse     bool
	logImprove   bool
	logCompile   bool
	logExpr      bool
	logExecutor  bool
	parseLevel   int
	improveLevel int
	execLevel    int
	execCount    int
}

// ----- ExecuteObserver methods

func (me *mainEngine) BeforeExecution(env *sxeval.Environment, expr sxeval.Expr) (sxeval.Expr, error) {
	if me.logExecutor {
		spaces := strings.Repeat(" ", me.execLevel)
		me.execLevel++
		bind := env.Binding()
		fmt.Printf("%v;X%d %v<-%v ", spaces, me.execLevel, bind, bind.Parent())
		_, _ = expr.Print(os.Stdout)
		fmt.Println()
	}
	return expr, nil
}

func (me *mainEngine) AfterExecution(_ *sxeval.Environment, _ sxeval.Expr, obj sx.Object, err error) {
	if me.logExecutor {
		spaces := strings.Repeat(" ", me.execLevel-1)
		me.execCount++
		if err == nil {
			fmt.Printf("%v;O%d %T %v\n", spaces, me.execLevel, obj, obj)
		} else {
			fmt.Printf("%v;o%d %v\n", spaces, me.execLevel, err)
		}
		me.execLevel--
	}
}

// ----- ParseObserver methods

func (me *mainEngine) BeforeParse(pe *sxeval.ParseEnvironment, form sx.Object) (sx.Object, error) {
	if me.logParse {
		spaces := strings.Repeat(" ", me.parseLevel)
		me.parseLevel++
		bind := pe.Binding()
		fmt.Printf("%v;P%v %v<-%v %T %v\n", spaces, me.parseLevel, bind, bind.Parent(), form, form)
	}
	return form, nil
}

func (me *mainEngine) AfterParse(pe *sxeval.ParseEnvironment, _ sx.Object, expr sxeval.Expr, err error) {
	if me.logParse {
		spaces := strings.Repeat(" ", me.parseLevel-1)
		bind := pe.Binding()
		fmt.Printf("%v;Q%v %v<-%v %v ", spaces, me.parseLevel, bind, bind.Parent(), err)
		if err == nil {
			_, _ = expr.Print(os.Stdout)
		}
		fmt.Println()
		me.parseLevel--
	}
}

//----- ImproveObserver methods

func (me *mainEngine) BeforeImprove(imp *sxeval.Improver, expr sxeval.Expr) sxeval.Expr {
	if me.logImprove {
		spaces := strings.Repeat(" ", me.improveLevel)
		me.improveLevel++
		bind := imp.Binding()
		fmt.Printf("%v;R%v %v<-%v ", spaces, me.improveLevel, bind, bind.Parent())
		_, _ = expr.Print(os.Stdout)
		fmt.Println()
	}
	return expr
}

func (me *mainEngine) AfterImprove(imp *sxeval.Improver, _, result sxeval.Expr, err error) {
	if me.logImprove {
		spaces := strings.Repeat(" ", me.improveLevel-1)
		bind := imp.Binding()
		fmt.Printf("%v;S%v %v<-%v %v ", spaces, me.improveLevel, bind, bind.Parent(), err)
		_, _ = result.Print(os.Stdout)
		fmt.Println()
		me.improveLevel--
	}
}

// ---- CompileObserver methods

func (me *mainEngine) LogCompile(sxc *sxeval.Compiler, s string, vals ...string) {
	if me.logCompile {
		level, pc, curPos, maxPos := sxc.Stats()
		fmt.Printf(";C%d %d %d %d: %s", level, maxPos, curPos, pc, s)
		for _, val := range vals {
			fmt.Print(" ", val)
		}
		fmt.Println()
	}
}

var specials = []*sxeval.Special{
	&sxbuiltins.QuoteS, &sxbuiltins.QuasiquoteS, // quote, quasiquote
	&sxbuiltins.UnquoteS, &sxbuiltins.UnquoteSplicingS, // unquote, unquote-splicing
	&sxbuiltins.DefVarS,                     // defvar
	&sxbuiltins.DefunS, &sxbuiltins.LambdaS, // defun, lambda
	&sxbuiltins.DefDynS, &sxbuiltins.DynLambdaS, // defdyn, dyn-lambda
	&sxbuiltins.DefMacroS, //  defmacro
	&sxbuiltins.LetS,      // let
	&sxbuiltins.SetXS,     // set!
	&sxbuiltins.IfS,       // if
	&sxbuiltins.BeginS,    // begin
}

var builtins = []*sxeval.Builtin{
	&sxbuiltins.Equal,                    // =
	&sxbuiltins.Identical,                // ==
	&sxbuiltins.SymbolP,                  // symbol?
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
	&sxbuiltins.VectorSetBang,                   // vset!
	&sxbuiltins.List2Vector,                     // list->vector
	&sxbuiltins.Length, &sxbuiltins.LengthEqual, // length, length=
	&sxbuiltins.LengthLess, &sxbuiltins.LengthGreater, // length<, length>
	&sxbuiltins.Nth,               // nth
	&sxbuiltins.Sequence2List,     // ->list
	&sxbuiltins.CallableP,         // callable?
	&sxbuiltins.Macroexpand0,      // macroexpand-0
	&sxbuiltins.DefinedP,          // defined?
	&sxbuiltins.CurrentBinding,    // current-binding
	&sxbuiltins.ParentBinding,     // parent-binding
	&sxbuiltins.Bindings,          // environment-bindings
	&sxbuiltins.BoundP,            // bound?
	&sxbuiltins.BindingLookup,     // binding-lookup
	&sxbuiltins.BindingResolve,    // binding-resolve
	&sxbuiltins.Pretty,            // pp
	&sxbuiltins.Error,             // error
	&sxbuiltins.NotBoundError,     // not-bound-error
	&sxbuiltins.ParseExpression,   // parse-expression
	&sxbuiltins.UnparseExpression, // unparse-expression
	&sxbuiltins.RunExpression,     // run-expression
	&sxbuiltins.Eval,              // eval
	{
		Name:     "panic",
		MinArity: 0,
		MaxArity: 1,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			panic("common panic")
		},
		Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
			panic(arg)
		},
	},
	{
		Name:     "stack",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(env *sxeval.Environment) (sx.Object, error) {
			return sx.Vector(slices.Clone(env.Stack())), nil
		},
	},
}

func (me *mainEngine) bindOwn(root *sxeval.Binding) {
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-reader",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			res := me.logReader
			me.logReader = !res
			return sx.MakeBoolean(res), nil
		},
	})
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-parse",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			res := me.logParse
			me.logParse = !res
			return sx.MakeBoolean(res), nil
		},
	})
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-improve",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			res := me.logImprove
			me.logImprove = !res
			return sx.MakeBoolean(res), nil
		},
	})
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-compile",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			res := me.logCompile
			me.logCompile = !res
			return sx.MakeBoolean(res), nil
		},
	})
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-expr",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			res := me.logExpr
			me.logExpr = !res
			return sx.MakeBoolean(res), nil
		},
	})
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-executor",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			res := me.logExecutor
			me.logExecutor = !res
			return sx.MakeBoolean(res), nil
		},
	})
	_ = root.BindBuiltin(&sxeval.Builtin{
		Name:     "log-off",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(*sxeval.Environment) (sx.Object, error) {
			me.logReader = false
			me.logParse = false
			me.logImprove = false
			me.logExecutor = false
			return sx.Nil(), nil
		},
	})

}

func main() {
	rd := sxreader.MakeReader(os.Stdin)

	root := sxeval.MakeRootBinding(len(specials) + len(builtins) + 16)
	for _, synDef := range specials {
		_ = root.BindSpecial(synDef)
	}
	for _, b := range builtins {
		_ = root.BindBuiltin(b)
	}
	_ = root.Bind(sx.MakeSymbol("UNDEFINED"), sx.MakeUndefined())
	_ = root.Bind(sx.MakeSymbol("NIL"), sx.Nil())
	_ = root.Bind(sx.MakeSymbol("T"), sx.MakeSymbol("T"))
	me := mainEngine{
		logReader:   true,
		logParse:    true,
		logImprove:  true,
		logCompile:  true,
		logExpr:     false,
		logExecutor: true,
	}
	me.bindOwn(root)
	err := readPrelude(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read prelude: %v\n", err)
		os.Exit(17)
	}
	root.Freeze()
	bind := root.MakeChildBinding("repl", 1024)
	_ = bind.Bind(sx.MakeSymbol("root-binding"), root)
	_ = bind.Bind(sx.MakeSymbol("repl-binding"), bind)

	// Good to disable const checking
	_ = bind.Bind(sx.MakeSymbol("a"), sx.Int64(3))
	_ = bind.Bind(sx.MakeSymbol("b"), sx.Int64(4))
	_ = bind.Bind(sx.MakeSymbol("c"), sx.Int64(11))

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
		env := sxeval.MakeExecutionEnvironment(bind)
		env.SetExecutor(me).SetParseObserver(me).SetImproveObserver(me).SetCompileObserver(me)
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
		cexpr, err := env.Compile(expr)
		if err == nil {
			expr = cexpr
		} else {
			fmt.Println(";c", err)
		}
		if me.logReader {
			fmt.Printf(";= ")
			_, _ = expr.Print(os.Stdout)
			fmt.Println()
		}
		if me.logExpr {
			printExpr(expr, 0)
			continue
		}
		me.execCount = 0
		res, err := env.Run(expr)
		if me.logExecutor {
			fmt.Println(";#", me.execCount)
		}
		if err != nil {
			fmt.Println(";e", err)
			var execErr sxeval.ExecuteError
			if errors.As(err, &execErr) {
				execErr.PrintStack(os.Stdout, ";", nil, "")
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
	case sxeval.UnboundSymbolExpr:
		fmt.Printf("UNBOUND %v\n", e.GetSymbol())
	case sxeval.ResolveSymbolExpr:
		fmt.Printf("RESOLVE %v\n", e.GetSymbol())
	case *sxeval.LookupSymbolExpr:
		fmt.Printf("LOOKUP/%d %v\n", e.GetLevel(), e.GetSymbol())
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
	case *sxbuiltins.LetExpr:
		fmt.Println("LET")
		level++
		for i, sym := range e.Symbols {
			fmt.Print(strings.Repeat(" ", level*2))
			fmt.Print(sym, ":")
			printExpr(e.Vals[i], -level)
		}
		printExpr(e.Body, level)
	case *sxbuiltins.IfExpr:
		fmt.Println("IF")
		printExpr(e.Test, level+1)
		printExpr(e.True, level+1)
		printExpr(e.False, level+1)
	case *sxbuiltins.DefineExpr:
		fmt.Println("DEFVAR", e.Sym)
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
