//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf_test

import (
	"testing"

	"zettelstore.de/sx.fossil/sxpf"
)

func TestBoolean(t *testing.T) {
	t.Parallel()
	if sxpf.True == sxpf.False {
		t.Error("True and False are identical")
	}
	if sxpf.True.IsEql(sxpf.False) {
		t.Error("True eql False")
	}
	if sxpf.False.IsEqual(sxpf.True) {
		t.Error("False equal True")
	}
	if !sxpf.Nil().IsEqual(sxpf.False) {
		t.Error("Nil() is not equal to False")
	}
	if !sxpf.False.IsEqual(sxpf.Nil()) {
		t.Error("False is not equal to Nil()")
	}
	if sxpf.Nil().IsEqual(sxpf.True) {
		t.Error("Nil() is equal to True")
	}
	if sxpf.True.IsEqual(sxpf.Nil()) {
		t.Error("True is equal to Nil()")
	}
	if sxpf.IsTrue(sxpf.MakeString("")) {
		t.Error("Empty string is True")
	}
	if !sxpf.IsFalse(sxpf.MakeString("")) {
		t.Error("Empty string is not False")
	}
	checkBoolean(t, sxpf.False, sxpf.FalseString, "false")
	checkBoolean(t, sxpf.True, sxpf.TrueString, "true")
}

func checkBoolean(t *testing.T, b sxpf.Boolean, s, bs string) {
	if !b.IsAtom() {
		t.Error("not an atom:", b)
	}
	if b.String() != bs {
		t.Error("Boolean", b, "has wrong string value:", b.String(), "but expected:", bs)
	}
	if b.Repr() != s {
		t.Error("Boolean", b, "has wrong repr value:", b.Repr())
	}
	if b.Negate() != sxpf.Negate(b) {
		t.Error("Negate functions differ")
	}
	if sxpf.MakeBoolean(sxpf.IsTrue(b)) != b {
		t.Error("MakeBoolean(IsTrue) is not", b)
	}
	if b.IsNil() || sxpf.IsNil(b) {
		t.Error(b, "is Nil()")
	}
}
