//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"t73f.de/r/sx"
)

func TestIsTrue(t *testing.T) {
	t.Parallel()
	if sx.IsTrue(sx.Nil()) {
		t.Error("Nil is True")
	}
	if sx.IsTrue(sx.String{}) {
		t.Error("Empty string is True")
	}
	if sx.IsTrue(sx.MakeUndefined()) {
		t.Error("Undefined is True")
	}
}

func TestIsFalse(t *testing.T) {
	t.Parallel()
	if !sx.IsFalse(sx.Nil()) {
		t.Error("Nil is True")
	}
	if !sx.IsFalse(sx.String{}) {
		t.Error("Empty string is True")
	}
	if !sx.IsFalse(sx.MakeUndefined()) {
		t.Error("Undefined is True")
	}
}
