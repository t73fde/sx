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
	"iter"
	"strings"
)

// Pair is a node containing a value for the element and a pointer to the tail.
// In other lisps it is often called "cell", "cons", "cons-cell", or "list".
type Pair struct {
	car Object
	cdr Object
}

// Nil returns the nil list.
func Nil() *Pair { return (*Pair)(nil) }

// Cons prepends a value in front of a given listreturning the new list.
func (pair *Pair) Cons(car Object) *Pair { return &Pair{car: car, cdr: pair} }

// Cons creates a pair, often called a "cons cell".
func Cons(car, cdr Object) *Pair { return &Pair{car: car, cdr: cdr} }

// MakeList creates a new list with the given objects.
func MakeList(objs ...Object) *Pair {
	var lb ListBuilder
	for _, obj := range objs {
		lb.Add(obj)
	}
	return lb.List()
}

// IsNil return true, if it is a nil pair object.
func (pair *Pair) IsNil() bool { return pair == nil }

// IsAtom returns true, if the list is an atom.
func (pair *Pair) IsAtom() bool { return pair == nil }

// IsEqual compare two objects.
func (pair *Pair) IsEqual(other Object) bool {
	if pair == other {
		return true
	}
	if pair.IsNil() {
		return IsNil(other)
	}
	if otherPair, ok := other.(*Pair); ok {
		node, otherNode := pair, otherPair
		for node != nil && otherNode != nil {
			if !node.Car().IsEqual(otherNode.Car()) {
				return false
			}
			cdr, otherCdr := node.Cdr(), otherNode.Cdr()
			if IsNil(cdr) {
				return IsNil(otherCdr)
			}
			next, isPair := GetPair(cdr)
			if !isPair {
				return cdr.IsEqual(otherCdr)
			}
			otherNext, isPair := GetPair(otherCdr)
			if !isPair {
				return false
			}
			node, otherNode = next, otherNext
		}
		return node == otherNode
	}
	return false
}

