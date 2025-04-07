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

package sxeval_test

import (
	"fmt"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

func BenchmarkEvenTCO(b *testing.B) {
	testcases := [...]int{0, 1, 2, 4, 16, 64, 512, 4096, 65536}
	root := createBindingForTCO()
	evenSym := sx.MakeSymbol("even?")
	for _, tc := range testcases {
		b.Run(fmt.Sprintf("%5d", tc), func(b *testing.B) {
			env := sxeval.MakeEnvironment()
			obj := sx.MakeList(evenSym, sx.Int64(tc))
			expr, err := parseAndMore(env, obj, root)
			if err != nil {
				b.Error(err)
			}
			b.ResetTimer()
			for b.Loop() {
				_, _ = env.Run(expr, root)
			}
		})
	}
}

func BenchmarkFac(b *testing.B) {
	runBenchmark(b, sx.MakeList(sx.MakeSymbol("fac"), sx.Int64(20)))
}

func BenchmarkFaa(b *testing.B) {
	runBenchmark(b, sx.MakeList(sx.MakeSymbol("faa"), sx.Int64(20), sx.Int64(1)))
}

func BenchmarkFib(b *testing.B) {
	runBenchmark(b, sx.MakeList(sx.MakeSymbol("fib"), sx.Int64(10)))
}

func BenchmarkTak(b *testing.B) {
	runBenchmark(b, sx.MakeList(sx.MakeSymbol("tak"), sx.Int64(15), sx.Int64(10), sx.Int64(5)))
}

func runBenchmark(b *testing.B, sexpr sx.Object) {
	root := createBindingForTCO()
	env := sxeval.MakeEnvironment()
	expr, err := parseAndMore(env, sexpr, root)
	if err != nil {
		b.Error(err)
	}

	if _, err = env.Run(expr, root); err != nil {
		b.Error(err)
	}
	if stack := env.Stack(); len(stack) > 0 {
		b.Error("stack not empty, but:", stack)
	}
	for b.Loop() {
		_, _ = env.Run(expr, root)
	}
	if stack := env.Stack(); len(stack) > 0 {
		b.Error("stack not empty, found", len(stack), "elements")
	}
}

func parseAndMore(env *sxeval.Environment, obj sx.Object, bind *sxeval.Binding) (sxeval.Expr, error) {
	expr, err := env.Parse(obj, bind)
	if err != nil {
		return nil, err
	}
	pe, err := env.Compile(expr)
	if err == nil {
		return pe, nil
	}
	return expr, nil
}

func BenchmarkAddExec(b *testing.B) {
	env, expr, root := prepareAddBenchmark()

	err := checkAddBenchmark(env, expr, root)
	if err != nil {
		b.Error(err)
		return
	}

	for b.Loop() {
		_, _ = env.Run(expr, root)
	}
}

func BenchmarkAddCompiled(b *testing.B) {
	env, expr, root := prepareAddBenchmark()
	cexpr, err := env.Compile(expr)
	if err != nil {
		b.Error(err)
		return
	}

	err = checkAddBenchmark(env, cexpr, root)
	if err != nil {
		b.Error(err)
		return
	}

	for b.Loop() {
		_, _ = env.Run(cexpr, root)
	}
	if stack := env.Stack(); len(stack) > 0 {
		b.Error("stack not empty", len(stack))
	}
}

func prepareAddBenchmark() (*sxeval.Environment, sxeval.Expr, *sxeval.Binding) {
	root := createBindingForTCO()
	obj := sx.MakeList(
		sx.MakeSymbol("let"),
		sx.MakeList(
			sx.MakeList(sx.MakeSymbol("a"), sx.Int64(5)),
			sx.MakeList(sx.MakeSymbol("x"), sx.Nil())),
		sx.MakeList(
			sx.MakeSymbol("if"),
			sx.MakeSymbol("a"),
			sx.MakeList(
				sx.MakeSymbol("if"),
				sx.MakeSymbol("x"),
				sx.Nil(),
				sx.MakeList(
					sx.MakeSymbol("*"),
					sx.MakeList(sx.MakeSymbol("+"), sx.MakeSymbol("a"), sx.Int64(4), sx.Int64(11), sx.Int64(13)),
					sx.Int64(17)),
			)),
	)

	env := sxeval.MakeEnvironment()
	expr, err := env.Parse(obj, root)
	if err != nil {
		panic(err)
	}
	return env, expr, root
}

func checkAddBenchmark(env *sxeval.Environment, expr sxeval.Expr, bind *sxeval.Binding) error {
	got, err := env.Run(expr, bind)
	if err != nil {
		return err
	}
	if exp := sx.Int64(561); !got.IsEqual(exp) {
		return fmt.Errorf("expected result %v, but got %v", exp, got)
	}
	if stack := env.Stack(); len(stack) > 0 {
		return fmt.Errorf("stack not empty: %v", stack)
	}
	return nil
}
