//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package sx provides the basic objects to work with symbolic expressions.
package sx

import (
	"fmt"
	"io"
	"strings"
)

// Object is the generic value all s-expressions must fulfill.
type Object interface {
	fmt.Stringer

	// IsNil checks if the concrete object is nil.
	IsNil() bool

	// IsAtom returns true iff the object is an object that is not further decomposable.
	IsAtom() bool

	// IsEqual compare two objects for deep equality.
	IsEqual(Object) bool

	// Repr returns the object representation.
	Repr() string
}

// IsNil returns true, if the given object is the nil object.
func IsNil(obj Object) bool { return obj == nil || obj.IsNil() }

// Printable is a object that has is specific representation, which is different to String().
type Printable interface {
	// Print emits the string representation on the given Writer
	Print(io.Writer) (int, error)
}

// Print writes the string representation to a io.Writer.
func Print(w io.Writer, obj Object) (int, error) {
	if pr, ok := obj.(Printable); ok {
		return pr.Print(w)
	}
	if IsNil(obj) {
		return Nil().Print(w)
	}
	return io.WriteString(w, obj.Repr())
}

// WriteStrings is a helper function to write multiple strings at once.
func WriteStrings(w io.Writer, sl ...string) (int, error) {
	length := 0
	for _, s := range sl {
		l, err := io.WriteString(w, s)
		length += l
		if err != nil {
			return length, err
		}
	}
	return length, nil
}

// Repr returns the string representation of the given object.
func Repr(obj Object) string {
	if IsNil(obj) {
		return "()"
	}
	if probj, ok := obj.(Printable); ok {
		var sb strings.Builder
		if _, err := probj.Print(&sb); err != nil {
			return err.Error()
		}
		return sb.String()
	}
	return obj.Repr()
}