// String returns the string representation.
func (pair *Pair) String() string {
	var sb strings.Builder
	if _, err := pair.Print(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

// GoString returns the go string representation.
func (pair *Pair) GoString() string { return pair.String() }

// Print write the string representation to the given Writer.
func (pair *Pair) Print(w io.Writer) (int, error) {
	if pair == nil {
		return io.WriteString(w, "()")
	}
	l, err := io.WriteString(w, "(")
	if err != nil {
		return l, err
	}
	var len2 int

	for node := pair; ; {
		if node != pair {
			len2, err = io.WriteString(w, " ")
			l += len2
			if err != nil {
				return l, err
			}
		}
		len2, err = Print(w, node.car)
		l += len2
		if err != nil {
			return l, err
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
		l += len2
		if err != nil {
			return l, err
		}
		len2, err = Print(w, cdr)
		l += len2
		if err != nil {
			return l, err
		}
		break
	}

	len2, err = io.WriteString(w, ")")
	l += len2
	return l, err
}

// --- Sequence methods

// Length returns the length of the pair list.
//
// The list must not be circular.
func (pair *Pair) Length() int {
	result := 0
	for range pair.Pairs() {
		result++
	}
	return result
}

// LengthLess returns true if the length of the pair list is less than the
// given number.
//
// pair.LengthLess(n) is typically much faster than pair.Length() < n.
func (pair *Pair) LengthLess(n int) bool {
	count := 0
	for range pair.Pairs() {
		count++
		if count >= n {
			return false
		}
	}
	return count < n
}

// LengthGreater returns true if the length of the pair list is greater than
// the given number.
//
// pair.LengthGreater(n) is typically much faster than pair.Length() > n.
func (pair *Pair) LengthGreater(n int) bool {
	count := 0
	for range pair.Pairs() {
		count++
		if count > n {
			return true
		}
	}
	return count > n
}

// LengthEqual returns true if the length of the pair list is equal to the
// given number.
//
// pair.LengthEqual(n) is typically much faster than pair.Length() == n.
func (pair *Pair) LengthEqual(n int) bool {
	count := 0
	for range pair.Pairs() {
		count++
		if count > n {
			return false
		}
	}
	return count == n
}

// Nth returns the n'th object of the pair list. It is an error, if n < 0
// or if the list length is less than n.
func (pair *Pair) Nth(n int) (Object, error) {
	if n < 0 {
		return Nil(), fmt.Errorf("negative index %d", n)
	}
	cnt := 0
	for val := range pair.Values() {
		if cnt == n {
			return val, nil
		}
		cnt++
	}
	return Nil(), fmt.Errorf("index too large: %d for %v", n, pair)
}

// MakeList builds a list. Basically, the same pair is returned.
// This method is needed to make Sequence interface happy.
func (pair *Pair) MakeList() *Pair { return pair }

// Values returns an iterator of all objects in the pair list.
func (pair *Pair) Values() iter.Seq[Object] {
	return func(yield func(Object) bool) {
		for node := pair; node != nil; node = node.Tail() {
			if !yield(node.car) {
				return
			}
		}
	}
}

// --- Pair / list methods

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

// SetCar sets the car of the pair to the given object.
func (pair *Pair) SetCar(obj Object) {
	if pair != nil {
		pair.car = obj
	}
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

// Assoc returns the first pair of a list where the car IsEqual to the given
// object.
func (pair *Pair) Assoc(obj Object) *Pair {
	for val := range pair.Values() {
		if p, ok := val.(*Pair); ok {
			if p.car.IsEqual(obj) {
				return p
			}
		}
	}
	return nil
}

// RemoveAssoc deletes all pairs from the association list, where the car
// IsEqual to the given object. A new list is created.
func (pair *Pair) RemoveAssoc(obj Object) *Pair {
	var result, curr *Pair
	for node := pair; node != nil; node = node.Tail() {
		if p, isPair := node.car.(*Pair); isPair {
			if p.car.IsEqual(obj) {
				continue
			}
			if result == nil {
				result = Cons(p, nil)
				curr = result
			} else {
				curr = curr.AppendBang(p)
			}
		}
	}
	return result
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
			cpy := Cons(next.car, next.cdr)
			prev.SetCdr(cpy)
			prev = cpy
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

// --- Go iterators

// Pairs returns an iterator of all pair nodes.
func (pair *Pair) Pairs() iter.Seq[*Pair] {
	return func(yield func(*Pair) bool) {
		for node := pair; node != nil; node = node.Tail() {
			if !yield(node) {
				return
			}
		}
	}
}

// ListBuilder is a helper to build a list sequentially from start to end.
type ListBuilder struct {
	first, last *Pair
}

// Reset the list builder.
func (lb *ListBuilder) Reset() {
	lb.first = nil
	lb.last = nil
}

// Add an object to the list builder.
func (lb *ListBuilder) Add(obj Object) {
	elem := Cons(obj, nil)
	if lb.first == nil {
		lb.first = elem
		lb.last = elem
		return
	}
	lb.last.cdr = elem
	lb.last = elem
}

// AddN adds multiple objects to the list builder.
func (lb *ListBuilder) AddN(objs ...Object) {
	if len(objs) == 0 {
		return
	}
	lb.Add(objs[0])
	// assert: lb.first != nil
	last := lb.last
	for _, obj := range objs[1:] {
		elem := Cons(obj, nil)
		last.cdr = elem
		last = elem
	}
	lb.last = last
}

// ExtendBang the list by the given list, reusing the given list
func (lb *ListBuilder) ExtendBang(lst *Pair) {
	if lst == nil {
		return
	}
	if lb.first == nil {
		lb.first = lst
		lb.last = lst.LastPair()
		return
	}
	lb.last.cdr = lst
	for {
		t := lst.Tail()
		if t == nil {
			lb.last = lst
			return
		}
		lst = t
	}
}

// List the result, but not resetting the builder.
func (lb *ListBuilder) List() *Pair { return lb.first }

// Last returns the last pair.
func (lb *ListBuilder) Last() *Pair { return lb.last }

// IsEmpty returns true, if no element was added.
func (lb *ListBuilder) IsEmpty() bool { return lb.first == nil }

// Collect all values of the iterator into a list.
func (lb *ListBuilder) Collect(seq iter.Seq[Object]) {
	for obj := range seq {
		lb.Add(obj)
	}
}
