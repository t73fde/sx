//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestError(t *testing.T) {
	t.Parallel()
	tcsError.Run(t)
}

var tcsError = tTestCases{
	{name: "error-0", src: "(error)", exp: "{[{unspecified user error}]}", withErr: true},
	{name: "error-1", src: "(error \"failure\")", exp: "{[{failure}]}", withErr: true},
	{name: "error-2", src: "(error 1 2)", exp: "{[{1 2}]}", withErr: true},

	{
		name:    "not-bound-error-0",
		src:     "(not-bound-error)",
		exp:     "{[{not-bound-error: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "not-bound-error-nosy",
		src:     "(not-bound-error 1)",
		exp:     "{[{argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "not-bound-error-sym",
		src:     "(not-bound-error 'sym)",
		exp:     "{[{symbol \"sym\" not bound in \"not-bound-error-sym\"}]}",
		withErr: true,
	},
	{
		name:    "not-bound-error-sym-noenv",
		src:     "(not-bound-error 'sym 1)",
		exp:     "{[{argument 2 is not a binding, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "not-bound-error-sym-env",
		src:     "(not-bound-error 'sym (current-binding))",
		exp:     "{[{symbol \"sym\" not bound in \"not-bound-error-sym-env\"}]}",
		withErr: true,
	},
	{
		name:    "not-bound-error-sym-env-parent",
		src:     "(not-bound-error 'sym (parent-binding (current-binding)))",
		exp:     "{[{symbol \"sym\" not bound in \"vars\"}]}",
		withErr: true,
	},
}
