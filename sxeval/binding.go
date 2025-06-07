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
	"slices"
	"strings"
	"unsafe"

	"t73f.de/r/sx"
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
		impl, oimpl := b.impl, ob.impl
		if impl.length() != oimpl.length() {
			return false
		}
		alist, oalist := b.impl.alist(), ob.impl.alist()
		for val := range alist.Values() {
			pair := val.(*sx.Pair)
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
	if !b.impl.bind(sym, obj) {
		lst := b.impl.alist()
		mb := makeMappedBinding(lst.Length() + 1)
		for val := range lst.Values() {
			pair := val.(*sx.Pair)
			mb.bind(pair.Car().(*sx.Symbol), pair.Cdr())
		}
		mb.bind(sym, obj)
		b.impl = mb
	}
	return nil
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
	for range n {
		b = b.parent
	}
	return b.Lookup(sym)
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

// Symbols returns all bound symbols, sorted by its GoString.
func (b *Binding) Symbols() []*sx.Symbol { return b.impl.symbols() }

// Bindings returns all bindings as an a-list in some random order.
func (b *Binding) Bindings() *sx.Pair { return b.impl.alist() }

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
		for range n {
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

// MakeNotBoundError builds an error to signal that a symbol was not bound in
// the environment.
func (b *Binding) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Binding: b, Sym: sym}
}

// NotBoundError signals that a symbol was not found in a binding.
type NotBoundError struct {
	Binding *Binding
	Sym     *sx.Symbol
}

func (e NotBoundError) Error() string {
	var sb strings.Builder
	if e.Sym == nil {
		sb.WriteString("symbol == nil, not bound in ")
	} else {
		fmt.Fprintf(&sb, "symbol %q not bound in ", e.Sym.String())
	}
	second := false
	for binding := e.Binding; binding != nil; binding = binding.Parent() {
		if second {
			sb.WriteString("->")
		}
		fmt.Fprintf(&sb, "%q", binding.Name())
		second = true
	}
	return sb.String()
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
	symbols() []*sx.Symbol
	alist() *sx.Pair
	length() int
}

type mapSymObj = map[*sx.Symbol]sx.Object

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
	mb.m[sym] = obj
	return true
}
func (mb mappedBinding) lookup(sym *sx.Symbol) (sx.Object, bool) {
	obj, found := mb.m[sym]
	return obj, found
}
func (mb mappedBinding) symbols() []*sx.Symbol {
	result := make([]*sx.Symbol, 0, len(mb.m))
	for sym := range mb.m {
		result = append(result, sym)
	}
	slices.SortFunc(result, func(symA, symB *sx.Symbol) int {
		facA, facB := symA.Package(), symB.Package()
		if facA == facB {
			return strings.Compare(symA.GetValue(), symB.GetValue())
		}

		// Make a stable descision, if symbols were created from different factories.
		if uintptr(unsafe.Pointer(facA)) < uintptr(unsafe.Pointer(facB)) {
			return -1
		}
		return 1
	})
	return result
}
func (mb mappedBinding) alist() *sx.Pair {
	var result sx.ListBuilder
	for sym, obj := range mb.m {
		result.Add(sx.Cons(sym, obj))
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
		if bsym != sym {
			return false
		}
	}
	sb.sym = sym
	sb.obj = obj
	return true
}
func (sb *singleBinding) lookup(sym *sx.Symbol) (sx.Object, bool) {
	if bsym := sb.sym; bsym != nil && bsym == sym {
		return sb.obj, true
	}
	return nil, false
}
func (sb *singleBinding) symbols() []*sx.Symbol {
	if bsym := sb.sym; bsym != nil {
		return []*sx.Symbol{bsym}
	}
	return nil
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
