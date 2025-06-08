//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"t73f.de/r/sx"
)

func TestMakePackage(t *testing.T) {
	testcases := []struct {
		name string
		exp  string
	}{
		{"my-test", "#<package:my-test>"},
		{"my-test", "package \"my-test\" already made"},
		{"", "invalid package name: \"\""},
		{"a", "#<package:a>"},
		{"a-b", "#<package:a-b>"},
		{"-b", "invalid package name: \"-b\""},
		{"a b", "invalid package name: \"a b\""},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pkg, err := sx.MakePackage(tc.name)

			var got string
			if err != nil {
				got = err.Error()
			} else {
				got = pkg.String()
			}
			if got != tc.exp {
				t.Errorf("MakePackage(%q) should result in %q, but got %q", tc.name, tc.exp, got)
			}
		})
	}
}

func TestCurrentPackage(t *testing.T) {
	t.Parallel()
	checkPackage(t, sx.CurrentPackage())
}

func checkPackage(t *testing.T, pkg *sx.Package) {
	symA := pkg.MakeSymbol("A")
	symB := pkg.MakeSymbol("B")
	if symA == symB {
		t.Errorf("symbols %v and %v are treated as identical, but are not", symA, symB)
	}
	if sym := pkg.MakeSymbol("A"); sym != symA {
		t.Errorf("symbol %v and %v should be identical, but are not", symA, sym)
	}
	if sym := pkg.MakeSymbol(""); sym != nil {
		t.Errorf("symbol with no value must result in nil, but got %v", sym)
	}
	if sym := pkg.FindSymbol("A"); sym != symA {
		t.Errorf("found symbol %v and %v should be identical, but are not", symA, sym)
	}
	if sym := pkg.FindSymbol(""); sym != nil {
		t.Errorf("found symbol with no value must result in nil, but got %v", sym)
	}
}

func TestOnePackage(t *testing.T) {
	t.Parallel()
	pkg := sx.MustMakePackage("uno")

	if got := pkg.Size(); got != 0 {
		t.Errorf("new package must not manage symbol, but does it for %d symbols", got)
	}
	checkPackage(t, pkg)
	if got := pkg.Size(); got != 2 {
		t.Errorf("new package must 2 symbols, but does it for %d symbols", got)
	}
}

func TestTwoPackages(t *testing.T) {
	t.Parallel()
	pkg1 := sx.MustMakePackage("one")
	pkg2 := sx.MustMakePackage("two")
	sym1 := pkg1.MakeSymbol("A")
	sym2 := pkg2.MakeSymbol("A")
	if sym1 == sym2 || sym1.IsEqual(sym2) {
		t.Errorf("symbols %v and %v came from two different factories, but are treated equal", sym1, sym2)
	}
	checkPackage(t, pkg1)
	checkPackage(t, pkg2)
}
