//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval_test

import (
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

func TestBindLookupUnbind(t *testing.T) {
	t.Parallel()
	sym1 := sx.MakeSymbol("sym1")
	sym2 := sx.MakeSymbol("sym2")
	root := sxeval.MakeRootBinding(1)

	if val, found := root.Lookup(nil); found {
		t.Errorf("nil symbol should not be found, but got: %v", val)
	}

	_ = root.Bind(sym1, sym2)

	if val, found := root.Lookup(sym2); found {
		t.Errorf("Symbol %v should not be found, but resolves to %v", sym2, val)
	}

	t.Run("child", func(t *testing.T) {
		newRoot := sxeval.MakeRootBinding(1)
		_ = newRoot.Bind(sym1, sym2)
		child := newRoot.MakeChildBinding("assoc", 30)
		bindLookup(t, newRoot, child, sym1, sym2)
	})
}

func bindLookup(t *testing.T, root, child *sxeval.Binding, sym1, sym2 *sx.Symbol) {
	_ = child.Bind(sym2, sym1)

	if _, found := child.Lookup(sym1); found {
		t.Error("Symbol", sym1, "was found in child")
	}

	if val, found := child.Lookup(sym2); !found {
		t.Error("Symbol", sym2, "not found")
	} else if val != sym1 {
		t.Errorf("Symbol %v should resolve to %v, but got %v", sym2, sym1, val)
	}

	if cc := root.Bindings().Assoc(sym1); cc == nil {
		t.Error("Symbol", sym1, "not found in root bindings")
	}
	if cc := child.Bindings().Assoc(sym2); cc == nil {
		t.Error("Symbol", sym2, "not found in child bindings")
	}
}

func TestAlist(t *testing.T) {
	t.Parallel()
	bind := sxeval.MakeRootBinding(7)
	_ = bind.Bind(sx.MakeSymbol("sym1"), sx.MakeString("sym1"))
	_ = bind.Bind(sx.MakeSymbol("sym2"), sx.MakeString("sym2"))
	_ = bind.Bind(sx.MakeSymbol("sym3"), sx.MakeString("sym3"))
	_ = bind.Bind(sx.MakeSymbol("sym4"), sx.MakeString("sym4"))
	_ = bind.Bind(sx.MakeSymbol("sym5"), sx.MakeString("sym5"))
	_ = bind.Bind(sx.MakeSymbol("sym6"), sx.MakeString("sym6"))
	_ = bind.Bind(sx.MakeSymbol("sym7"), sx.MakeString("sym7"))
	alist := bind.Bindings()
	if alist.Length() != 7 {
		t.Error("Not 7 elements:", alist)
		return
	}
	cnt := 0
	for elem := alist; elem != nil; elem = elem.Tail() {
		cnt++
		cons := elem.Car().(*sx.Pair)
		sym := cons.Car().(*sx.Symbol)
		str := cons.Cdr().(sx.String)
		if got := sx.MakeSymbol(str.GetValue()); !sym.IsEqual(got) {
			t.Error("Symbol", sym, "is not equal to", str, "but to", got)
		}
	}
	if cnt != 7 {
		t.Error("Count is not 7:", cnt)
	}
}

func TestRootBindingEqual(t *testing.T) {
	t.Parallel()
	root1 := sxeval.MakeRootBinding(1)
	root2 := sxeval.MakeRootBinding(7)
	checkBindingEqual(t, root1, root2)

	root := sxeval.MakeRootBinding(3)
	child1 := root.MakeChildBinding("child1", 7)
	child2 := root.MakeChildBinding("child22", 1)
	checkBindingEqual(t, child1, child2)
}

func checkBindingEqual(t *testing.T, bind1, bind2 *sxeval.Binding) {
	t.Helper()
	if !bind1.IsEqual(bind2) {
		t.Error("empty", bind1, "is not equal to empty", bind2)
		return
	}
	sym1 := sx.MakeSymbol("sym1")
	_ = bind1.Bind(sym1, sym1)
	if bind1.IsEqual(bind2) {
		t.Error("after adding sym1 just to", bind1, "both bindings are equal")
		return
	}
	sym2 := sx.MakeSymbol("sym2")
	_ = bind2.Bind(sym2, sym2)
	if bind1.IsEqual(bind2) {
		t.Error("after adding sym2 just to", bind2, "both bindings are equal")
		return
	}
	_ = bind1.Bind(sym2, sym1)
	_ = bind2.Bind(sym1, sym2)
	if bind1.IsEqual(bind2) {
		t.Error("bindings are equal, but bindings differ")
		return
	}
	_ = bind1.Bind(sym2, sym2)
	_ = bind2.Bind(sym1, sym1)
	if !bind1.IsEqual(bind2) {
		t.Error("equal bindings differ")
	}
}
