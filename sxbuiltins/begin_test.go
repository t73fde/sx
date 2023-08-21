//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestBegin(t *testing.T) {
	t.Parallel()
	tcsBegin.Run(t)
}

var tcsBegin = tTestCases{
	{name: "begin-0", src: "(begin)", exp: "()"},
	{name: "begin-1", src: "(begin 1)", exp: "1"},
	{name: "begin-2", src: "(begin 1 2)", exp: "2"},
}
