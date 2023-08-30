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

func TestStrings(t *testing.T) {
	t.Parallel()
	tcsStrings.Run(t)
}

var tcsStrings = tTestCases{
	{name: "err->string-0", src: "(->string)", exp: "{[{->string: exactly 1 arguments required, but 0 given: []}]}", withErr: true},
	{name: "->string-1", src: "(->string 1)", exp: `"1"`},
	{name: "->string-2", src: "(->string 1 2)", exp: "{[{->string: exactly 1 arguments required, but 2 given: [1 2]}]}", withErr: true},
	{name: "->string-cons", src: "(->string (cons 1 2))", exp: `"(1 . 2)"`},
	{name: "->string-string", src: `(->string "a")`, exp: `"a"`},

	{name: "string-append-0", src: "(string-append)", exp: `""`},
	{name: "err-string-append-1", src: "(string-append 1)", exp: "{[{string-append: argument 1 is not a string, but sx.Int64/1}]}", withErr: true},
	{name: "err-string-append-2", src: `(string-append "1" 2)`, exp: "{[{string-append: argument 2 is not a string, but sx.Int64/2}]}", withErr: true},
	{name: "string-append-1", src: `(string-append "a")`, exp: `"a"`},
	{name: "string-append-3", src: `(string-append "3" " " "4")`, exp: `"3 4"`},
}
