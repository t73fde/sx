//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf_test

import (
	"testing"

	"zettelstore.de/sx.fossil/sxpf"
)

func TestGetNumber(t *testing.T) {
	if _, ok := sxpf.GetNumber(nil); ok {
		t.Error("nil is not a number")
	}
	if _, ok := sxpf.GetNumber((sxpf.Number)(nil)); ok {
		t.Error("nil number is not a number")
	}
	var o sxpf.Object = sxpf.Int64(17)
	res, ok := sxpf.GetNumber(o)
	if !ok {
		t.Error("Is a number:", o)
	} else if !o.IsEqual(res) {
		t.Error("Different numbers, exptected:", o, "but got:", res)
	}
}
