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

func (Undefined) IsNil() bool                    { return false }
func (Undefined) IsAtom() bool                   { return false }
func (udef Undefined) IsEqual(other Object) bool { return IsUndefined(other) }
func (udef Undefined) String() string            { return "#<undefined>" }

// IsUndefined returns true iff the object is a undefined value
func IsUndefined(obj Object) bool {
	_, ok := obj.(Undefined)
	return ok
}
