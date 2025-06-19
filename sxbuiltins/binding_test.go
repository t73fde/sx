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
	{name: "err-symbol-bound?-0",
		src:     "(symbol-bound?)",
		exp:     "{[{symbol-bound?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-symbol-bound?-1",
		src:     "(symbol-bound? 1)",
		exp:     "{[{symbol-bound?: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "symbol-bound?-a", src: "(symbol-bound? 'a)", exp: "()"},
	{name: "symbol-bound?-b", src: "(symbol-bound? 'b)", exp: "T"},
	{name: "symbol-bound?-bound?", src: "(symbol-bound? 'symbol-bound?)", exp: "T"},

	{name: "err-resolve-symbol-0",
		src:     "(resolve-symbol)",
		exp:     "{[{resolve-symbol: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-resolve-symbol-1-nosym",
		src:     "(resolve-symbol 1)",
		exp:     "{[{resolve-symbol: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-resolve-symbol-1-nosym-2",
		src:     "(resolve-symbol 1 2)",
		exp:     "{[{resolve-symbol: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-resolve-symbol-2-noenv",
		src:     "(resolve-symbol 'a 1)",
		exp:     "{[{resolve-symbol: argument 2 is not a frame, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-resolve-symbol-3-noenv",
		src:     "(resolve-symbol 'a (current-frame) 1)",
		exp:     "{[{resolve-symbol: between 1 and 2 arguments required, but 3 given: [a <nil> 1]}]}",
		withErr: true,
	},
	{name: "resolve-symbol-b", src: "(resolve-symbol 'b)", exp: "11"},
	{name: "resolve-symbol-b-nil", src: "(resolve-symbol 'b ())", exp: "11"},
	{name: "resolve-symbol-b-parent",
		src: "(resolve-symbol 'b (parent-frame (current-frame)))",
		exp: "11",
	},
	{name: "resolve-symbol-xyz", src: "(resolve-symbol 'xyz)", exp: "#<undefined>"},
	{name: "resolve-symbol-xyz-2", src: "(resolve-symbol 'xyz (current-frame))", exp: "#<undefined>"},
}
