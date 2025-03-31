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

func TestBinding(t *testing.T) {
	t.Parallel()
	tcsBinding.Run(t)
}

var tcsBinding = tTestCases{
	{
		name:    "err-current-binding-1",
		src:     "(current-binding 1)",
		exp:     "{[{current-binding: exactly 0 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},

	{
		name:    "err-parent-binding-0",
		src:     "(parent-binding)",
		exp:     "{[{parent-binding: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "err-parent-binding-1-noenv",
		src:     "(parent-binding 1)",
		exp:     "{[{parent-binding: argument 1 is not a binding, but sx.Int64/1}]}",
		withErr: true,
	},

	{
		name:    "err-bindings-0",
		src:     "(bindings)",
		exp:     "{[{bindings: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "err-bindings-1-noenv",
		src:     "(bindings 1)",
		exp:     "{[{bindings: argument 1 is not a binding, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "current-bindings", src: "(bindings (current-binding))", exp: "()"},
	{
		name: "let-bindings",
		src:  "(let ((a 3)) (bindings (current-binding)))",
		exp:  "((a . 3))",
	},

	{
		name:    "err-bound?-0",
		src:     "(bound?)",
		exp:     "{[{bound?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "err-bound?-1",
		src:     "(bound? 1)",
		exp:     "{[{bound?: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "bound?-a", src: "(bound? 'a)", exp: "()"},
	{name: "bound?-b", src: "(bound? 'b)", exp: "T"},
	{name: "bound?-bound?", src: "(bound? 'bound?)", exp: "T"},

	{
		name:    "err-lookup-0",
		src:     "(binding-lookup)",
		exp:     "{[{binding-lookup: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-1-nosym",
		src:     "(binding-lookup 1)",
		exp:     "{[{binding-lookup: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-2-noenv",
		src:     "(binding-lookup 'a 1)",
		exp:     "{[{binding-lookup: argument 2 is not a binding, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-lookup-3-noenv",
		src:     "(binding-lookup 'a (current-binding) 1)",
		exp:     "{[{binding-lookup: between 1 and 2 arguments required, but 3 given: [a (#<builtin:current-binding>) 1]}]}",
		withErr: true,
	},
	{name: "lookup-b", src: "(binding-lookup 'b)", exp: "#<undefined>"},
	{
		name: "lookup-b-parent",
		src:  "(binding-lookup 'b (parent-binding (current-binding)))",
		exp:  "11",
	},

	{
		name:    "err-resolve-0",
		src:     "(binding-resolve)",
		exp:     "{[{binding-resolve: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-1-nosym",
		src:     "(binding-resolve 1)",
		exp:     "{[{binding-resolve: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-2-noenv",
		src:     "(binding-resolve 'a 1)",
		exp:     "{[{binding-resolve: argument 2 is not a binding, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-resolve-3-noenv",
		src:     "(binding-resolve 'a (current-binding) 1)",
		exp:     "{[{binding-resolve: between 1 and 2 arguments required, but 3 given: [a (#<builtin:current-binding>) 1]}]}",
		withErr: true,
	},
	{name: "resolve-b", src: "(binding-resolve 'b)", exp: "11"},
	{
		name: "resolve-b-parent",
		src:  "(binding-resolve 'b (parent-binding (current-binding)))",
		exp:  "11",
	},
	{name: "resolve-xyz", src: "(binding-resolve 'xyz)", exp: "#<undefined>"},
}
