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

func TestStrings(t *testing.T) {
	t.Parallel()
	tcsStrings.Run(t)
}

var tcsStrings = tTestCases{
	{name: "err->string-0",
		src:     "(->string)",
		exp:     "{[{->string: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "->string-1", src: "(->string 1)", exp: `"1"`},
	{name: "err->string-2",
		src:     "(->string 1 2)",
		exp:     "{[{->string: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "->string-cons", src: "(->string (cons 1 2))", exp: `"(1 . 2)"`},
	{name: "->string-string", src: `(->string "a")`, exp: `"a"`},

	{name: "concat-0",
		src: "(concat)", exp: `""`},
	{name: "err-concat-1",
		src:     "(concat 1)",
		exp:     "{[{concat: argument 1 is not a string, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-concat-1-1",
		src:     "(concat 1 1)",
		exp:     "{[{concat: argument 1 is not a string, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-concat-2",
		src:     `(concat "1" 2)`,
		exp:     "{[{concat: argument 2 is not a string, but sx.Int64/2}]}",
		withErr: true,
	},
	{name: "concat-1", src: `(concat "a")`, exp: `"a"`},
	{name: "concat-3", src: `(concat "3" " " "4")`, exp: `"3 4"`},
}
