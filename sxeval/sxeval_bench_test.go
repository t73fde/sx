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
	"strconv"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

func BenchmarkEvenTCO(b *testing.B) {
	testcases := [...]int{0, 1, 2, 4, 16, 64, 256, 1024, 4096, 16384, 65536}
	root := createBindingForTCO()
	evenSym := sx.MakeSymbol("even?")
	b.ResetTimer()
	for _, tc := range testcases {
		b.Run(strconv.Itoa(tc), func(b *testing.B) {
			env := sxeval.MakeExecutionEnvironment(root)
			obj := sx.MakeList(evenSym, sx.Int64(tc))
			expr, err := env.Parse(obj)
			if err != nil {
				panic(err)
			}
			b.ResetTimer()
			for b.Loop() {
				_, _ = env.Run(expr)
			}
		})
	}
}

func BenchmarkFac(b *testing.B) {
	root := createBindingForTCO()
	facSym := sx.MakeSymbol("fac")
	obj := sx.MakeList(facSym, sx.Int64(20))
	env := sxeval.MakeExecutionEnvironment(root)
	expr, err := env.Parse(obj)
	if err != nil {
		panic(err)
	}

	for b.Loop() {
		_, _ = env.Run(expr)
	}
}

func BenchmarkFaa(b *testing.B) {
	root := createBindingForTCO()
	faaSym := sx.MakeSymbol("faa")
	obj := sx.MakeList(faaSym, sx.Int64(20), sx.Int64(1))
	env := sxeval.MakeExecutionEnvironment(root)
	expr, err := env.Parse(obj)
	if err != nil {
		panic(err)
	}

	for b.Loop() {
		_, _ = env.Run(expr)
	}
}

func BenchmarkFib(b *testing.B) {
	root := createBindingForTCO()
	fibSym := sx.MakeSymbol("fib")
	obj := sx.MakeList(fibSym, sx.Int64(10))
	env := sxeval.MakeExecutionEnvironment(root)
	expr, err := env.Parse(obj)
	if err != nil {
		panic(err)
	}

	for b.Loop() {
		_, _ = env.Run(expr)
	}
}
