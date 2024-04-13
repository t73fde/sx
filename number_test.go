//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
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

func TestGetNumber(t *testing.T) {
	if _, ok := sx.GetNumber(nil); ok {
		t.Error("nil is not a number")
	}
	if _, ok := sx.GetNumber((sx.Number)(nil)); ok {
		t.Error("nil number is not a number")
	}
	var o sx.Object = sx.Int64(17)
	res, ok := sx.GetNumber(o)
	if !ok {
		t.Error("Is a number:", o)
	} else if !o.IsEqual(res) {
		t.Error("Different numbers, exptected:", o, "but got:", res)
	}
}
