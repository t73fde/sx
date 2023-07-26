//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf_test

import (
	"testing"

	"zettelstore.de/sx.fossil/sxpf"
)

func TestGetEnvironment(t *testing.T) {
	if _, ok := sxpf.GetEnvironment(nil); ok {
		t.Error("nil is not an environment")
	}
	if _, ok := sxpf.GetEnvironment(sxpf.Environment(nil)); ok {
		t.Error("nil environment is not an environment")
	}
	if _, ok := sxpf.GetEnvironment(sxpf.Nil()); ok {
		t.Error("Nil() is not an environment")
	}
	var o sxpf.Object = sxpf.MakeRootEnvironment()
	res, ok := sxpf.GetEnvironment(o)
	if !ok {
		t.Error("is an environment:", o)
	} else if !o.IsEqual(res) {
		t.Error("Different environments, expected:", o, "but got", res)
	}
}

func TestEnvRoot(t *testing.T) {
	t.Parallel()
	root := sxpf.MakeRootEnvironment()
	if got := root.Parent(); got != nil {
		t.Error("root env has a parent", got)
	}
	child := sxpf.MakeChildEnvironment(root, "child", 0)
	if got := child.Parent(); got != root {
		t.Error("root child parent is not root", got)
	}
	grandchild := sxpf.MakeChildEnvironment(child, "grandchild", 0)
	if got := grandchild.Parent(); got != child {
		t.Error("grandchild parent is not child", got)
	}
}

func TestBindLookupUnbind(t *testing.T) {
	t.Parallel()
	sf := sxpf.MakeMappedFactory()
	sym1 := sf.MustMake("sym1")
	sym2 := sf.MustMake("sym2")
	root := sxpf.MakeRootEnvironment()
	root.Bind(sym1, sym2)

	val, found := root.Lookup(sym2)
	if found {
		t.Errorf("Symbol %v should not be found, but resolves to %v", sym2, val)
	}

	child := sxpf.MakeChildEnvironment(root, "child1", 0)
	child.Bind(sym2, sym1)

	_, found = child.Lookup(sym1)
	if found {
		t.Error("Symbol", sym1, "was found in child")
	}
	val, found = sxpf.Resolve(child, sym1)
	if !found {
		t.Error("Symbol", sym1, "not resolved")
	} else if val != sym2 {
		t.Errorf("Symbol %v should resolve to %v, but got %v", sym1, sym2, val)
	}

	val, found = child.Lookup(sym2)
	if !found {
		t.Error("Symbol", sym2, "not found")
	} else if val != sym1 {
		t.Errorf("Symbol %v should resolve to %v, but got %v", sym2, sym1, val)
	}
	val, found = sxpf.Resolve(child, sym2)
	if !found {
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
	bindings := sxpf.AllBindings(child)
	if cc := bindings.Assoc(sym1); cc == nil {
		t.Error("Symbol", sym1, "not found in bindings")
	}
	if cc := bindings.Assoc(sym2); cc == nil {
		t.Error("Symbol", sym2, "not found in bindings")
	}

	root.Unbind(sym1)
	_, found = root.Lookup(sym1)
	if found {
		t.Error("Symbol", sym1, "was found in root")
	}
	child.Unbind(sym2)
	_, found = child.Lookup(sym2)
	if found {
		t.Error("Symbol", sym2, "was found in child")
	}
}

func TestAlist(t *testing.T) {
	t.Parallel()
	sf := sxpf.MakeMappedFactory()
	env := sxpf.MakeRootEnvironment()
	env.Bind(sf.MustMake("sym1"), sxpf.MakeString("sym1"))
	env.Bind(sf.MustMake("sym2"), sxpf.MakeString("sym2"))
	env.Bind(sf.MustMake("sym3"), sxpf.MakeString("sym3"))
	env.Bind(sf.MustMake("sym4"), sxpf.MakeString("sym4"))
	env.Bind(sf.MustMake("sym5"), sxpf.MakeString("sym5"))
	env.Bind(sf.MustMake("sym6"), sxpf.MakeString("sym6"))
	env.Bind(sf.MustMake("sym7"), sxpf.MakeString("sym7"))
	alist := env.Bindings()
	if alist.Length() != 7 {
		t.Error("Not 7 elements:", alist)
		return
	}
	cnt := 0
	for elem := alist; elem != nil; elem = elem.Tail() {
		cnt++
		cons := elem.Car().(*sxpf.Pair)
		sym := cons.Car().(*sxpf.Symbol)
		str := cons.Cdr().(sxpf.String)
		if got := sf.MustMake(str.String()); !sym.IsEqual(got) {
			t.Error("Symbol", sym, "is not equal to", str, "but to", got)
		}
	}
	if cnt != 7 {
		t.Error("Count is not 7:", cnt)
	}
}

func TestRootEnvEqual(t *testing.T) {
	t.Parallel()
	root1 := sxpf.MakeRootEnvironment()
	root2 := sxpf.MakeRootEnvironment()
	checkEnvEqual(t, root1, root2)

	root := sxpf.MakeRootEnvironment()
	child1 := sxpf.MakeChildEnvironment(root, "child1", 7)
	child2 := sxpf.MakeChildEnvironment(root, "child2", 1)
	checkEnvEqual(t, child1, child2)
}

func checkEnvEqual(t *testing.T, env1, env2 sxpf.Environment) {
	if !env1.IsEqual(env2) {
		t.Error("empty", env1, "is not equal to empty", env2)
		return
	}
	sf := sxpf.MakeMappedFactory()
	sym1 := sf.MustMake("sym1")
	env1.Bind(sym1, sym1)
	if env1.IsEqual(env2) {
		t.Error("after adding sym1 just to", env1, "both envs are equal")
		return
	}
	sym2 := sf.MustMake("sym2")
	env2.Bind(sym2, sym2)
	if env1.IsEqual(env2) {
		t.Error("after adding sym2 just to", env2, "both envs are equal")
		return
	}
	env1.Bind(sym2, sym1)
	env2.Bind(sym1, sym2)
	if env1.IsEqual(env2) {
		t.Error("envs are equal, but bindings differ")
		return
	}
	env1.Bind(sym2, sym2)
	env2.Bind(sym1, sym1)
	if !env1.IsEqual(env2) {
		t.Error("equal envs differ")
	}
}
