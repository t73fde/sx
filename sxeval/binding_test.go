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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

func TestGetBinding(t *testing.T) {
	if _, ok := sxeval.GetBinding(nil); ok {
		t.Error("nil is not a binding")
	}
	if _, ok := sxeval.GetBinding(sxeval.Binding(nil)); ok {
		t.Error("nil binding is not a binding")
	}
	if _, ok := sxeval.GetBinding(sx.Nil()); ok {
		t.Error("Nil() is a binding")
	}
	var o sx.Object = sxeval.MakeRootBinding(0)
	res, ok := sxeval.GetBinding(o)
	if !ok {
		t.Error("is a binding:", o)
	} else if !o.IsEqual(res) {
		t.Error("Different bindings, expected:", o, "but got", res)
	}
}

func TestBindingRoot(t *testing.T) {
	t.Parallel()
	root := sxeval.MakeRootBinding(0)
	if got := root.Parent(); got != nil {
		t.Error("root binding has a parent", got)
	}
	child := sxeval.MakeChildBinding(root, "child", 0)
	if got := child.Parent(); got != root {
		t.Error("root child parent is not root", got)
	}
	grandchild := sxeval.MakeChildBinding(child, "grandchild", 0)
	if got := grandchild.Parent(); got != child {
		t.Error("grandchild parent is not child", got)
	}
}

func TestBindLookupUnbind(t *testing.T) {
	t.Parallel()
	sf := sx.MakeMappedFactory(2)
	sym1 := sf.MustMake("sym1")
	sym2 := sf.MustMake("sym2")
	root := sxeval.MakeRootBinding(1)
	root.Bind(sym1, sym2)

	if val, found := root.Lookup(sym2); found {
		t.Errorf("Symbol %v should not be found, but resolves to %v", sym2, val)
	}

	t.Run("child", func(t *testing.T) {
		newRoot := sxeval.MakeRootBinding(1)
		newRoot.Bind(sym1, sym2)
		child := sxeval.MakeChildBinding(newRoot, "assoc", 30)
		bindLookupUnbind(t, newRoot, child, sym1, sym2)
	})
}

func bindLookupUnbind(t *testing.T, root, child sxeval.Binding, sym1, sym2 *sx.Symbol) {
	child.Bind(sym2, sym1)

	if _, found := child.Lookup(sym1); found {
		t.Error("Symbol", sym1, "was found in child")
	}
	if val, found := sxeval.Resolve(child, sym1); !found {
		t.Error("Symbol", sym1, "not resolved")
	} else if val != sym2 {
		t.Errorf("Symbol %v should resolve to %v, but got %v", sym1, sym2, val)
	}

	if val, found := child.Lookup(sym2); !found {
		t.Error("Symbol", sym2, "not found")
	} else if val != sym1 {
		t.Errorf("Symbol %v should resolve to %v, but got %v", sym2, sym1, val)
	}
	if val, found := sxeval.Resolve(child, sym2); !found {
		t.Error("Symbol", sym2, "not resolved")
	} else if val != sym1 {
		t.Errorf("Symbol %v should resolve to %v, but got %v", sym2, sym1, val)
	}

	if cc := root.Bindings().Assoc(sym1); cc == nil {
		t.Error("Symbol", sym1, "not found in root bindings")
	}
	if cc := child.Bindings().Assoc(sym2); cc == nil {
		t.Error("Symbol", sym2, "not found in child bindings")
	}
	bindings := sxeval.AllBindings(child)
	if cc := bindings.Assoc(sym1); cc == nil {
		t.Error("Symbol", sym1, "not found in bindings")
	}
	if cc := bindings.Assoc(sym2); cc == nil {
		t.Error("Symbol", sym2, "not found in bindings")
	}

	root.Unbind(sym1)
	if _, found := root.Lookup(sym1); found {
		t.Error("Symbol", sym1, "was found in root")
	}
	child.Unbind(sym2)
	if _, found := child.Lookup(sym2); found {
		t.Error("Symbol", sym2, "was found in child")
	}
}

