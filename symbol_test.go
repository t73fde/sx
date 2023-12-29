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

package sx_test

import (
	"testing"

	"zettelstore.de/sx.fossil"
)

func TestSymbolFactory(t *testing.T) {
	if got := sx.FindSymbolFactory(nil); got != nil {
		t.Errorf("symbol factory for nil should be nil, but is %T/%v", got, got)
	}
	if got := sx.FindSymbolFactory(sx.Int64(17)); got != nil {
		t.Errorf("symbol factory for 17 should be nil, but is %T/%v", got, got)
	}
	sf1 := sx.MakeMappedFactory(15)
	sym1_1 := sf1.MustMake("sym1")
	if got := sx.FindSymbolFactory(sym1_1); got != sf1 {
		t.Errorf("symbol factory for %v %T/%v expected, but got %T/%v", sym1_1, sf1, sf1, got, got)
	}
	lst := sx.MakeList(sx.Nil(), sx.Int64(1))
	if got := sx.FindSymbolFactory(lst); got != nil {
		t.Errorf("symbol factory for %s should be nil, but got %T/%v", lst, got, got)
	}
	lst = sx.MakeList(sx.Nil(), sx.Int64(1), sx.MakeList(sym1_1))
	if got := sx.FindSymbolFactory(lst); got != sf1 {
		t.Errorf("symbol factory for %s %T/%v expected, but got %T/%v", lst, sf1, sf1, got, got)
	}
	pair := sx.Cons(sx.Nil(), sym1_1)
	if got := sx.FindSymbolFactory(pair); got != sf1 {
		t.Errorf("symbol factory for %s %T/%v expected, but got %T/%v", pair, sf1, sf1, got, got)
	}
}
