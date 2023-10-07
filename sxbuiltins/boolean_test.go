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

func TestBoolean(t *testing.T) {
	t.Parallel()
	tcsBoolean.Run(t)
}

var tcsBoolean = tTestCases{
	{name: "TRUE", src: "TRUE", exp: "1"},
	{name: "FALSE", src: "FALSE", exp: "()"},
	{
		name: "boolean",
		src:  "(map boolean (list 1 () 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(1 () 1 1 () 1 1 ())",
	},
	{
		name: "boolean-var",
		src:  "(map boolean (list TRUE FALSE 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(1 () 1 1 () 1 1 ())",
	},
	{
		name:    "boolean-err-0",
		src:     "(boolean)",
		exp:     "{[{boolean: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "boolean-err-2",
		src:     "(boolean 2 3)",
		exp:     "{[{boolean: exactly 1 arguments required, but 2 given: [2 3]}]}",
		withErr: true,
	},
	{
		name: "Not",
		src:  "(map not (list 1 () 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(() 1 () () 1 () () 1)",
	},
	{
		name: "Not-var",
		src:  "(map not (list TRUE FALSE 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(() 1 () () 1 () () 1)",
	},
	{
		name:    "not-err-0",
		src:     "(not)",
		exp:     "{[{not: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "not-err-2",
		src:     "(not 2 3)",
		exp:     "{[{not: exactly 1 arguments required, but 2 given: [2 3]}]}",
		withErr: true,
	},
}
