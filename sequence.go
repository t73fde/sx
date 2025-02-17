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

import "iter"

// Sequence is an Object that has a finite, ordered set of elements.
type Sequence interface {
	Object

	// Length returns the length of the sequence.
	Length() int

	// LengthLess reports whether the length of the sequence is less than
	// the given number.
	LengthLess(n int) bool

	// LengthGreater reports whether the length of the sequence is greater than
	// the given number.
	LengthGreater(n int) bool

	// LengthEqual reports whether the length of the sequence is equal to the
	// given number.
	LengthEqual(n int) bool

	// Nth returns the n-th element of a sequence.
	Nth(n int) (Object, error)

	// MakeList returns the sequence as a pair list. If sequence is already a
	// pair list, it is returned without copying it.
	MakeList() *Pair

	// Values() returns an iterator over all elements of the sequence in
	// natural order.
	Values() iter.Seq[Object]
}

// --- Sequence functions

// GetSequence returns the object as a Sequence, if possible.
func GetSequence(obj Object) (Sequence, bool) {
	if IsNil(obj) {
		return nil, true
	}
	v, ok := obj.(Sequence)
	return v, ok
}
