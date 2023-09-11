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

	// For (and ...) and (or ...) bollean values should be variables, so that nothing is optimized
	{name: "and-0", src: "(and)", exp: "1"},
	{name: "and-1-T", src: "(and TRUE)", exp: "1"},
	{name: "and-1-F", src: "(and FALSE)", exp: "()"},
	{name: "and-2-TF", src: "(and TRUE FALSE)", exp: "()"},
	{name: "and-2-FT", src: "(and FALSE TRUE)", exp: "()"},
	{name: "and-2-TF-noopt", src: "(and 1 ())", exp: "()"},
	{name: "and-2-FT-noopt", src: "(and () 1)", exp: "()"},
	{name: "and-shotcut", src: "(and FALSE (div ONE ZERO))", exp: "()"},
	{name: "or-0", src: "(or)", exp: "()"},
	{name: "or-1-T", src: "(or TRUE)", exp: "1"},
	{name: "or-1-F", src: "(or FALSE)", exp: "()"},
	{name: "or-2-TF", src: "(or TRUE FALSE)", exp: "1"},
	{name: "or-2-FT", src: "(or FALSE TRUE)", exp: "1"},
	{name: "or-2-TF-noopt", src: "(or 1 ())", exp: "1"},
	{name: "or-2-FT-noopt", src: "(or () 1)", exp: "1"},
	{name: "or-shotcut", src: "(or TRUE (div ONE ZERO))", exp: "1"},
	{
		name: "",
		src:  "",
		exp:  "",
	},
	{
		name: "",
		src:  "",
		exp:  "",
	},
}
