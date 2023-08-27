//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval_test

import (
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

func BenchmarkEnv(b *testing.B) {
	sf := sx.MakeMappedFactory()
	symA, symB, symC := sf.MustMake("a"), sf.MustMake("b"), sf.MustMake("c")
	root := sxeval.MakeRootEnvironment(0)
	root.Bind(symA, symB)
	fixed1 := sxeval.MakeFixedEnvironment(root, "fixed-1", 1)
	fixed1.Bind(symB, symC)
	fixed77 := sxeval.MakeFixedEnvironment(root, "fixed-77", 77)
	fixed77.Bind(symB, symC)
	child77 := sxeval.MakeChildEnvironment(root, "child-77", 77)
	child77.Bind(symB, symC)

	uuts := []sxeval.Environment{root, fixed1, fixed77, child77}
	b.ResetTimer()
	for _, uut := range uuts {
		b.Run("lookup/"+uut.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				uut.Lookup(symA)
				uut.Lookup(symB)
				uut.Lookup(symC)
			}
		})
	}
	for _, uut := range uuts {
		b.Run("resolve/"+uut.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sxeval.Resolve(uut, symA)
				sxeval.Resolve(uut, symB)
				sxeval.Resolve(uut, symC)
			}
		})
	}
}
