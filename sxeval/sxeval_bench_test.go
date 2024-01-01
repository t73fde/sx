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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

func BenchmarkEvenTCO(b *testing.B) {
	testcases := [...]int{0, 1, 2, 4, 16, 64, 256, 1024, 4096, 16384, 65536}
	engine := createEngineForTCO()
	root := engine.GetToplevelBinding()
	evenSym := sx.Symbol("even?")
	b.ResetTimer()
	for _, tc := range testcases {
		b.Run(strconv.Itoa(tc), func(b *testing.B) {
			env := sxeval.MakeExecutionEnvironment(engine, nil, root)
			obj := sx.MakeList(evenSym, sx.Int64(tc))
			expr, err := env.Parse(obj)
			if err != nil {
				panic(err)
			}
			expr = env.Rework(expr)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				env.Run(expr)
			}
		})
	}
}

func BenchmarkFac(b *testing.B) {
	engine := createEngineForTCO()
	facSym := sx.Symbol("fac")
	root := engine.GetToplevelBinding()
	obj := sx.MakeList(facSym, sx.Int64(20))
	env := sxeval.MakeExecutionEnvironment(engine, nil, root)
	expr, err := env.Parse(obj)
	if err != nil {
		panic(err)
	}
	expr = env.Rework(expr)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		env.Run(expr)
	}
}
