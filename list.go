//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sx

import (
	"fmt"
	"io"
	"strings"
)

// Pair is a node containing a value for the element and a pointer to the tail.
// In other lisps it is often called "cell", "cons", "cons-cell", or "list".
type Pair struct {
	car Object
	cdr Object
}

// Nil() returns the nil list.
func Nil() *Pair { return (*Pair)(nil) }

// Cons prepends a value in front of a given listreturning the new list.
func (pair *Pair) Cons(car Object) *Pair { return &Pair{car: car, cdr: pair} }

func Cons(car, cdr Object) *Pair { return &Pair{car: car, cdr: cdr} }

// MakeList creates a new list with the given objects.
func MakeList(objs ...Object) *Pair {
	if len(objs) == 0 {
		return Nil()
	}
	result := Nil()
	for i := len(objs) - 1; i >= 0; i-- {
		result = result.Cons(objs[i])
	}
	return result
}

// IsNil return true, if it is a nil pair object.
func (pair *Pair) IsNil() bool { return pair == nil }

func (pair *Pair) IsAtom() bool { return pair == nil }

// IsEqual compare two objects.
func (pair *Pair) IsEqual(other Object) bool {
	if pair == other {
		return true
	}
	if pair == nil {
		return other.IsNil()
	}
	if otherPair, ok := other.(*Pair); ok {
		node, otherNode := pair, otherPair
		for ; node != nil && otherNode != nil; node = node.Tail() {
			if !node.Car().IsEqual(otherNode.Car()) {
				return false
			}
			otherNode = otherNode.Tail()
		}
		return node == otherNode
	}
	return false
}

// String returns the string representation.
func (pair *Pair) String() string {
	var sb strings.Builder
	pair.Print(&sb)
	return sb.String()
}

// Print write the string representation to the given Writer.
func (pair *Pair) Print(w io.Writer) (int, error) {
	if pair == nil {
		return io.WriteString(w, "()")
	}
	len, err := io.WriteString(w, "(")
	if err != nil {
		return len, err
	}
	var len2 int

	for node := pair; ; {
		if node != pair {
			len2, err = io.WriteString(w, " ")
			len += len2
			if err != nil {
				return len, err
			}
		}
		len2, err = Print(w, node.car)
		len += len2
		if err != nil {
			return len, err
		}

		cdr := node.cdr
		if IsNil(cdr) {
			break
		}
		if n, ok := cdr.(*Pair); ok {
			node = n
			continue
		}

		len2, err = io.WriteString(w, " . ")
		len += len2
		if err != nil {
			return len, err
		}
		len2, err = Print(w, cdr)
		len += len2
		if err != nil {
			return len, err
		}
		break
	}

	len2, err = io.WriteString(w, ")")
	len += len2
	return len, err
}

// GetPair returns the object as a pair.
func GetPair(obj Object) (*Pair, bool) {
	if IsNil(obj) {
		return nil, true
	}
	if lst, ok := obj.(*Pair); ok {
		return lst, true
	}
	return nil, false
}

// IsList returns true, if the object is a list, not just a pair.
// A list must have a nil value at the last cdr.
func IsList(obj Object) bool {
	pair, isPair := GetPair(obj)
	if !isPair {
		return false
	}
	if pair == nil {
		return true
	}
	for node := pair; ; {
		next, isPair2 := GetPair(node.cdr)
		if !isPair2 {
			return false
		}
		if next == nil {
			return true
		}
		node = next
	}
}

// Car returns the first object of a pair.
func (pair *Pair) Car() Object {
	if pair == nil {
		return Nil()
	}
	return pair.car
}

// Cdr returns the second object of a pair.
func (pair *Pair) Cdr() Object {
	if pair == nil {
		return Nil()
	}
	return pair.cdr
}

// SetCdr sets the cdr of the pair to the given object.
func (pair *Pair) SetCdr(obj Object) {
	if pair != nil {
		pair.cdr = obj
	}
}

