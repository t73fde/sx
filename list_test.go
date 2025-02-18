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

package sx_test

import (
	"testing"

	"t73f.de/r/sx"
)

func TestListNil(t *testing.T) {
	t.Parallel()

	var obj sx.Object
	if !sx.IsNil(obj) {
		t.Error("A nil interface value is not considered IsNil(val)")
	}

	var i sx.Number
	if !sx.IsNil(i) {
		t.Error("A uninitialized integerpointer  is not considered IsNil(val)")
	}

	var pair *sx.Pair
	if pair != sx.Nil() {
		t.Error("An uninitialized pair pointer is not Nil()")
	}
	if !sx.IsNil(pair) {
		t.Error("An uninitialized pair pointer is not IsNil(p)")
	}
}

func TestGetList(t *testing.T) {
	t.Parallel()

	if res, isPair := sx.GetPair(nil); !isPair {
		t.Error("nil is a list")
	} else if res != nil {
		t.Error("Nil() must be nil")
	}
	res, isPair := sx.GetPair(sx.Nil())
	if !isPair {
		t.Error("Nil() is a list")
	} else if res != nil {
		t.Error("Nil() must be nil")
	}
	if _, isPair = sx.GetPair(sx.MakeString("nil")); isPair {
		t.Error("A string is not a list")
	}
}

func TestListIsList(t *testing.T) {
	t.Parallel()
	if !sx.IsList(nil) {
		t.Error("nil is a list")
	}
	if !sx.IsList(sx.Nil()) {
		t.Error("Nil() is a list")
	}
	if !sx.IsList(sx.MakeList(sx.Nil(), sx.Nil())) {
		t.Error("MakeList produces lists")
	}
	one := sx.Int64(1)
	if sx.IsList(sx.Cons(one, one)) {
		t.Error("(1 . 1) is not a list")
	}
	if sx.IsList(sx.Cons(one, sx.Cons(one, one))) {
		t.Error("(1 1 . 1) is not a list")
	}
}

func TestListLength(t *testing.T) {
	t.Parallel()

	if got := sx.Nil().Length(); got != 0 {
		t.Error("Nil().Length() != 0, but", got)
	}
	objs := make(sx.Vector, 0, 100)
	for i := range cap(objs) {
		objs = append(objs, sx.Nil())
		l := sx.MakeList(objs...)
		if got := l.Length(); got != len(objs) {
			t.Errorf("List %v should contain %d element, but got %d", l, i, got)
		}
	}
}

func TestPairIsEqual(t *testing.T) {
	t.Parallel()

	if !sx.Nil().IsEqual(sx.Nil()) {
		t.Error("Nil() != Nil()")
	}
	sym1, sym2 := sx.MakeSymbol("sym1"), sx.MakeSymbol("sym2")
	if sx.MakeList(sym1, sym2).IsEqual(sx.MakeList(sym1, sym1)) {
		t.Error("(sym1 sym2) == (sym1 sym1)")
	}
	if sx.Cons(sym1, sym2).IsEqual(sx.Cons(sym1, sym1)) {
		t.Error("(sym1 . sym2) == (sym1 . sym1)")
	}
}

func TestListAssoc(t *testing.T) {
	t.Parallel()

	val1, val2 := sx.Int64(1), sx.Int64(-1)
	p1, p2 := sx.Cons(val1, val2), sx.Cons(val2, sx.Nil())
	p3 := sx.Cons(sx.Nil(), val1)
	lst1 := sx.MakeList(p1, p2)
	testcases := []struct {
		name string
		list *sx.Pair
		val  sx.Object
		exp  *sx.Pair
	}{
		{name: "AllEmpty", list: nil, val: nil, exp: nil},
		{name: "ListEmpty", list: nil, val: val1, exp: nil},
		{name: "FoundFirst", list: lst1, val: val1, exp: p1},
		{name: "FoundSecond", list: lst1, val: val2, exp: p2},
		{name: "FoundNix", list: lst1, val: sx.Nil(), exp: nil},
		{name: "FoundAgain", list: lst1.Cons(p3), val: sx.Nil(), exp: p3},
		{name: "NoAList", list: sx.MakeList(val1, val2, p3), val: val2, exp: nil},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.list.Assoc(tc.val); !tc.exp.IsEqual(got) {
				t.Errorf("%v.Assoc(%v) is %v, but got %v", tc.list, tc.val, tc.exp, got)
			}
		})
	}
}

