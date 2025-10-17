//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sx

// Undefined is an object that signal a 'no value'.
type Undefined struct{}

// MakeUndefined creates an undefined objact.
func MakeUndefined() Undefined { return Undefined{} }

// IsNil always returns false because an undefined value is never nil.
func (Undefined) IsNil() bool { return false }

// IsAtom always returns false because an undefined value is never atomic.
func (Undefined) IsAtom() bool { return false }

// IsTrue returns true if undefined can be interpreted as a "true" value.
// Hint: it will never ;)
func (Undefined) IsTrue() bool { return false }

// IsEqual returns true if the other value has the same content.
func (Undefined) IsEqual(other Object) bool { return IsUndefined(other) }

// String returns a strinf representation.
func (Undefined) String() string { return "#<undefined>" }

// GoString returns a string representation to be used in Go code.
func (udef Undefined) GoString() string { return udef.String() }

// IsUndefined returns true iff the object is a undefined value
func IsUndefined(obj Object) bool {
	_, ok := obj.(Undefined)
	return ok
}
