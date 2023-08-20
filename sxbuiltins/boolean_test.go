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
	{name: "TRUE", src: "TRUE", exp: "True"},
	{name: "FALSE", src: "FALSE", exp: "False"},
	{
		name: "boolean?",
		src:  "(map boolean? (list True False () 0 'a))",
		exp:  "(True True False False False)",
	},
	{
		name: "boolean?-var",
		src:  "(map boolean? (list TRUE FALSE () 0 'a))",
		exp:  "(True True False False False)",
	},
	{
		name:    "boolean?-err-0",
		src:     "(boolean?)",
		exp:     "{[{boolean?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "boolean?-err-2",
		src:     "(boolean? True 3)",
		exp:     "{[{boolean?: exactly 1 arguments required, but 2 given: [true 3]}]}",
		withErr: true,
	},
	{
		name: "boolean",
		src:  "(map boolean (list True False 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(True False True True False True True False)",
	},
	{
		name: "boolean-var",
		src:  "(map boolean (list TRUE FALSE 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(True False True True False True True False)",
	},
	{
		name:    "boolean-err-0",
		src:     "(boolean)",
		exp:     "{[{boolean: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "boolean-err-2",
		src:     "(boolean True 3)",
		exp:     "{[{boolean: exactly 1 arguments required, but 2 given: [true 3]}]}",
		withErr: true,
	},
	{
		name: "Not",
		src:  "(map not (list True False 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(False True False False True False False True)",
	},
	{
		name: "Not-var",
		src:  "(map not (list TRUE FALSE 0 1 \"\" \"0\" '(1 2 3) ()))",
		exp:  "(False True False False True False False True)",
	},
	{
		name:    "not-err-0",
		src:     "(not)",
		exp:     "{[{not: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "not-err-2",
		src:     "(not True 3)",
		exp:     "{[{not: exactly 1 arguments required, but 2 given: [true 3]}]}",
		withErr: true,
	},

	// For (and ...) and (or ...) bollean values should be variables, so that nothing is optimized
	{name: "and-0", src: "(and)", exp: "True"},
	{name: "and-1-T", src: "(and TRUE)", exp: "True"},
	{name: "and-1-F", src: "(and FALSE)", exp: "False"},
	{name: "and-2-TF", src: "(and TRUE FALSE)", exp: "False"},
	{name: "and-2-FT", src: "(and FALSE TRUE)", exp: "False"},
	{name: "and-2-TF-noopt", src: "(and True False)", exp: "False"},
	{name: "and-2-FT-noopt", src: "(and False True)", exp: "False"},
	{name: "and-shotcut", src: "(and FALSE (div ONE ZERO))", exp: "False"},
	{name: "or-0", src: "(or)", exp: "False"},
	{name: "or-1-T", src: "(or TRUE)", exp: "True"},
	{name: "or-1-F", src: "(or FALSE)", exp: "False"},
	{name: "or-2-TF", src: "(or TRUE FALSE)", exp: "True"},
	{name: "or-2-FT", src: "(or FALSE TRUE)", exp: "True"},
	{name: "or-2-TF-noopt", src: "(or True False)", exp: "True"},
	{name: "or-2-FT-noopt", src: "(or False True)", exp: "True"},
	{name: "or-shotcut", src: "(or True (div ONE ZERO))", exp: "True"},
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