// Last returns the last element of a non-empty list.
func (pair *Pair) Last() (Object, error) {
	if pair == nil {
		return nil, ErrImproper{Pair: pair}
	}
	for node := pair; ; {
		next, isPair := GetPair(node.cdr)
		if !isPair {
			return nil, ErrImproper{Pair: pair}
		}
		if next == nil {
			return node.car, nil
		}
		node = next
	}
}

// LastPair returns the last pair of the given pair list, or nil.
func (pair *Pair) LastPair() *Pair {
	if pair == nil {
		return nil
	}
	elem := pair
	for {
		rest := elem.Cdr()
		if IsNil(rest) {
			return elem
		}
		next, ok := rest.(*Pair)
		if !ok {
			return elem
		}
		elem = next
	}
}

// Head returns the first object as a pair, if possible.
// Otherwise it returns nil.
func (pair *Pair) Head() *Pair {
	if pair != nil {
		if head, ok := pair.car.(*Pair); ok {
			return head
		}
	}
	return nil
}

// Tail returns the second object as a pair, if possible.
// Otherwise it returns nil.
func (pair *Pair) Tail() *Pair {
	if pair != nil {
		if tail, ok := pair.cdr.(*Pair); ok {
			return tail
		}
	}
	return nil
}

// Length returns the length of the pair list.
func (pair *Pair) Length() int {
	result := 0
	for n := pair; n != nil; n = n.Tail() {
		result++
	}
	return result
}

// Assoc returns the first pair of a list where the car IsEqual to the given
// object.
func (pair *Pair) Assoc(obj Object) *Pair {
	for node := pair; node != nil; node = node.Tail() {
		if p, ok := node.car.(*Pair); ok {
			if p.car.IsEqual(obj) {
				return p
			}
		}
	}
	return nil
}

// AppendBang updates the given pair by setting a new pair with the given
// object and nil as its new second object.
func (pair *Pair) AppendBang(obj Object) *Pair {
	if pair == nil || !IsNil(pair.cdr) {
		panic("AppendBang")
	}

	t := &Pair{car: obj}
	pair.cdr = t
	return t
}

// ExtendBang updates the given pair by extending it with the second pair list
// after its end. Returns the last list node of the newly formed list
// beginning with `lst`, which is also the last list node of the list starting
// with `val`.
func (pair *Pair) ExtendBang(obj *Pair) *Pair {
	if obj == nil {
		return pair
	}
	if pair == nil || !IsNil(pair.cdr) {
		panic("ExtendBang")
	}
	pair.cdr = obj
	elem := obj
	for {
		t := elem.Tail()
		if t == nil {
			return elem
		}
		elem = t
	}
}

// Reverse returns a reversed pair list.
func (pair *Pair) Reverse() (*Pair, error) {
	if pair == nil {
		return nil, nil
	}
	result := Nil()
	for node := pair; ; {
		result = result.Cons(node.Car())
		cdr := node.Cdr()
		if IsNil(cdr) {
			return result, nil
		}
		if next, isPair := GetPair(cdr); isPair {
			node = next
			continue
		}
		return nil, ErrImproper{Pair: pair}
	}
}

// Copy returns a copy of the given pair list.
func (pair *Pair) Copy() *Pair {
	if pair == nil {
		return nil
	}
	result := Cons(pair.car, pair.cdr)
	prev := result
	for {
		curr := prev.cdr
		if next, isPair := GetPair(curr); isPair && next != nil {
			copy := Cons(next.car, next.cdr)
			prev.SetCdr(copy)
			prev = copy
			continue
		}
		prev.SetCdr(curr)
		return result
	}
}

// ErrImproper is signalled if an improper list is found where it is not
// appropriate.
type ErrImproper struct{ Pair *Pair }

// Error returns a textual representation for this error.
func (err ErrImproper) Error() string { return fmt.Sprintf("improper list: %v", err.Pair) }
