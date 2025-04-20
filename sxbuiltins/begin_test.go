//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
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
	{name: "begin-cons", src: "(begin 1 2 . 3)", exp: "3"},

	// (and ...)
	{name: "and-0", src: "(and)", exp: "T"},
	{name: "begin-T", src: "(and 3)", exp: "3"},
	{name: "begin-NIL", src: "(and ())", exp: "()"},
	{name: "begin-T-T", src: "(and 3 4)", exp: "4"},
	{name: "begin-T-NIL", src: "(and 3 \"\")", exp: "\"\""},
	{name: "begin-T-NIL-T", src: "(and 3 \"\" 7)", exp: "\"\""},
	{name: "begin-NIL-T", src: "(and NIL 4)", exp: "()"},
}
