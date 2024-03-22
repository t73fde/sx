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

// Binding is a binding based on maps.
type Binding struct {
	impl   bindingImpl
	parent *Binding
	inline singleBinding
	name   string
	frozen bool
}

func makeBinding(name string, parent *Binding, sizeHint int) *Binding {
	b := Binding{
		parent: parent,
		name:   name,
	}
	switch sizeHint {
	case 0, 1:
		b.impl = &b.inline
	default:
		b.impl = makeMappedBinding(sizeHint)
	}
	return &b
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
		impl, oimpl := b.impl, ob.impl
		if impl.length() != oimpl.length() {
			return false
		}
		alist, oalist := b.impl.alist(), ob.impl.alist()
		for node := alist; node != nil; node = node.Tail() {
			pair := node.Car().(*sx.Pair)
			opair := oalist.Assoc(pair.Car())
			if opair == nil || !pair.Cdr().IsEqual(opair.Cdr()) {
				return false
			}
		}
		return true
	}
	return false
}
func (b *Binding) String() string {
	return fmt.Sprintf("#<binding:%s/%d>", b.name, b.impl.length())
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
// A previous mapping will be overwritten.
func (b *Binding) Bind(sym *sx.Symbol, obj sx.Object) error {
	if b.frozen {
		return ErrBindingFrozen{Binding: b}
	}
	if !b.impl.bind(sym, obj) {
		lst := b.impl.alist()
		mb := makeMappedBinding(lst.Length() + 1)
		for node := lst; node != nil; node = node.Tail() {
			pair := node.Car().(*sx.Pair)
			mb.bind(pair.Car().(*sx.Symbol), pair.Cdr())
		}
		mb.bind(sym, obj)
		b.impl = mb
	}
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
	if sym != nil {
		return b.impl.lookup(sym)
	}
	return sx.Nil(), false
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
	return b.impl.alist()
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
	if sym != nil {
		for curr := b; curr != nil; curr = curr.parent {
			if obj, found := curr.Lookup(sym); found {
				return obj, curr, depth
			}
			depth++
		}
	}
	return sx.Nil(), nil, depth
}

// Resolve a symbol in a binding and all of its parent bindings.
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

// ResolveN resolves a symbol in the N-th parent binding and all of its parent
// bindings.
func (b *Binding) ResolveN(sym *sx.Symbol, n int) (sx.Object, bool) {
	if sym != nil {
		curr := b
		for i := 0; i < n; i++ {
			curr = curr.parent
		}
		for ; curr != nil; curr = curr.parent {
			if obj, found := curr.Lookup(sym); found {
				return obj, true
			}
		}
	}
	return sx.Nil(), false
}

// GetBinding returns the object as a binding, if possible.
func GetBinding(obj sx.Object) (*Binding, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	b, ok := obj.(*Binding)
	return b, ok
}

type bindingImpl interface {
	bind(*sx.Symbol, sx.Object) bool
	lookup(*sx.Symbol) (sx.Object, bool)
	alist() *sx.Pair
	length() int
}

type mapSymObj = map[string]sx.Object

type mappedBinding struct {
	m mapSymObj
}

func makeMappedBinding(sizeHint int) mappedBinding {
	if sizeHint < 3 {
		sizeHint = 3
	}
	return mappedBinding{make(mapSymObj, sizeHint)}
}
func (mb mappedBinding) bind(sym *sx.Symbol, obj sx.Object) bool {
	mb.m[sym.GetValue()] = obj
	return true
}
func (mb mappedBinding) lookup(sym *sx.Symbol) (sx.Object, bool) {
	obj, found := mb.m[sym.GetValue()]
	return obj, found
}
func (mb mappedBinding) alist() *sx.Pair {
	var result sx.ListBuilder
	for s, obj := range mb.m {
		result.Add(sx.Cons(sx.MakeSymbol(s), obj))
	}
	return result.List()
}
func (mb mappedBinding) length() int { return len(mb.m) }

type singleBinding struct {
	sym *sx.Symbol
	obj sx.Object
}

func (sb *singleBinding) bind(sym *sx.Symbol, obj sx.Object) bool {
	if bsym := sb.sym; bsym != nil {
		if !bsym.IsEqual(sym) {
			return false
		}
	}
	sb.sym = sym
	sb.obj = obj
	return true
}
func (sb *singleBinding) lookup(sym *sx.Symbol) (sx.Object, bool) {
	if bsym := sb.sym; bsym != nil && bsym.IsEqual(sym) {
		return sb.obj, true
	}
	return nil, false
}
func (sb *singleBinding) alist() *sx.Pair {
	if bsym := sb.sym; bsym != nil {
		return sx.Cons(sx.Cons(bsym, sb.obj), sx.Nil())
	}
	return nil
}
func (sb *singleBinding) length() int {
	if sb.sym == nil {
		return 0
	}
	return 1
}