func TestAlist(t *testing.T) {
	t.Parallel()
	sf := sx.MakeMappedFactory(7)
	bind := sxeval.MakeRootBinding(7)
	bind.Bind(sf.MustMake("sym1"), sx.String("sym1"))
	bind.Bind(sf.MustMake("sym2"), sx.String("sym2"))
	bind.Bind(sf.MustMake("sym3"), sx.String("sym3"))
	bind.Bind(sf.MustMake("sym4"), sx.String("sym4"))
	bind.Bind(sf.MustMake("sym5"), sx.String("sym5"))
	bind.Bind(sf.MustMake("sym6"), sx.String("sym6"))
	bind.Bind(sf.MustMake("sym7"), sx.String("sym7"))
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
		if got := sf.MustMake(str.String()); !sym.IsEqual(got) {
			t.Error("Symbol", sym, "is not equal to", str, "but to", got)
		}
	}
	if cnt != 7 {
		t.Error("Count is not 7:", cnt)
	}
}

func TestRootBindingEqual(t *testing.T) {
	t.Parallel()
	root1 := sxeval.MakeRootBinding(0)
	root2 := sxeval.MakeRootBinding(7)
	checkBindingEqual(t, root1, root2)

	root := sxeval.MakeRootBinding(3)
	child1 := sxeval.MakeChildBinding(root, "child1", 17)
	child2 := sxeval.MakeChildBinding(root, "child22", 11)
	checkBindingEqual(t, child1, child2)
}

func checkBindingEqual(t *testing.T, bind1, bind2 sxeval.Binding) {
	t.Helper()
	if !bind1.IsEqual(bind2) {
		t.Error("empty", bind1, "is not equal to empty", bind2)
		return
	}
	sf := sx.MakeMappedFactory(2)
	sym1 := sf.MustMake("sym1")
	bind1.Bind(sym1, sym1)
	if bind1.IsEqual(bind2) {
		t.Error("after adding sym1 just to", bind1, "both bindings are equal")
		return
	}
	sym2 := sf.MustMake("sym2")
	bind2.Bind(sym2, sym2)
	if bind1.IsEqual(bind2) {
		t.Error("after adding sym2 just to", bind2, "both bindings are equal")
		return
	}
	bind1.Bind(sym2, sym1)
	bind2.Bind(sym1, sym2)
	if bind1.IsEqual(bind2) {
		t.Error("bindings are equal, but bindings differ")
		return
	}
	bind1.Bind(sym2, sym2)
	bind2.Bind(sym1, sym1)
	if !bind1.IsEqual(bind2) {
		t.Error("equal bindings differ")
	}
}

func TestConstBinding(t *testing.T) {
	t.Parallel()
	sf := sx.MakeMappedFactory(2)
	bind := sxeval.MakeRootBinding(2)
	symVar, symConst := sf.MustMake("sym-var"), sf.MustMake("sym-const")
	err := bind.Bind(symVar, sx.Int64(0))
	if err != nil {
		t.Errorf("error in symVar.Bind(1): %v", err)
		return
	}
	if bind.IsConst(symVar) {
		t.Errorf("symbol %v is wrongly seen a constant", symVar)
		return
	}
	err = bind.BindConst(symVar, sx.Int64(2))
	if err != nil {
		t.Error("symVar was not changed, but should:", err)
	}

	err = bind.BindConst(symConst, sx.Int64(0))
	if err != nil {
		t.Errorf("error in symConst.Bind(1): %v", err)
		return
	}
	if !bind.IsConst(symConst) {
		t.Errorf("symbol %v is wrongly seen a variable", symConst)
		return
	}
	err = bind.BindConst(symConst, sx.Int64(2))
	if err == nil {
		t.Error("symConst was changed, but should not")
	}
}
