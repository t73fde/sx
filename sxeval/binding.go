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

// ErrConstBinding is returned when a constant binding should be changed.
type ErrConstBinding struct{ Sym sx.Symbol }

func (err ErrConstBinding) Error() string {
	return fmt.Sprintf("constant bindung for symbol %v", err.Sym)
}

type mapSymObj = map[sx.Symbol]sx.Object

// MakeRootBinding creates a new root binding.
func MakeRootBinding(sizeHint int) *Binding {
	return &Binding{
		name:   "root",
		parent: nil,
		vars:   make(mapSymObj, sizeHint),
		isRoot: true,
		frozen: false,
	}
}

// MakeChildBinding creates a new binding with a given parent.
func MakeChildBinding(parent *Binding, name string, sizeHint int) *Binding {
	if sizeHint <= 0 {
		sizeHint = 3
	}
	return &Binding{
		name:   name,
		parent: parent,
		vars:   make(mapSymObj, sizeHint),
		isRoot: false,
		frozen: false,
	}
}

// Binding is a binding based on maps.
type Binding struct {
	name   string
	parent *Binding
	vars   mapSymObj
	consts map[sx.Symbol]struct{}
	isRoot bool
	frozen bool
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
	if omb, ok := other.(*Binding); ok {
		mvars, ovars := b.vars, omb.vars
		if len(mvars) != len(ovars) {
			return false
		}
		for k, v := range mvars {
			ov, found := ovars[k]
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
func (b *Binding) Bind(sym sx.Symbol, val sx.Object) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	if b.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if _, found := b.vars[sym]; found {
		b.vars[sym] = val
		return nil
	}
	b.vars[sym] = val
	return nil
}

// BindConst creates a local mapping of the symbol to the object, which
// cannot be changed afterwards.
func (b *Binding) BindConst(sym sx.Symbol, val sx.Object) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	if b.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if b.consts == nil {
		b.consts = map[sx.Symbol]struct{}{sym: {}}
	} else {
		b.consts[sym] = struct{}{}
	}
	if _, found := b.vars[sym]; found {
		b.vars[sym] = val
		return nil
	}
	b.vars[sym] = val
	return nil
}

func (b *Binding) BindSpecial(syn *Special) error {
	return b.BindConst(sx.MakeSymbol(syn.Name), syn)
}

// BindBuiltin binds the given builtin with its given name.
func (b *Binding) BindBuiltin(bi *Builtin) error {
	return b.BindConst(sx.MakeSymbol(bi.Name), bi)
}

// Lookup will search for a local binding of the given symbol. If not
// found, the search will *not* be continued in the parent binding.
// Use the global `Resolve` function, if you want a search up to the parent.
func (b *Binding) Lookup(sym sx.Symbol) (sx.Object, bool) {
	obj, found := b.vars[sym]
	return obj, found
}

// IsConst returns true if the binding of the symbol is a constant binding.
func (b *Binding) IsConst(sym sx.Symbol) bool {
	if b == nil {
		return false
	}
	if b.frozen {
		if _, found := b.vars[sym]; found {
			return true
		}
	}
	if b.consts == nil {
		return false
	}
	_, found := b.consts[sym]
	return found
}

// Bindings returns all bindings as an a-list in some random order.
func (b *Binding) Bindings() *sx.Pair {
	result := sx.Nil()
	for k, v := range b.vars {
		result = result.Cons(sx.Cons(k, v))
	}
	return result
}

// Unbind removes the mapping of the given symbol to an object.
func (b *Binding) Unbind(sym sx.Symbol) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	delete(b.vars, sym)
	return nil
}

// Freeze sets the binding in a read-only state.
func (b *Binding) Freeze() { b.frozen = true }

// Resolve a symbol in a binding and all of its parent bindings.
func (b *Binding) Resolve(sym sx.Symbol) (sx.Object, bool) {
	for curr := b; curr != nil; curr = curr.parent {
		if obj, found := curr.Lookup(sym); found {
			return obj, true
		}
	}
	return nil, false
}

// isConstBinding returns true if the symbol is defined with a constant
// binding in the given binding or its parent bindings.
func (b *Binding) isConstantBind(sym sx.Symbol) bool {
	for curr := b; curr != nil; curr = curr.parent {
		if curr.IsConst(sym) {
			return true
		}
		if _, found := curr.Lookup(sym); found {
			return false
		}
	}
	return false
}

// GetBinding returns the object as a binding, if possible.
func GetBinding(obj sx.Object) (*Binding, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	b, ok := obj.(*Binding)
	return b, ok
}
