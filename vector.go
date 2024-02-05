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
	"strings"
)

// Vector is a sequence of Objects.
type Vector []Object

// VectorName is the name of the (vector ...) builtin
const VectorName = "vector"

func (v Vector) IsNil() bool  { return len(v) == 0 }
func (v Vector) IsAtom() bool { return len(v) == 0 }
func (v Vector) IsEqual(other Object) bool {
	if Object(v) == other {
		return true
	}
	if len(v) == 0 {
		return other.IsNil()
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
	v.Print(&sb)
	return sb.String()
}

// Print write the string representation to the given Writer.
func (v Vector) Print(w io.Writer) (int, error) {
	if len(v) == 0 {
		return io.WriteString(w, "()")
	}
	len, err := io.WriteString(w, "(")
	if err != nil {
		return len, err
	}
	var len2 int
	len2, err = io.WriteString(w, VectorName)
	len += len2
	if err != nil {
		return len, err
	}

	for _, obj := range v {
		len2, err = io.WriteString(w, " ")
		len += len2
		if err != nil {
			return len, err
		}
		len2, err = Print(w, obj)
		len += len2
		if err != nil {
			return len, err
		}
	}
	len2, err = io.WriteString(w, ")")
	len += len2
	return len, err
}

// --- Sequence methods

func (v Vector) Length() int { return len(v) }

func (v Vector) LengthLess(l int) bool { return len(v) < l }

func (v Vector) Nth(n int) (Object, error) {
	if n < 0 || len(v) <= n {
		return Nil(), fmt.Errorf("index out of range: %d (max: %d)", n, len(v)-1)
	}
	return v[n], nil
}

func (v Vector) MakeList() *Pair { return MakeList(v...) }

func (v Vector) Iterator() SequenceIterator {
	return &vectorIterator{v, 0}
}

// --- vectorIterator implements SequenceIterator
type vectorIterator struct {
	vec Vector
	pos int
}

func (vi *vectorIterator) HasElement() bool { return vi.pos < len(vi.vec) }
func (vi *vectorIterator) Element() Object {
	if pos, vec := vi.pos, vi.vec; pos < len(vec) {
		return vec[pos]
	}
	return MakeUndefined()
}
func (vi *vectorIterator) Advance() bool {
	vi.pos++
	return vi.pos < len(vi.vec)
}

// --- Vector functions

// GetVector returns the object as a vector, if possible.
func GetVector(obj Object) (Vector, bool) {
	if IsNil(obj) {
		return nil, true
	}
	v, ok := obj.(Vector)
	return v, ok
}
