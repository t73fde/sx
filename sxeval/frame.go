//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval

import (
	"fmt"
	"maps"

	"t73f.de/r/sx"
)

// Frame is a binding specific for a call.
type Frame struct {
	mso    mapSymObj  // used if more than one symbol is bound
	sym    *sx.Symbol // used if zero or one symbol is bound
	obj    sx.Object  // object bound to sym
	name   string
	parent *Frame
}

func makeFrame(name string, parent *Frame, sizeHint int) *Frame {
	f := Frame{
		parent: parent,
		name:   name,
	}
	if sizeHint > 1 {
		f.mso = make(mapSymObj, sizeHint)
	}
	return &f
}

// MakeChildFrame creates a new frame with a given parent.
func (f *Frame) MakeChildFrame(name string, sizeHint int) *Frame {
	return makeFrame(name, f, sizeHint)
}

// IsNil returns true if the frame is the nil frame.
func (f *Frame) IsNil() bool { return f == nil }

// IsAtom returns true if the frame is an atom.
func (f *Frame) IsAtom() bool { return f == nil }

// IsEqual returns true if both objects have the same value.
func (f *Frame) IsEqual(other sx.Object) bool {
	if f == other {
		return true
	}
	if f.IsNil() {
		return sx.IsNil(other)
	}
	if of, isFrame := other.(*Frame); isFrame {
		fmso, ofmso := f.mso, of.mso
		if fmso != nil && ofmso != nil {
			return maps.EqualFunc(fmso, ofmso, func(o1, o2 sx.Object) bool { return o1.IsEqual(o2) })
		}
		fsym, ofsym := f.sym, of.sym
		if fmso == nil && ofmso == nil {
			if fsym == nil {
				return ofsym == nil
			}
			if ofsym == nil {
				return false
			}
			return f.obj.IsEqual(of.obj)
		}

		if fsym != nil {
			if len(ofmso) != 1 {
				return false
			}
			obj, found := ofmso[fsym]
			return found && obj.IsEqual(f.obj)
		}
		if ofsym != nil {
			if len(fmso) != 1 {
				return false
			}
			obj, found := fmso[ofsym]
			return found && obj.IsEqual(of.obj)
		}
		return len(fmso) == 0 && len(ofmso) == 0
	}
	return false
}

func (f *Frame) String() string {
	return fmt.Sprintf("#<frame:%s/%d>", f.name, f.length())
}

// GoString returns the frame as a string suitable to be used in Go code.
func (f *Frame) GoString() string { return f.String() }

// Name returns the local name of this frame.
func (f *Frame) Name() string {
	if f == nil {
		return "<nil>"
	}
	return f.name
}

// Parent returns the parent frame.
func (f *Frame) Parent() *Frame {
	if f == nil {
		return nil
	}
	return f.parent
}

func (f *Frame) length() int {
	if m := f.mso; m != nil {
		return len(m)
	}
	if f.sym == nil {
		return 0
	}
	return 1
}

// Bind creates a local mapping with a given symbol and object.
//
// A previous mapping will be overwritten.
func (f *Frame) Bind(sym *sx.Symbol, obj sx.Object) {
	if m := f.mso; m != nil {
		m[sym] = obj
	} else if f.sym == nil {
		f.sym = sym
		f.obj = obj
	} else if f.sym == sym {
		f.obj = obj
	} else {
		f.mso = make(mapSymObj, 2)
		f.mso[f.sym] = f.obj
		f.sym = nil
		f.mso[sym] = obj
	}
}

// Lookup will search for a local binding of the given symbol. If not
// found, the search will *not* be continued in the parent frame.
// Use the global `Resolve` function, if you want a search up to the parent.
func (f *Frame) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	if sym == nil || f == nil {
		return sx.Nil(), false
	}
	if m := f.mso; m != nil {
		obj, found := m[sym]
		return obj, found
	}
	if f.sym == sym {
		return f.obj, true
	}
	return sx.Nil(), false
}

// lookupN will lookup the symbol in the N-th parent.
func (f *Frame) lookupN(sym *sx.Symbol, n int) (sx.Object, bool) {
	for range n {
		f = f.parent
	}
	return f.Lookup(sym)
}

// FindFrame returns the frame, where the symbol is bound to a value.
// If no binding was found, nil is returned.
func (f *Frame) FindFrame(sym *sx.Symbol) *Frame {
	for curr := f; curr != nil; curr = curr.parent {
		if _, found := curr.Lookup(sym); found {
			return curr
		}
	}
	return nil
}

// Bindings returns all bindings as an a-list in some random order.
func (f *Frame) Bindings() *sx.Pair {
	if f == nil {
		return sx.Nil()
	}
	if m := f.mso; m != nil {
		var result sx.ListBuilder
		for sym, obj := range m {
			result.Add(sx.Cons(sym, obj))
		}
		return result.List()
	}
	if bsym := f.sym; bsym != nil {
		return sx.Cons(sx.Cons(bsym, f.obj), sx.Nil())
	}
	return nil
}

// GetFrame returns the object as a frame, if possible.
func GetFrame(obj sx.Object) (*Frame, bool) {
	f, ok := obj.(*Frame)
	return f, ok
}
