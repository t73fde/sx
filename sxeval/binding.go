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

package sxeval

import (
	"fmt"

	"zettelstore.de/sx.fossil"
)

// ErrBindingFrozen is returned when trying to update a frozen binding.
type ErrBindingFrozen struct{ Binding *Binding }

func (err ErrBindingFrozen) Error() string { return fmt.Sprintf("binding is frozen: %v", err.Binding) }

type mapSymObj = map[string]sx.Object

// Binding is a binding based on maps.
type Binding struct {
	vars   mapSymObj
	parent *Binding
	name   string
	frozen bool
}

func makeBinding(name string, parent *Binding, sizeHint int) *Binding {
	if sizeHint <= 0 {
		sizeHint = 3
	}
	return &Binding{
		vars:   make(mapSymObj, sizeHint),
		parent: parent,
		name:   name,
		frozen: false,
	}
}

// MakeRootBinding creates a new root binding.
func MakeRootBinding(sizeHint int) *Binding {
	return makeBinding("root", nil, sizeHint)
}

// MakeChildBinding creates a new binding with a given parent.
func (b *Binding) MakeChildBinding(name string, sizeHint int) *Binding {
	return makeBinding(name, b, sizeHint)
}

func (b *Binding) IsNil() bool  { return b == nil }
func (b *Binding) IsAtom() bool { return b == nil }
func (b *Binding) IsEqual(other sx.Object) bool {
	if b == other {
		return true
	}
	if b.IsNil() {
		return sx.IsNil(other)
	}
	if ob, isBinding := other.(*Binding); isBinding {
		bvars, obvars := b.vars, ob.vars
		if len(bvars) != len(obvars) {
			return false
		}
		for k, v := range bvars {
			ov, found := obvars[k]
			if !found || !v.IsEqual(ov) {
				return false
			}
		}
		return true
	}
	return false
}
func (b *Binding) String() string {
	return fmt.Sprintf("#<binding:%s/%d>", b.name, len(b.vars))
}

func (b *Binding) GoString() string { return b.String() }

// Name returns the local name of this binding.
func (b *Binding) Name() string { return b.name }

// Parent returns the parent binding.
func (b *Binding) Parent() *Binding {
	if b == nil {
		return nil
	}
	return b.parent
}

// Bind creates a local mapping with a given symbol and object.
//
// A previous, non-const mapping will be overwritten.
func (b *Binding) Bind(sym *sx.Symbol, val sx.Object) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	b.vars[sym.GetValue()] = val
	return nil
}

func (b *Binding) BindSpecial(syn *Special) error {
	return b.Bind(sx.MakeSymbol(syn.Name), syn)
}

// BindBuiltin binds the given builtin with its given name.
func (b *Binding) BindBuiltin(bi *Builtin) error {
	return b.Bind(sx.MakeSymbol(bi.Name), bi)
}

// Lookup will search for a local binding of the given symbol. If not
// found, the search will *not* be continued in the parent binding.
// Use the global `Resolve` function, if you want a search up to the parent.
func (b *Binding) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	obj, found := b.vars[sym.GetValue()]
	return obj, found
}

// LookupN will lookup the symbol in the N-th parent.
func (b *Binding) LookupN(sym *sx.Symbol, n int) (sx.Object, bool) {
	for i := 0; i < n; i++ {
		b = b.parent
	}
	return b.Lookup(sym)
}

// Bindings returns all bindings as an a-list in some random order.
func (b *Binding) Bindings() *sx.Pair {
	result := sx.Nil()
	for k, v := range b.vars {
		result = result.Cons(sx.Cons(sx.MakeSymbol(k), v))
	}
	return result
}

// Unbind removes the mapping of the given symbol to an object.
func (b *Binding) Unbind(sym *sx.Symbol) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	delete(b.vars, sym.GetValue())
	return nil
}

// Freeze sets the binding in a read-only state.
func (b *Binding) Freeze() { b.frozen = true }

// IsFrozen returns true if binding is frozen.
func (b *Binding) IsFrozen() bool { return b.frozen }

// resolveFull resolves a symbol and returns its possible binding, an
// indication about its const-ness and a reference to the binding where
// the symbol was bound.
func (b *Binding) resolveFull(sym *sx.Symbol) (sx.Object, *Binding, int) {
	depth := 0
	for curr := b; curr != nil; curr = curr.parent {
		if obj, found := curr.Lookup(sym); found {
			return obj, curr, depth
		}
		depth++
	}
	return nil, nil, depth
}

// Resolve a symbol in a binding and all of its parent bindings.
func (b *Binding) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	for curr := b; curr != nil; curr = curr.parent {
		if obj, found := curr.Lookup(sym); found {
			return obj, true
		}
	}
	return nil, false
}

// ResolveN resolves a symbol in the N-th parent binding and all of its parent
// bindings.
func (b *Binding) ResolveN(sym *sx.Symbol, n int) (sx.Object, bool) {
	curr := b
	for i := 0; i < n; i++ {
		curr = curr.parent
	}
	for ; curr != nil; curr = curr.parent {
		if obj, found := curr.Lookup(sym); found {
			return obj, true
		}
	}
	return nil, false
}

// GetBinding returns the object as a binding, if possible.
func GetBinding(obj sx.Object) (*Binding, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	b, ok := obj.(*Binding)
	return b, ok
}
