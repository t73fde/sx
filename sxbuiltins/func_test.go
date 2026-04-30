//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestIdentity(t *testing.T) {
	t.Parallel()
	tcsIdentity.Run(t)
}

var tcsIdentity = tTestCases{
	{name: "T-sym", src: "T", exp: "T"},

	{name: "err-identity-0",
		src:     "(identity)",
		exp:     "{[{identity: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-identity-2",
		src:     "(identity 1 2)",
		exp:     "{[{identity: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "identity-nil", src: "(identity ())", exp: "()"},
	{name: "identity identity-nil", src: "(identity (identity ()))", exp: "()"},
	{name: "identity-3", src: "(identity 3)", exp: "3"},
	{name: "identity identity-3", src: "(identity (identity 3))", exp: "3"},
	{name: "identity map", src: "(map identity '(1 2 3))", exp: "(1 2 3)"},
}
