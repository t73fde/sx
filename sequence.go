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

// Sequence is an Object that has a finite, ordered set of elements.
type Sequence interface {
	Object

	// Length returns the length of the sequence.
	Length() int

	// LengthLess returns true if the legth is less than the given one.
	LengthLess(l int) bool

	// Nth returns the n-th element of a sequence.
	Nth(n int) (Object, error)

	// MakeList returns the sequence as a pair list. If sequence is already a
	// pair list, it is returned without copying it.
	MakeList() *Pair

	// Iterator return an SequenceIterator to iterate over all elements of
	// the sequence in natural order.
	Iterator() SequenceIterator
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

// SequenceIterator allows to iterate over all elements of a Sequence.
type SequenceIterator interface {
	// HasElement returns true if the iterator is able to return an element.
	HasElement() bool

	// Element returns the current element, or Undefined{} if there is no such element.
	Element() Object

	// Advance moves to the next element and return true if there is one.
	Advance() bool
}
