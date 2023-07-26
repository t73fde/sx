//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf

import (
	"errors"
	"strconv"
)

// Number value store numbers.
type Number interface {
	Object

	// Returns true iff the number is equal to zero.
	IsZero() bool
}

// ParseInteger parses the string as an integer value and returns its value as a number.
func ParseInteger(s string) (Number, error) {
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return Int64(i64), err
}

// Int64 is a number that store 64 bit integer values.
type Int64 int64

func (i Int64) IsZero() bool { return i == 0 }

// IsNil return true, if it is a nil integer value.
func (i Int64) IsNil() bool { return false }

func (Int64) IsAtom() bool { return true }

// IsEql compare two objects.
func (i Int64) IsEql(other Object) bool {
	if otherI, ok := other.(Int64); ok {
		return i == otherI
	}
	return false
}

// IsEqual is currently the same as IsEqv, since we support only integer values.
func (i Int64) IsEqual(other Object) bool { return i.IsEql(other) }

// String returns the Go string representation.
func (i Int64) String() string { return strconv.FormatInt(int64(i), 10) }

// Repr returns the value representation.
func (i Int64) Repr() string { return i.String() }

// GetNumber returns the object as a number, if possible.
func GetNumber(obj Object) (Number, bool) {
	if IsNil(obj) {
		return nil, false
	}
	num, ok := obj.(Number)
	return num, ok
}

// NumCmp compares the two number and returns -1 if x < y, 0 if x = y, and 1 if x > y.
func NumCmp(x, y Number) int {
	if xi, yi := int64(x.(Int64)), int64(y.(Int64)); xi < yi {
		return -1
	} else if xi == yi {
		return 0
	}
	return 1
}

// NumNeg negates the given number.
func NumNeg(x Number) Number {
	return Int64(-int64(x.(Int64)))
}

// NumAdd adds the two numbers.
func NumAdd(x, y Number) Number {
	return Int64(int64(x.(Int64)) + int64(y.(Int64)))
}

// NumSub subtracts the two numbers.
func NumSub(x, y Number) Number {
	return Int64(int64(x.(Int64)) - int64(y.(Int64)))
}

// NumMul multiplies the two numbers.
func NumMul(x, y Number) Number {
	return Int64(int64(x.(Int64)) * int64(y.(Int64)))
}

// ErrZeroNotAllowed
var ErrZeroNotAllowed = errors.New("number zero not allowed")

// NumDiv divides the first by the second number.
func NumDiv(x, y Number) (Number, error) {
	if y.IsZero() {
		return nil, ErrZeroNotAllowed
	}
	return Int64(int64(x.(Int64)) / int64(y.(Int64))), nil
}

// NumMod divides the first by the second number.
func NumMod(x, y Number) (Number, error) {
	if y.IsZero() {
		return nil, ErrZeroNotAllowed
	}
	return Int64(int64(x.(Int64)) % int64(y.(Int64))), nil
}
