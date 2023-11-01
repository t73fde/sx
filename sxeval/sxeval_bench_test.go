//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval_test

import (
	"strconv"
	"testing"

	"zettelstore.de/sx.fossil"
)

func BenchmarkEvenTCO(b *testing.B) {
	testcases := [...]int{0, 1, 2, 4, 16, 64, 256, 1024, 4096, 16384, 65536}
	engine := createEngineForTCO()
	root := engine.GetToplevelEnv()
	evenSym := engine.SymbolFactory().MustMake("even?")
	b.ResetTimer()
	for _, tc := range testcases {
		b.Run(strconv.Itoa(tc), func(b *testing.B) {
			obj := sx.MakeList(evenSym, sx.Int64(tc))
			expr, err := engine.Parse(obj, root)
			if err != nil {
				panic(err)
			}
			expr = engine.Rework(expr, root)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				engine.Execute(expr, root)
			}
		})
	}
}
