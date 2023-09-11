//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sx

// MakeBoolean creates a new Boolean object.
func MakeBoolean(b bool) Object {
	if b {
		return Int64(1)
	}
	return Nil()
}

// IsTrue returns true, if object is a true value.
//
// Everything except a nil object, the False object, and the empty string, is a true value.
func IsTrue(obj Object) bool {
	if IsNil(obj) {
		return false
	}
	if s, ok := GetString(obj); ok && s.String() == "" {
		return false
	}
	return true
}

// IsFalse returns true, if object is a false value.
//
// A nil object, the False object or an empty string are false values.
func IsFalse(obj Object) bool { return !IsTrue(obj) }
