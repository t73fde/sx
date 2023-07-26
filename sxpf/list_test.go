//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf_test

import (
	"testing"

	"zettelstore.de/sx.fossil/sxpf"
)

func TestListNil(t *testing.T) {
	t.Parallel()

	var obj sxpf.Object
	if obj != nil {
		t.Error("A nil interface value is not equal to nil")
	}
	if !sxpf.IsNil(obj) {
		t.Error("A nil interface value is not considered IsNil(val)")
	}
	if obj == sxpf.Nil() {
		t.Error("A nil interface value is wrongly treated as Nil()")
	}

	var i sxpf.Number
	if !sxpf.IsNil(i) {
		t.Error("A uninitialized integerpointer  is not considered IsNil(val)")
	}

	var pair *sxpf.Pair
	if pair != sxpf.Nil() {
		t.Error("An uninitialized pair pointer is not Nil()")
	}
	if !sxpf.IsNil(pair) {
		t.Error("An uninitialized pair pointer is not IsNil(p)")
	}
}

func TestGetList(t *testing.T) {
	t.Parallel()

	if res, isPair := sxpf.GetPair(nil); !isPair {
		t.Error("nil is a list")
	} else if res != nil {
		t.Error("Nil() must be nil")
	}
	res, isPair := sxpf.GetPair(sxpf.Nil())
	if !isPair {
		t.Error("Nil() is a list")
	} else if res != nil {
		t.Error("Nil() must be nil")
	}
	if _, isPair = sxpf.GetPair(sxpf.MakeString("nil")); isPair {
		t.Error("A string is not a list")
	}
}

func TestListIsList(t *testing.T) {
	t.Parallel()
	if !sxpf.IsList(nil) {
		t.Error("nil is a list")
	}
	if !sxpf.IsList(sxpf.Nil()) {
		t.Error("Nil() is a list")
	}
	if !sxpf.IsList(sxpf.MakeList(sxpf.Nil(), sxpf.Nil())) {
		t.Error("MakeList produces lists")
	}
	one := sxpf.Int64(1)
	if sxpf.IsList(sxpf.Cons(one, one)) {
		t.Error("(1 . 1) is not a list")
	}
	if sxpf.IsList(sxpf.Cons(one, sxpf.Cons(one, one))) {
		t.Error("(1 1 . 1) is not a list")
	}
}

func TestListLength(t *testing.T) {
	t.Parallel()

	if got := sxpf.Nil().Length(); got != 0 {
		t.Error("Nil().Length() != 0, but", got)
	}
	objs := make([]sxpf.Object, 0, 100)
	for i := 0; i < cap(objs); i++ {
		objs = append(objs, sxpf.Nil())
		l := sxpf.MakeList(objs...)
		if got := l.Length(); got != len(objs) {
			t.Errorf("List %v should contain %d element, but got %d", l, i, got)
		}
	}
}

func TestListAssoc(t *testing.T) {
	t.Parallel()

	val1, val2 := sxpf.Int64(1), sxpf.Int64(-1)
	p1, p2 := sxpf.Cons(val1, val2), sxpf.Cons(val2, sxpf.Nil())
	p3 := sxpf.Cons(sxpf.Nil(), val1)
	lst1 := sxpf.MakeList(p1, p2)
	testcases := []struct {
		name string
		list *sxpf.Pair
		val  sxpf.Object
		exp  *sxpf.Pair
	}{
		{name: "AllEmpty", list: nil, val: nil, exp: nil},
		{name: "ListEmpty", list: nil, val: val1, exp: nil},
		{name: "FoundFirst", list: lst1, val: val1, exp: p1},
		{name: "FoundSecond", list: lst1, val: val2, exp: p2},
		{name: "FoundNix", list: lst1, val: sxpf.Nil(), exp: nil},
		{name: "FoundAgain", list: lst1.Cons(p3), val: sxpf.Nil(), exp: p3},
		{name: "NoAList", list: sxpf.MakeList(val1, val2, p3), val: val2, exp: nil},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.list.Assoc(tc.val); !tc.exp.IsEqual(got) {
				t.Errorf("%v.Assoc(%v) is %v, but got %v", tc.list, tc.val, tc.exp, got)
			}
		})
	}
}

func TestListReverse(t *testing.T) {
	t.Parallel()

	res, err := sxpf.Nil().Reverse()
	if err != nil {
		t.Error("ERR1", err)
	} else if !sxpf.IsNil(res) {
		t.Error("REV1", res)
	}

	res, err = sxpf.MakeList(sxpf.Int64(1)).Reverse()
	if err != nil {
		t.Error("ERR2", err)
	} else if !sxpf.Int64(1).IsEqual(res.Car()) {
		t.Error("RES2", res)
	}

	res, err = sxpf.MakeList(sxpf.Int64(1), sxpf.Int64(2)).Reverse()
	if err != nil {
		t.Error("ERR3", err)
	} else if !sxpf.Int64(2).IsEqual(res.Car()) || !sxpf.Int64(1).IsEqual(res.Tail().Car()) {
		t.Error("RES3", res)
	}

	res, err = sxpf.Cons(sxpf.Int64(1), sxpf.Int64(2)).Reverse()
	if err == nil {
		t.Error("ERR4", res)
	}
}