func TestListRemoveAssoc(t *testing.T) {
	testcases := []struct {
		name string
		list *sx.Pair
		obj  sx.Object
		exp  *sx.Pair
	}{
		{name: "AllEmpty", list: nil, obj: nil, exp: nil},
		{
			name: "RemoveFirstOnly",
			list: sx.MakeList(sx.Cons(sx.Int64(3), nil)),
			obj:  sx.Int64(3),
			exp:  nil,
		},
		{
			name: "RemoveFirstOnlyMultiple",
			list: sx.MakeList(sx.Cons(sx.Int64(3), nil), sx.Cons(sx.Int64(3), nil)),
			obj:  sx.Int64(3),
			exp:  nil,
		},
		{
			name: "RemoveFirstRest",
			list: sx.MakeList(sx.Cons(sx.Int64(3), nil), sx.Cons(sx.Int64(5), nil)),
			obj:  sx.Int64(3),
			exp:  sx.MakeList(sx.Cons(sx.Int64(5), nil)),
		},
		{
			name: "RemoveFirstRestMultiple",
			list: sx.MakeList(sx.Cons(sx.Int64(3), nil), sx.Cons(sx.Int64(3), nil), sx.Cons(sx.Int64(5), nil)),
			obj:  sx.Int64(3),
			exp:  sx.MakeList(sx.Cons(sx.Int64(5), nil)),
		},
		{
			name: "RemoveFirstLastMultiple",
			list: sx.MakeList(sx.Cons(sx.Int64(3), nil), sx.Cons(sx.Int64(5), nil), sx.Cons(sx.Int64(3), nil)),
			obj:  sx.Int64(3),
			exp:  sx.MakeList(sx.Cons(sx.Int64(5), nil)),
		},
		{
			name: "RemoveFirstLastMultipleLeaveMultiple",
			list: sx.MakeList(
				sx.Cons(sx.Int64(3), nil),
				sx.Cons(sx.Int64(5), nil),
				sx.Cons(sx.Int64(3), nil),
				sx.Cons(sx.Int64(5), nil),
			),
			obj: sx.Int64(3),
			exp: sx.MakeList(sx.Cons(sx.Int64(5), nil), sx.Cons(sx.Int64(5), nil)),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.list.RemoveAssoc(tc.obj); !tc.exp.IsEqual(got) {
				t.Errorf("%v.RemoveAssoc(%v) is %v, but got %v", tc.list, tc.obj, tc.exp, got)
			}
		})
	}
}

func TestListReverse(t *testing.T) {
	t.Parallel()

	res, err := sx.Nil().Reverse()
	if err != nil {
		t.Error("ERR1", err)
	} else if !sx.IsNil(res) {
		t.Error("REV1", res)
	}

	res, err = sx.MakeList(sx.Int64(1)).Reverse()
	if err != nil {
		t.Error("ERR2", err)
	} else if !sx.Int64(1).IsEqual(res.Car()) {
		t.Error("RES2", res)
	}

	res, err = sx.MakeList(sx.Int64(1), sx.Int64(2)).Reverse()
	if err != nil {
		t.Error("ERR3", err)
	} else if !sx.Int64(2).IsEqual(res.Car()) || !sx.Int64(1).IsEqual(res.Tail().Car()) {
		t.Error("RES3", res)
	}

	res, err = sx.Cons(sx.Int64(1), sx.Int64(2)).Reverse()
	if err == nil {
		t.Error("ERR4", res)
	}
}

func TestListCopy(t *testing.T) {
	testcases := []*sx.Pair{
		sx.Nil(),
		sx.Cons(sx.Nil(), sx.Nil()),
		sx.Cons(sx.Int64(3), sx.Nil()),
		sx.Cons(sx.Int64(3), sx.Int64(7)),
		sx.Cons(sx.Int64(3), sx.Nil()),
		sx.Cons(sx.Int64(3), sx.Cons(sx.Int64(5), sx.Nil())),
		sx.MakeList(sx.Int64(2), sx.Int64(3), sx.Int64(5), sx.Int64(7)),
	}
	for i, tc := range testcases {
		cpy := tc.Copy()
		if !tc.IsEqual(cpy) {
			t.Errorf("%d: %v != %v", i, tc, cpy)
		}
	}
}

func TestListBuilder(t *testing.T) {
	var lb sx.ListBuilder
	if !lb.IsEmpty() {
		t.Errorf("initial list is not empty, but: %v", lb.List())
	}
	lb.Add(sx.MakeSymbol("a"))
	if got, exp := lb.List(), sx.MakeList(sx.MakeSymbol("a")); !got.IsEqual(exp) {
		t.Errorf("expected %v, but got %v", exp, got)
	}
	if lb.IsEmpty() {
		t.Errorf("list is empty, expected: %v", lb.List())
	}
	lb.Reset()
	a, b, c := sx.MakeSymbol("a"), sx.MakeSymbol("b"), sx.MakeSymbol("c")
	lb.AddN()
	lb.AddN(a, b)
	lb.AddN(c)
	if got, exp := lb.List(), sx.MakeList(a, b, c); !got.IsEqual(exp) {
		t.Errorf("expected %v, but got %v", exp, got)
	}
	lb.Reset()

	lst := sx.MakeList(a)
	lb.ExtendBang(lst)
	lb.ExtendBang(nil)
	lb.ExtendBang(sx.MakeList(b, c))
	exp := sx.MakeList(a, b, c)
	if got := lb.List(); !got.IsEqual(exp) {
		t.Errorf("expected %v, but got %v", exp, got)
	}
	if !lst.IsEqual(exp) {
		t.Errorf("%v!=%v", lst, exp)
	}
}
