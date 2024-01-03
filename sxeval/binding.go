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
	"io"
	"strconv"

	"zettelstore.de/sx.fossil"
)

// ErrBindingFrozen is returned when trying to update a frozen binding.
type ErrBindingFrozen struct{ Binding *Binding }

func (err ErrBindingFrozen) Error() string { return fmt.Sprintf("binding is frozen: %v", err.Binding) }

// ErrConstBinding is returned when a constant binding should be changed.
type ErrConstBinding struct{ Sym sx.Symbol }

func (err ErrConstBinding) Error() string {
	return fmt.Sprintf("constant bindung for symbol %v", err.Sym.Repr())
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

func (mb *Binding) IsNil() bool  { return mb == nil }
func (mb *Binding) IsAtom() bool { return mb == nil }
func (mb *Binding) IsEqual(other sx.Object) bool {
	if mb == other {
		return true
	}
	if mb.IsNil() {
		return sx.IsNil(other)
	}
	if omb, ok := other.(*Binding); ok {
		mvars, ovars := mb.vars, omb.vars
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
func (mb *Binding) Repr() string { return sx.Repr(mb) }
func (mb *Binding) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<binding:", mb.name, "/", strconv.Itoa(len(mb.vars)), ">")
}

// String returns the local name of this binding.
func (mb *Binding) String() string { return mb.name }

// Parent returns the parent binding.
func (mb *Binding) Parent() *Binding {
	if mb == nil {
		return nil
	}
	return mb.parent
}

// Bind creates a local mapping with a given symbol and object.
//
// A previous, non-const mapping will be overwritten.
func (mb *Binding) Bind(sym sx.Symbol, val sx.Object) error {
	if mb.frozen {
		return ErrBindingFrozen{Binding: mb}
	}
	if mb.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if _, found := mb.vars[sym]; found {
		mb.vars[sym] = val
		return nil
	}
	mb.vars[sym] = val
	return nil
}

// BindConst creates a local mapping of the symbol to the object, which
// cannot be changed afterwards.
func (mb *Binding) BindConst(sym sx.Symbol, val sx.Object) error {
	if mb.frozen {
		return ErrBindingFrozen{Binding: mb}
	}
	if mb.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if mb.consts == nil {
		mb.consts = map[sx.Symbol]struct{}{sym: {}}
	} else {
		mb.consts[sym] = struct{}{}
	}
	if _, found := mb.vars[sym]; found {
		mb.vars[sym] = val
		return nil
	}
	mb.vars[sym] = val
	return nil
}

func (mb *Binding) BindSpecial(syn *Special) error {
	return mb.BindConst(sx.Symbol(syn.Name), syn)
}

// BindBuiltin binds the given builtin with its given name in the engine's
// root binding.
func (mb *Binding) BindBuiltin(b *Builtin) error {
	return mb.BindConst(sx.Symbol(b.Name), b)
}

// Lookup will search for a local binding of the given symbol. If not
// found, the search will *not* be continued in the parent binding.
// Use the global `Resolve` function, if you want a search up to the parent.
func (mb *Binding) Lookup(sym sx.Symbol) (sx.Object, bool) {
	obj, found := mb.vars[sym]
	return obj, found
}

// IsConst returns true if the binding of the symbol is a constant binding.
func (mb *Binding) IsConst(sym sx.Symbol) bool {
	if mb == nil {
		return false
	}
	if mb.frozen {
		if _, found := mb.vars[sym]; found {
			return true
		}
	}
	if mb.consts == nil {
		return false
	}
	_, found := mb.consts[sym]
	return found
}

// Bindings returns all bindings as an a-list in some random order.
func (mb *Binding) Bindings() *sx.Pair {
	result := sx.Nil()
	for k, v := range mb.vars {
		result = result.Cons(sx.Cons(k, v))
	}
	return result
}

// Unbind removes the mapping of the given symbol to an object.
func (mb *Binding) Unbind(sym sx.Symbol) error {
	if mb.frozen {
		return ErrBindingFrozen{Binding: mb}
	}
	delete(mb.vars, sym)
	return nil
}

// Freeze sets the binding in a read-only state.
func (mb *Binding) Freeze() { mb.frozen = true }

// Resolve a symbol in a binding and all of its parent bindings.
func (mb *Binding) Resolve(sym sx.Symbol) (sx.Object, bool) {
	for curr := mb; curr != nil; curr = curr.parent {
		if obj, found := curr.Lookup(sym); found {
			return obj, true
		}
	}
	return nil, false
}

// isConstBinding returns true if the symbol is defined with a constant
// binding in the given binding or its parent bindings.
func (mb *Binding) isConstantBind(sym sx.Symbol) bool {
	for curr := mb; curr != nil; curr = curr.parent {
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
	env, ok := obj.(*Binding)
	return env, ok
}
