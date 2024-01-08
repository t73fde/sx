//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

// Package sx provides the basic objects to work with symbolic expressions.
package sx

import (
	"fmt"
	"io"
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
	return io.WriteString(w, obj.String())
}
