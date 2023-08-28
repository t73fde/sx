//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"zettelstore.de/sx.fossil"
)

func TestBoolean(t *testing.T) {
	t.Parallel()
	if sx.True == sx.False {
		t.Error("True and False are identical")
	}
	if sx.True.IsEql(sx.False) {
		t.Error("True eql False")
	}
	if sx.False.IsEqual(sx.True) {
		t.Error("False equal True")
	}
	if !sx.Nil().IsEqual(sx.False) {
		t.Error("Nil() is not equal to False")
	}
	if !sx.False.IsEqual(sx.Nil()) {
		t.Error("False is not equal to Nil()")
	}
	if sx.Nil().IsEqual(sx.True) {
		t.Error("Nil() is equal to True")
	}
	if sx.True.IsEqual(sx.Nil()) {
		t.Error("True is equal to Nil()")
	}
	if sx.IsTrue(sx.MakeString("")) {
		t.Error("Empty string is True")
	}
	if !sx.IsFalse(sx.MakeString("")) {
		t.Error("Empty string is not False")
	}
	checkBoolean(t, sx.False, sx.FalseString, "false")
	checkBoolean(t, sx.True, sx.TrueString, "true")
}

func checkBoolean(t *testing.T, b sx.Boolean, s, bs string) {
	if !b.IsAtom() {
		t.Error("not an atom:", b)
	}
	if b.String() != bs {
		t.Error("Boolean", b, "has wrong string value:", b.String(), "but expected:", bs)
	}
	if b.Repr() != s {
		t.Error("Boolean", b, "has wrong repr value:", b.Repr())
	}
	if sx.MakeBoolean(sx.IsTrue(b)) != b {
		t.Error("MakeBoolean(IsTrue) is not", b)
	}
	if b.IsNil() || sx.IsNil(b) {
		t.Error(b, "is Nil()")
	}
}
