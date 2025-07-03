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
	"maps"

	"t73f.de/r/sx"
)

// ----- Notes
//
// TODO: Remove Binding.Freeze, since it is the **dynamic** environment.
//
// -----

// Binding is a binding binds symbol to values. It is the dynamic environment
// of a call, in contrast to `Frame`. It can be frozen, so it cannot be
// updated any more.
type Binding struct {
	mso    mapSymObj
	name   string
	parent *Binding
	frozen bool
}

type mapSymObj = map[*sx.Symbol]sx.Object

func makeBinding(name string, parent *Binding, sizeHint int) *Binding {
	return &Binding{
		mso:    make(mapSymObj, sizeHint),
		parent: parent,
		name:   name,
	}
}

// MakeRootBinding creates a new root binding.
func MakeRootBinding(sizeHint int) *Binding { return makeBinding("root", nil, sizeHint) }

// MakeChildBinding creates a new binding with a given parent.
func (b *Binding) MakeChildBinding(name string, sizeHint int) *Binding {
	return makeBinding(name, b, sizeHint)
}

// IsNil returns true if the binding is the nil binding.
func (b *Binding) IsNil() bool { return b == nil }

// IsAtom returns true if the binding is an atom.
func (b *Binding) IsAtom() bool { return b == nil }

// IsEqual returns true if both objects have the same value.
func (b *Binding) IsEqual(other sx.Object) bool {
	if b == other {
		return true
	}
	if b.IsNil() {
		return sx.IsNil(other)
	}
	if ob, isBinding := other.(*Binding); isBinding {
		return maps.EqualFunc(b.mso, ob.mso, func(o1, o2 sx.Object) bool { return o1.IsEqual(o2) })
	}
	return false
}

func (b *Binding) String() string {
	return fmt.Sprintf("#<binding:%s/%d>", b.name, len(b.mso))
}

// GoString returns the binding as a string suitable to be used in Go code.
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
// A previous mapping will be overwritten.
func (b *Binding) Bind(sym *sx.Symbol, obj sx.Object) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	b.mso[sym] = obj
	return nil
}

// ErrBindingFrozen is returned when trying to update a frozen binding.
type ErrBindingFrozen struct{ Binding *Binding }

func (err ErrBindingFrozen) Error() string { return fmt.Sprintf("binding is frozen: %v", err.Binding) }

// Lookup will search for a local binding of the given symbol. If not
// found, the search will *not* be continued in the parent binding.
// Use `Resolve`, if you want a search up to the parent.
func (b *Binding) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	if sym != nil {
		obj, found := b.mso[sym]
		return obj, found
	}
	return sx.Nil(), false
}

// Resolve returns the object that is bound to a symbol. It searches in the
// binding and in all parent bindings.
func (b *Binding) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	if sym != nil {
		for curr := b; curr != nil; curr = curr.parent {
			if obj, found := curr.Lookup(sym); found {
				return obj, true
			}
		}
	}
	return sx.Nil(), false
}

// FindBinding returns the binding, where the symbol is bound to a value.
// If no binding was found, nil is returned.
func (b *Binding) FindBinding(sym *sx.Symbol) *Binding {
	for curr := b; curr != nil; curr = curr.parent {
		if _, found := curr.Lookup(sym); found {
			return curr
		}
	}
	return nil
}

// Bindings returns all bindings as an a-list in some random order.
func (b *Binding) Bindings() *sx.Pair {
	var result sx.ListBuilder
	for sym, obj := range b.mso {
		result.Add(sx.Cons(sym, obj))
	}
	return result.List()
}

// Freeze sets the binding in a read-only state.
func (b *Binding) Freeze() { b.frozen = true }

// IsFrozen returns true if binding is frozen.
func (b *Binding) IsFrozen() bool { return b.frozen }
