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

func TestEnv(t *testing.T) {
	t.Parallel()
	tcsEnv.Run(t)
}

var tcsEnv = tTestCases{
	{
		name:    "err-current-environment-1",
		src:     "(current-environment 1)",
		exp:     "{[{current-environment: exactly 0 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},

	{
		name:    "err-parent-environment-0",
		src:     "(parent-environment)",
		exp:     "{[{parent-environment: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-parent-environment-1-noenv",
		src:     "(parent-environment 1)",
		exp:     "{[{parent-environment: argument 1 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},

	{
		name:    "err-bindings-0",
		src:     "(bindings)",
		exp:     "{[{bindings: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-bindings-1-noenv",
		src:     "(bindings 1)",
		exp:     "{[{bindings: argument 1 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "current-bindings", src: "(bindings (current-environment))", exp: "()"},
	{
		name: "let-bindings",
		src:  "(let (a 3) (bindings (current-environment)))",
		exp:  "((a . 3))",
	},

	{
		name:    "err-all-bindings-0",
		src:     "(all-bindings)",
		exp:     "{[{all-bindings: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-all-bindings-1-noenv",
		src:     "(all-bindings 1)",
		exp:     "{[{all-bindings: argument 1 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name: "let-all-bindings",
		src:  "(car (let (a 3) (all-bindings (current-environment))))",
		exp:  "(a . 3)",
	},
	{
		name: "num-all-bindings",
		src:  "(< 70 (length (all-bindings (current-environment))))",
		exp:  "True",
	},

	{
		name:    "err-bound?-0",
		src:     "(bound?)",
		exp:     "{[{bound?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-bound?-1",
		src:     "(bound? 1)",
		exp:     "{[{bound?: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "bound?-a", src: "(bound? 'a)", exp: "False"},
	{name: "bound?-b", src: "(bound? 'b)", exp: "True"},
	{name: "bound?-bound?", src: "(bound? 'bound?)", exp: "True"},

	{
		name:    "err-lookup-0",
		src:     "(lookup)",
		exp:     "{[{lookup: between 1 and 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-1-nosym",
		src:     "(lookup 1)",
		exp:     "{[{lookup: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-2-noenv",
		src:     "(lookup 'a 1)",
		exp:     "{[{lookup: argument 2 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-3-noenv",
		src:     "(lookup 'a (current-environment) 1)",
		exp:     "{[{lookup: between 1 and 2 arguments required, but 3 given: [a err-lookup-3-noenv 1]}]}",
		withErr: true,
	},
	{name: "lookup-b", src: "(lookup 'b)", exp: "#<undefined>"},
	{
		name: "lookup-b-parent",
		src:  "(lookup 'b (parent-environment (current-environment)))",
		exp:  "11",
	},

	{
		name:    "err-resolve-0",
		src:     "(resolve)",
		exp:     "{[{resolve: between 1 and 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-1-nosym",
		src:     "(resolve 1)",
		exp:     "{[{resolve: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-2-noenv",
		src:     "(resolve 'a 1)",
		exp:     "{[{resolve: argument 2 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-3-noenv",
		src:     "(resolve 'a (current-environment) 1)",
		exp:     "{[{resolve: between 1 and 2 arguments required, but 3 given: [a err-resolve-3-noenv 1]}]}",
		withErr: true,
	},
	{name: "resolve-b", src: "(resolve 'b)", exp: "11"},
	{
		name: "resolve-b-parent",
		src:  "(resolve 'b (parent-environment (current-environment)))",
		exp:  "11",
	},
	{name: "resolve-xyz", src: "(resolve 'xyz)", exp: "#<undefined>"},
}
