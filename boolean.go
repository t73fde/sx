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

// MakeBoolean creates a new Boolean object.
func MakeBoolean(b bool) Object {
	if b {
		return T
	}
	return Nil()
}

// T is the default true object.
var T = MakeSymbol("T")

func init() {
	_ = T.Bind(T)
	T.Freeze()
}

// IsTrue returns true, if object is a true value.
//
// Everything except a nil object and the empty string, is a true value.
func IsTrue(obj Object) bool {
	if IsNil(obj) {
		return false
	}
	if s, ok := GetString(obj); ok && s.GetValue() == "" {
		return false
	}
	return !IsUndefined(obj)
}

// IsFalse returns true, if the object is a false value.
func IsFalse(obj Object) bool {
	if IsNil(obj) {
		return true
	}
	if s, ok := GetString(obj); ok && s.GetValue() == "" {
		return true
	}
	return IsUndefined(obj)
}
