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
		src:     "(environment-bindings)",
		exp:     "{[{environment-bindings: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-bindings-1-noenv",
		src:     "(environment-bindings 1)",
		exp:     "{[{environment-bindings: argument 1 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "current-bindings", src: "(environment-bindings (current-environment))", exp: "()"},
	{
		name: "let-bindings",
		src:  "(let (a 3) (environment-bindings (current-environment)))",
		exp:  "((a . 3))",
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
	{name: "bound?-a", src: "(bound? 'a)", exp: "()"},
	{name: "bound?-b", src: "(bound? 'b)", exp: "1"},
	{name: "bound?-bound?", src: "(bound? 'bound?)", exp: "1"},

	{
		name:    "err-lookup-0",
		src:     "(environment-lookup)",
		exp:     "{[{environment-lookup: between 1 and 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-1-nosym",
		src:     "(environment-lookup 1)",
		exp:     "{[{environment-lookup: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-2-noenv",
		src:     "(environment-lookup 'a 1)",
		exp:     "{[{environment-lookup: argument 2 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-3-noenv",
		src:     "(environment-lookup 'a (current-environment) 1)",
		exp:     "{[{environment-lookup: between 1 and 2 arguments required, but 3 given: [a err-lookup-3-noenv 1]}]}",
		withErr: true,
	},
	{name: "lookup-b", src: "(environment-lookup 'b)", exp: "#<undefined>"},
	{
		name: "lookup-b-parent",
		src:  "(environment-lookup 'b (parent-environment (current-environment)))",
		exp:  "11",
	},

	{
		name:    "err-resolve-0",
		src:     "(environment-resolve)",
		exp:     "{[{environment-resolve: between 1 and 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-1-nosym",
		src:     "(environment-resolve 1)",
		exp:     "{[{environment-resolve: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-2-noenv",
		src:     "(environment-resolve 'a 1)",
		exp:     "{[{environment-resolve: argument 2 is not an environment, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-3-noenv",
		src:     "(environment-resolve 'a (current-environment) 1)",
		exp:     "{[{environment-resolve: between 1 and 2 arguments required, but 3 given: [a err-resolve-3-noenv 1]}]}",
		withErr: true,
	},
	{name: "resolve-b", src: "(environment-resolve 'b)", exp: "11"},
	{
		name: "resolve-b-parent",
		src:  "(environment-resolve 'b (parent-environment (current-environment)))",
		exp:  "11",
	},
	{name: "resolve-xyz", src: "(environment-resolve 'xyz)", exp: "#<undefined>"},
}
