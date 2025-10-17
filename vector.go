//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sx

import (
	"fmt"
	"io"
	"iter"
	"slices"
	"strings"
)

// Vector is a sequence of Objects.
type Vector []Object

// VectorName is the name of the (vector ...) builtin
const VectorName = "vector"

// IsNil returns true, if the vector should be treated like the Nil() object.
func (v Vector) IsNil() bool { return len(v) == 0 }

// IsAtom signals an atomic value. Only the empty vector is atomic.
func (v Vector) IsAtom() bool { return len(v) == 0 }

// IsTrue returns true if vector can be interpreted as a "true" value.
func (v Vector) IsTrue() bool { return len(v) > 0 }

// IsEqual compares the vector with another object to have the same content.
func (v Vector) IsEqual(other Object) bool {
	if v.IsNil() {
		return IsNil(other)
	}
	if otherVector, ok := other.(Vector); ok && len(v) == len(otherVector) {
		for i, obj := range v {
			if !obj.IsEqual(otherVector[i]) {
				break
			}
		}
		return true
	}
	return false
}

func (v Vector) String() string {
	var sb strings.Builder
	if _, err := v.Print(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

// GoString returns the string representation to be used in Go code.
func (v Vector) GoString() string { return v.String() }

// Print write the string representation to the given Writer.
func (v Vector) Print(w io.Writer) (int, error) {
	l, err := io.WriteString(w, "(")
	if err != nil {
		return l, err
	}
	var len2 int
	len2, err = io.WriteString(w, VectorName)
	l += len2
	if err != nil {
		return l, err
	}

	for _, obj := range v {
		len2, err = io.WriteString(w, " ")
		l += len2
		if err != nil {
			return l, err
		}
		len2, err = Print(w, obj)
		l += len2
		if err != nil {
			return l, err
		}
	}
	len2, err = io.WriteString(w, ")")
	l += len2
	return l, err
}

// --- Sequence methods

// Length returns the length of the vector.
func (v Vector) Length() int { return len(v) }

// LengthLess return true, if the length of the vector is less than the given
// value.
func (v Vector) LengthLess(n int) bool { return len(v) < n }

// LengthGreater return true, if the length of the vector is greater than the
// given value.
func (v Vector) LengthGreater(n int) bool { return len(v) > n }

// LengthEqual return true, if the length of the vector is equal to the given
// value.
func (v Vector) LengthEqual(n int) bool { return len(v) == n }

// Nth returns the object at the given position. It is an error if the position
// is less than zero or greater than the vector length.
func (v Vector) Nth(n int) (Object, error) {
	if n < 0 || len(v) <= n {
		return Nil(), fmt.Errorf("index out of range: %d (max: %d)", n, len(v)-1)
	}
	return v[n], nil
}

// MakeList builds a pair list from the vector.
func (v Vector) MakeList() *Pair { return MakeList(v...) }

// Values returns an iterator over the sequence of vector elements.
func (v Vector) Values() iter.Seq[Object] { return slices.Values(v) }

// --- Vector functions

// GetVector returns the object as a vector, if possible.
func GetVector(obj Object) (Vector, bool) {
	if IsNil(obj) {
		return nil, true
	}
	v, ok := obj.(Vector)
	return v, ok
}

// Collect values from seq into a new vector and return it.
func Collect(it iter.Seq[Object]) Vector { return slices.Collect(it) }
