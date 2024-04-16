//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"t73f.de/r/sx"
)

func TestDefaultSymbolFactory(t *testing.T) {
	t.Parallel()
	checkFactory(t, sx.DefaultSymbolFactory())
}

func checkFactory(t *testing.T, fac *sx.SymbolFactory) {
	symA := fac.MakeSymbol("A")
	symB := fac.MakeSymbol("B")
	if symA == symB {
		t.Errorf("symbols %v and %v are treated as identical, but are not", symA, symB)
	}
	if sym := fac.MakeSymbol("A"); sym != symA {
		t.Errorf("symbol %v and %v should be identical, but are not", symA, sym)
	}
	if sym := fac.MakeSymbol(""); sym != nil {
		t.Errorf("symbol with no value must result in nil, but got %v", sym)
	}
}

func TestOneSymbolFactory(t *testing.T) {
	t.Parallel()
	var sf sx.SymbolFactory

	if got := sf.Size(); got != 0 {
		t.Errorf("new symbol factory must not manage symbol, but does it for %d symbols", got)
	}
	checkFactory(t, &sf)
	if got := sf.Size(); got != 2 {
		t.Errorf("new symbol factory must 2 symbols, but does it for %d symbols", got)
	}
}

func TestTwoFactories(t *testing.T) {
	t.Parallel()
	var sf1, sf2 sx.SymbolFactory
	sym1 := sf1.MakeSymbol("A")
	sym2 := sf2.MakeSymbol("A")
	if sym1 == sym2 || sym1.IsEqual(sym2) {
		t.Errorf("symbols %v and %v came from two different factories, but are treated equal", sym1, sym2)
	}
}

func TestChildFactory(t *testing.T) {
	t.Parallel()
	var sf sx.SymbolFactory
	childSF := sf.NewChild()
	sym1 := childSF.MakeSymbol("A")
	childSF.MoveSymbols()
	sym2 := childSF.MakeSymbol("A")
	if sym1 != sym2 {
		t.Errorf("symbols %v and %v must be identical, but are not because of error in child.MakeSymbol", sym1, sym2)
	}
	sym3 := sf.MakeSymbol("A")
	if sym3 != sym1 {
		t.Errorf("child.MoveSymbols did not work: %v vs %v", sym1, sym3)
	}

}
