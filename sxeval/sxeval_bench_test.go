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
			env := sxeval.MakeExecutionEnvironment(root)
			obj := sx.MakeList(evenSym, sx.Int64(tc))
			expr, err := parseAndMore(env, obj)
			if err != nil {
				b.Error(err)
			}
			b.ResetTimer()
			for b.Loop() {
				_, _ = env.Run(expr)
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
	env := sxeval.MakeExecutionEnvironment(root)
	expr, err := parseAndMore(env, sexpr)
	if err != nil {
		b.Error(err)
	}

	if _, err = env.Run(expr); err != nil {
		b.Error(err)
	}
	if stack := env.Stack(); len(stack) > 0 {
		b.Error("stack not empty, but:", stack)
	}
	for b.Loop() {
		_, _ = env.Run(expr)
	}
	if stack := env.Stack(); len(stack) > 0 {
		b.Error("stack not empty, found", len(stack), "elements")
	}
}

func parseAndMore(env *sxeval.Environment, obj sx.Object) (sxeval.Expr, error) {
	expr, err := env.Parse(obj)
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
	env, expr := prepareAddBenchmark()

	err := checkAddBenchmark(env, expr)
	if err != nil {
		b.Error(err)
		return
	}

	for b.Loop() {
		_, _ = env.Run(expr)
	}
}

func BenchmarkAddCompiled(b *testing.B) {
	env, expr := prepareAddBenchmark()
	cexpr, err := env.Compile(expr)
	if err != nil {
		b.Error(err)
		return
	}

	err = checkAddBenchmark(env, cexpr)
	if err != nil {
		b.Error(err)
		return
	}

	for b.Loop() {
		_, _ = env.Run(cexpr)
	}
	if stack := env.Stack(); len(stack) > 0 {
		b.Error("stack not empty", len(stack))
	}
}

func prepareAddBenchmark() (*sxeval.Environment, sxeval.Expr) {
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

	env := sxeval.MakeExecutionEnvironment(root)
	expr, err := env.Parse(obj)
	if err != nil {
		panic(err)
	}
	return env, expr
}

func checkAddBenchmark(env *sxeval.Environment, expr sxeval.Expr) error {
	got, err := env.Run(expr)
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
