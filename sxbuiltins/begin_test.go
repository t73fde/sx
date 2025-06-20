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
	{name: "begin-1-s", src: "(begin b)", exp: "11"},
	{name: "begin-1", src: "(begin (fb))", exp: "11"},
	{name: "begin-2", src: "(begin (fb) (fc))", exp: "(22 33)"},
	{name: "begin-2-s", src: "(begin b c)", exp: "(22 33)"},
	{name: "begin-10", src: "(begin (apply car '((1 2))) 1 2 3 4 5 6 7 8 9)", exp: "9"},
	{name: "begin-cons", src: "(begin (fb) (fc) . ((fd)))", exp: "(44 55)"},

	{name: "begin1-0", src: "(begin1)", exp: "()"},
	{name: "begin1-1", src: "(begin1 (fb))", exp: "11"},
	{name: "begin1-1-s", src: "(begin1 b)", exp: "11"},
	{name: "begin1-2", src: "(begin1 b c)", exp: "11"},
	{name: "begin1-2-s", src: "(begin1 (fb) (fc))", exp: "11"},
	{name: "begin1-10", src: "(begin1 (apply car '((1 2))) 1 2 3 4 5 6 7 8 9)", exp: "1"},
	{name: "begin1-cons", src: "(begin1 (fb) (fc) . ((fd)))", exp: "11"},
	{name: "begin-1-nested", src: "(begin1 (begin1 (fb) (fc)) (begin1 (fc) (fd)))", exp: "11"},

	// (and ...)
	{name: "and-0", src: "(and)", exp: "T"},
	{name: "and-T", src: "(and 3)", exp: "3"},
	{name: "and-NIL", src: "(and ())", exp: "()"},
	{name: "and-T-T", src: "(and 3 4)", exp: "4"},
	{name: "and-T-T-T", src: "(and (apply car '((1 2))) 3 4)", exp: "4"},
	{name: "and-T-NIL", src: "(and 3 \"\")", exp: "\"\""},
	{name: "and-T-NIL-T", src: "(and 3 \"\" 7)", exp: "\"\""},
	{name: "and-NIL-T", src: "(and NIL 4)", exp: "()"},

	// (or ...)
	{name: "or-0", src: "(or)", exp: "()"},
	{name: "or-T", src: "(or 3)", exp: "3"},
	{name: "or-NIL", src: "(or ())", exp: "()"},
	{name: "or-F-F", src: "(or () ())", exp: "()"},
	{name: "or-F-T", src: "(or () (apply car '((1 2))))", exp: "1"},
	{name: "or-NIL-T-T-T", src: "(or () (apply car '((1 2))) (or (apply car '((3 2)))) 4)", exp: "1"},
	{name: "or-T-NIL", src: "(or 3 \"\")", exp: "3"},
	{name: "or-NIL-T", src: "(or NIL 4 5)", exp: "4"},
}
