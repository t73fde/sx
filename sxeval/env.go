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

// Binding maintains a mapping between symbols and values.
type Binding interface {
	// A binding is an object by itself
	sx.Object

	// String returns the local name of this binding.
	String() string

	// Parent allows to retrieve the parent binding. If the binding is the root
	// binding, nil is returned. Lookups that cannot be satisfied in an
	// binding are often delegated to the parent binding.
	Parent() Binding

	// IsRoot returns true for the root binding.
	IsRoot() bool

	// Bind creates a local mapping with a given symbol and object.
	//
	// A previous, non-const mapping will be overwritten.
	Bind(*sx.Symbol, sx.Object) error

	// BindConst creates a local mapping of the symbol to the object, which
	// cannot be changed afterwards.
	BindConst(*sx.Symbol, sx.Object) error

	// Lookup will search for a local binding of the given symbol. If not
	// found, the search will *not* be continued in the parent binding.
	// Use the global `Resolve` function, if you want a search up to the parent.
	Lookup(*sx.Symbol) (sx.Object, bool)

	// IsConst returns true if the binding of the symbol is a constant binding.
	IsConst(*sx.Symbol) bool

	// Bindings returns all bindings as an a-list in some random order.
	Bindings() *sx.Pair

	// Unbind removes the mapping of the given symbol to an object.
	Unbind(*sx.Symbol) error

	// Freeze sets the binding in a read-only state.
	Freeze()
}

// ErrBindingFrozen is returned when trying to update a frozen binding.
type ErrBindingFrozen struct{ Binding Binding }

func (err ErrBindingFrozen) Error() string { return fmt.Sprintf("binding is frozen: %v", err.Binding) }

// ErrConstBinding is returned when a constant binding should be changed.
type ErrConstBinding struct{ Sym *sx.Symbol }

func (err ErrConstBinding) Error() string {
	return fmt.Sprintf("constant bindung for symbol %v", err.Sym.Repr())
}

type mapSymObj = map[*sx.Symbol]sx.Object

// MakeRootBinding creates a new root binding.
func MakeRootBinding(sizeHint int) Binding {
	return &mappedBinding{
		name:   "root",
		parent: nil,
		vars:   make(mapSymObj, sizeHint),
		isRoot: true,
		frozen: false,
	}
}

// MakeChildBinding creates a new binding with a given parent.
func MakeChildBinding(parent Binding, name string, sizeHint int) Binding {
	if sizeHint <= 0 {
		sizeHint = 3
	}
	return &mappedBinding{
		name:   name,
		parent: parent,
		vars:   make(mapSymObj, sizeHint),
		isRoot: false,
		frozen: false,
	}
}

type mappedBinding struct {
	name   string
	parent Binding
	vars   mapSymObj
	consts map[*sx.Symbol]struct{}
	isRoot bool
	frozen bool
}

func (mb *mappedBinding) IsNil() bool  { return mb == nil }
func (mb *mappedBinding) IsAtom() bool { return mb == nil }
func (mb *mappedBinding) IsEqual(other sx.Object) bool {
	if mb == other {
		return true
	}
	if mb.IsNil() {
		return sx.IsNil(other)
	}
	if omb, ok := other.(*mappedBinding); ok {
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
func (mb *mappedBinding) Repr() string { return sx.Repr(mb) }
func (mb *mappedBinding) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<binding:", mb.name, "/", strconv.Itoa(len(mb.vars)), ">")
}
func (mb *mappedBinding) String() string { return mb.name }
func (mb *mappedBinding) Parent() Binding {
	if mb == nil {
		return nil
	}
	return mb.parent
}
func (mb *mappedBinding) IsRoot() bool { return mb == nil || mb.isRoot }
func (mb *mappedBinding) Bind(sym *sx.Symbol, val sx.Object) error {
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
func (mb *mappedBinding) BindConst(sym *sx.Symbol, val sx.Object) error {
	if mb.frozen {
		return ErrBindingFrozen{Binding: mb}
	}
	if mb.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if mb.consts == nil {
		mb.consts = map[*sx.Symbol]struct{}{sym: {}}
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
func (mb *mappedBinding) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	obj, found := mb.vars[sym]
	return obj, found
}
func (mb *mappedBinding) IsConst(sym *sx.Symbol) bool {
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
func (mb *mappedBinding) Bindings() *sx.Pair {
	result := sx.Nil()
	for k, v := range mb.vars {
		result = result.Cons(sx.Cons(k, v))
	}
	return result
}
func (mb *mappedBinding) Unbind(sym *sx.Symbol) error {
	if mb.frozen {
		return ErrBindingFrozen{Binding: mb}
	}
	delete(mb.vars, sym)
	return nil
}
func (mb *mappedBinding) Freeze() { mb.frozen = true }

// GetBinding returns the object as a binding, if possible.
func GetBinding(obj sx.Object) (Binding, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	bind, ok := obj.(Binding)
	return bind, ok
}

// RootBinding returns the root binding of the given binding.
func RootBinding(bind Binding) Binding {
	currBind := bind
	for {
		if currBind.IsRoot() {
			return currBind
		}
		currBind = currBind.Parent()
	}
}

// Resolve a symbol is a binding and all of its parent bindings.
func Resolve(bind Binding, sym *sx.Symbol) (sx.Object, bool) {
	currBind := bind
	for {
		obj, found := currBind.Lookup(sym)
		if found {
			return obj, true
		}
		if currBind.IsRoot() {
			return sx.Nil(), false
		}
		currBind = currBind.Parent()
	}
}

// IsConstBinding returns true if the symbol is defined with a constant
// binding in the given binding or its parent bindings.
func IsConstantBind(bind Binding, sym *sx.Symbol) bool {
	currBind := bind
	for !sx.IsNil(currBind) {
		if currBind.IsConst(sym) {
			return true
		}
		if _, found := currBind.Lookup(sym); found {
			return false
		}
		currBind = currBind.Parent()
	}
	return false
}

// AllBindings returns an a-list of all bindings in the given binding and its parent bindinga.
func AllBindings(bind Binding) *sx.Pair {
	currBind := bind
	result := currBind.Bindings()
	currResult := result
	if currResult != nil {
		for currResult.Tail() != nil {
			currResult = currResult.Tail()
		}
	}
	for {
		if currBind.IsRoot() {
			return result
		}
		currBind = currBind.Parent()
		if currBind == nil {
			return result
		}
		res := currBind.Bindings()
		if result == nil {
			result = res
			currResult = result
			if currResult != nil {
				for currResult.Tail() != nil {
					currResult = currResult.Tail()
				}
			}
		} else {
			currResult = currResult.ExtendBang(res)
		}
	}
}
