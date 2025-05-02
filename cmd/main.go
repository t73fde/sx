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
	logInterpret bool
	parseLevel   int
	improveLevel int
	execLevel    int
	execCount    int
}

// ----- ComputeObserver methods

func (me *mainEngine) BeforeCompute(_ *sxeval.Environment, expr sxeval.Expr, bind *sxeval.Binding) (sxeval.Expr, error) {
	if me.logExecutor {
		spaces := strings.Repeat(" ", me.execLevel)
		me.execLevel++
		fmt.Printf("%s;X%d %v<-%v ", spaces, me.execLevel, bind, bind.Parent())
		_, _ = expr.Print(os.Stdout)
		fmt.Println()
	}
	return expr, nil
}

func (me *mainEngine) AfterCompute(_ *sxeval.Environment, _ sxeval.Expr, _ *sxeval.Binding, obj sx.Object, err error) {
	if me.logExecutor {
		spaces := strings.Repeat(" ", me.execLevel-1)
		me.execCount++
		if err == nil {
			fmt.Printf("%s;O%d %T %v\n", spaces, me.execLevel, obj, obj)
		} else {
			fmt.Printf("%s;o%d %v\n", spaces, me.execLevel, err)
		}
		me.execLevel--
	}
}

// ----- ParseObserver methods

func (me *mainEngine) BeforeParse(_ *sxeval.ParseEnvironment, form sx.Object, bind *sxeval.Binding) (sx.Object, error) {
	if me.logParse {
		spaces := strings.Repeat(" ", me.parseLevel)
		me.parseLevel++
		fmt.Printf("%s;P%v %v<-%v %T %v\n", spaces, me.parseLevel, bind, bind.Parent(), form, form)
	}
	return form, nil
}

func (me *mainEngine) AfterParse(_ *sxeval.ParseEnvironment, _ sx.Object, expr sxeval.Expr, bind *sxeval.Binding, err error) {
	if me.logParse {
		spaces := strings.Repeat(" ", me.parseLevel-1)
		fmt.Printf("%s;Q%v %v<-%v %v ", spaces, me.parseLevel, bind, bind.Parent(), err)
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
		fmt.Printf("%s;R%v %v<-%v ", spaces, me.improveLevel, bind, bind.Parent())
		_, _ = expr.Print(os.Stdout)
		fmt.Println()
	}
	return expr
}

func (me *mainEngine) AfterImprove(imp *sxeval.Improver, _, result sxeval.Expr, err error) {
	if me.logImprove {
		spaces := strings.Repeat(" ", me.improveLevel-1)
		bind := imp.Binding()
		fmt.Printf("%s;S%v %v<-%v %v ", spaces, me.improveLevel, bind, bind.Parent(), err)
		_, _ = result.Print(os.Stdout)
		fmt.Println()
		me.improveLevel--
	}
}

// ---- CompileObserver methods

func (me *mainEngine) LogCompile(sxc *sxeval.Compiler, s string, vals ...string) {
	if me.logCompile {
		level, pc, curPos, maxPos := sxc.Stats()
		spaces := strings.Repeat(" ", level)
		fmt.Printf("%s;C%d %d %d %d: %s", spaces, level+1, maxPos, curPos, pc, s)
		for _, val := range vals {
			fmt.Print(" ", val)
		}
		fmt.Println()
	}
}

// ----- InterpretObserver methods

func (me *mainEngine) LogInterpreter(_ *sxeval.ProgramExpr, level, ip int, s string, err error) {
	if me.logInterpret {
		spaces := strings.Repeat(" ", level)
		if err == nil {
			fmt.Printf("%v;I%d %d %s\n", spaces, level+1, ip, s)
		} else {
			fmt.Printf("%v;i%d %d %s: %v\n", spaces, level+1, ip, s, err)
		}
	}
}

var myBuiltins = []*sxeval.Builtin{
	{
		Name:     "panic",
		MinArity: 0,
		MaxArity: 1,
		TestPure: nil,
		Fn0:      func(*sxeval.Environment, *sxeval.Binding) error { panic("common panic") },
		Fn1:      func(env *sxeval.Environment, _ *sxeval.Binding) error { panic(env.Top()) },
	},
	{
		Name:     "stack",
		MinArity: 0,
		MaxArity: 0,
		TestPure: nil,
		Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
			env.Push(sx.Vector(slices.Collect(env.Stack())))
			return nil
		},
	},
}

func (me *mainEngine) bindOwn(root *sxeval.Binding) {
	_ = sxeval.BindBuiltins(root,
		&sxeval.Builtin{
			Name:     "log-reader",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logReader
				me.logReader = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-parse",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logParse
				me.logParse = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-improve",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logImprove
				me.logImprove = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-compile",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logCompile
				me.logCompile = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-interpret",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logInterpret
				me.logInterpret = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-expr",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logExpr
				me.logExpr = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-executor",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				res := me.logExecutor
				me.logExecutor = !res
				env.Push(sx.MakeBoolean(res))
				return nil
			},
		},
		&sxeval.Builtin{
			Name:     "log-off",
			MinArity: 0,
			MaxArity: 0,
			TestPure: nil,
			Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
				me.logReader = false
				me.logParse = false
				me.logImprove = false
				me.logExecutor = false
				env.Push(sx.Nil())
				return nil
			},
		},
	)
}

func main() {
	rd := sxreader.MakeReader(os.Stdin)

	root := sxeval.MakeRootBinding(256)
	_ = sxbuiltins.BindAll(root)
	_ = sxeval.BindBuiltins(root, myBuiltins...)
	_ = root.Bind(sx.MakeSymbol("UNDEFINED"), sx.MakeUndefined())
	_ = root.Bind(sx.MakeSymbol("NIL"), sx.Nil())
	_ = root.Bind(sx.MakeSymbol("T"), sx.MakeSymbol("T"))
	me := mainEngine{
		logReader:    true,
		logParse:     true,
		logImprove:   true,
		logCompile:   true,
		logInterpret: true,
		logExpr:      false,
		logExecutor:  true,
	}
	me.bindOwn(root)
	err := sxbuiltins.LoadPrelude(root)
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
		env := sxeval.MakeEnvironment()
		env.SetComputeObserver(me).
			SetParseObserver(me).
			SetImproveObserver(me).
			SetCompileObserver(me).
			SetInterpretObserver(me)
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
		expr, err := env.Parse(obj, bind)
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
		res, err := env.Run(expr, bind)
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
	case *sxeval.CallExpr:
		fmt.Println("CALL")
		printExpr(e.Proc, level+1)
		for _, arg := range e.Args {
			printExpr(arg, level+1)
		}
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
