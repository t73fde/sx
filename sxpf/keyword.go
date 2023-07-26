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

// Keyword represents a symbolic value.
//
// A keyword is like a string, but contains only printable characters.
// A keyword is like a symbol, but does not allow to associate additional data.
// A keyword is often used as a key of an association list, as it is more lightweight
// compared to a string and to a symbol.
type Keyword string

// IsNil return true, if it is a nil keyword object.
func (Keyword) IsNil() bool { return false }

func (Keyword) IsAtom() bool { return true }

// IsEql compare two objects.
func (kw Keyword) IsEql(other Object) bool {
	otherKw, ok := other.(Keyword)
	return ok && string(kw) == string(otherKw)
}

// IsEqual is the same a IsEqv for keywords.
func (kw Keyword) IsEqual(other Object) bool { return kw.IsEql(other) }

// String returns the Go string representation.
func (kw Keyword) String() string { return string(kw) }

// Repr returns the value representation.
func (kw Keyword) Repr() string { return "&" + string(kw) }

// Print write the string representation to the given Writer.
func (kw Keyword) Print(w io.Writer) (int, error) {
	return io.WriteString(w, kw.Repr())
}

// GetKeyword returns the object as a keyword, if possible
func GetKeyword(obj Object) (Keyword, bool) {
	if IsNil(obj) {
		return Keyword(""), false
	}
	kw, ok := obj.(Keyword)
	return kw, ok
}
