//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf

import "io"

// Undefined is an object that signal a 'no value'.
type Undefined struct{}

// MakeUndefined creates an undefined objact.
func MakeUndefined() Undefined { return Undefined{} }

func (Undefined) IsNil() bool  { return false }
func (Undefined) IsAtom() bool { return false }
func (kw Undefined) IsEql(other Object) bool {
	_, ok := other.(Undefined)
	return ok
}
func (udef Undefined) IsEqual(other Object) bool { return udef.IsEql(other) }
func (udef Undefined) String() string            { return udef.Repr() }
func (udef Undefined) Repr() string              { return "#<undefined>" }
func (udef Undefined) Print(w io.Writer) (int, error) {
	return io.WriteString(w, udef.Repr())
}

// IsUndefined returns true iff the object is a undefined value
func IsUndefined(obj Object) bool {
	_, ok := obj.(Undefined)
	return ok
}
