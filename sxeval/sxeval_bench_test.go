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
			env := sxeval.MakeComputeEnvironment(root)
			obj := sx.MakeList(evenSym, sx.Int64(tc))
			expr, err := env.Parse(obj)
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
	env := sxeval.MakeComputeEnvironment(root)
	expr, err := env.Parse(sexpr)
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
