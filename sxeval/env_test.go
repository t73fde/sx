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

func TestGetEnvironment(t *testing.T) {
	if _, ok := sxeval.GetEnvironment(nil); ok {
		t.Error("nil is not an environment")
	}
	if _, ok := sxeval.GetEnvironment(sxeval.Environment(nil)); ok {
		t.Error("nil environment is not an environment")
	}
	if _, ok := sxeval.GetEnvironment(sx.Nil()); ok {
		t.Error("Nil() is not an environment")
	}
	var o sx.Object = sxeval.MakeRootEnvironment(0)
	res, ok := sxeval.GetEnvironment(o)
	if !ok {
		t.Error("is an environment:", o)
	} else if !o.IsEqual(res) {
		t.Error("Different environments, expected:", o, "but got", res)
	}
}

func TestEnvRoot(t *testing.T) {
	t.Parallel()
	root := sxeval.MakeRootEnvironment(0)
	if got := root.Parent(); got != nil {
		t.Error("root env has a parent", got)
	}
	child := sxeval.MakeChildEnvironment(root, "child", 0)
	if got := child.Parent(); got != root {
		t.Error("root child parent is not root", got)
	}
	grandchild := sxeval.MakeChildEnvironment(child, "grandchild", 0)
	if got := grandchild.Parent(); got != child {
		t.Error("grandchild parent is not child", got)
	}
}

func TestBindLookupUnbind(t *testing.T) {
	t.Parallel()
	sf := sx.MakeMappedFactory()
	sym1 := sf.MustMake("sym1")
	sym2 := sf.MustMake("sym2")
	root := sxeval.MakeRootEnvironment(1)
	root.Bind(sym1, sym2)

	if val, found := root.Lookup(sym2); found {
		t.Errorf("Symbol %v should not be found, but resolves to %v", sym2, val)
	}

	t.Run("child", func(t *testing.T) {
		root := sxeval.MakeRootEnvironment(1)
		root.Bind(sym1, sym2)
		child := sxeval.MakeChildEnvironment(root, "assoc", 30)
		bindLookupUnbind(t, root, child, sym1, sym2)
	})
}

func bindLookupUnbind(t *testing.T, root, child sxeval.Environment, sym1, sym2 *sx.Symbol) {
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
	sf := sx.MakeMappedFactory()
	env := sxeval.MakeRootEnvironment(7)
	env.Bind(sf.MustMake("sym1"), sx.MakeString("sym1"))
	env.Bind(sf.MustMake("sym2"), sx.MakeString("sym2"))
	env.Bind(sf.MustMake("sym3"), sx.MakeString("sym3"))
	env.Bind(sf.MustMake("sym4"), sx.MakeString("sym4"))
	env.Bind(sf.MustMake("sym5"), sx.MakeString("sym5"))
	env.Bind(sf.MustMake("sym6"), sx.MakeString("sym6"))
	env.Bind(sf.MustMake("sym7"), sx.MakeString("sym7"))
	alist := env.Bindings()
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

func TestRootEnvEqual(t *testing.T) {
	t.Parallel()
	root1 := sxeval.MakeRootEnvironment(0)
	root2 := sxeval.MakeRootEnvironment(7)
	checkEnvEqual(t, root1, root2)

	root := sxeval.MakeRootEnvironment(3)
	child1 := sxeval.MakeChildEnvironment(root, "child1", 17)
	child2 := sxeval.MakeChildEnvironment(root, "child22", 11)
	checkEnvEqual(t, child1, child2)
}

func checkEnvEqual(t *testing.T, env1, env2 sxeval.Environment) {
	t.Helper()
	if !env1.IsEqual(env2) {
		t.Error("empty", env1, "is not equal to empty", env2)
		return
	}
	sf := sx.MakeMappedFactory()
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
