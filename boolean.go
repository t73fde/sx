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

// Boolean represents the two boolean objects, `False` and `True`.
type Boolean bool

// The two boolean values.
//
// Do not use other constants.
// There are defined string values other code must respect (e.g. symbol factory, reader, ...)
const (
	True  = Boolean(true)
	False = Boolean(false)

	TrueString  = "True"
	FalseString = "False"
)

// MakeBoolean creates a new Boolean object.
func MakeBoolean(b bool) Boolean {
	if b {
		return True
	}
	return False
}

// IsNil return true, if it is a nil boolean value.
// Since there is no such value, it always returns false.
func (Boolean) IsNil() bool { return false }

// IsAtom always returns true, because all booleans are atoms.
func (Boolean) IsAtom() bool { return true }

// IsEql compares the boolean with an object for equivalence.
func (b Boolean) IsEql(other Object) bool {
	otherB, ok := other.(Boolean)
	return ok && b == otherB

}

// IsEqual compares the boolean with an object for equivalence.
func (b Boolean) IsEqual(other Object) bool {
	if b {
		return IsTrue(other)
	}
	return IsFalse(other)
}

// String returns the Go string representation.
func (b Boolean) String() string {
	if b == True {
		return "true"
	}
	return "false"
}

// Repr returns the value representation.
func (b Boolean) Repr() string {
	if b == True {
		return TrueString
	}
	return FalseString
}

// Negate returns the other boolean value.
func (b Boolean) Negate() Boolean {
	if b {
		return False
	}
	return True
}

// GetBoolean returns the object as a boolean, if possible.
func GetBoolean(obj Object) (Boolean, bool) {
	if IsNil(obj) {
		return False, false
	}
	b, ok := obj.(Boolean)
	return b, ok
}

// IsTrue returns true, if object is a true value.
//
// Everything except a nil object, the False object, and the empty string, is a true value.
func IsTrue(obj Object) bool {
	if IsNil(obj) || obj == False {
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
