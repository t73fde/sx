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

func TestBoolean(t *testing.T) {
	t.Parallel()
	tcsBoolean.Run(t)
}

var tcsBoolean = tTestCases{
	{name: "T-sym", src: "T", exp: "T"},

	{name: "err-not-0",
		src:     "(not)",
		exp:     "{[{not: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-not-2",
		src:     "(not 1 2)",
		exp:     "{[{not: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "not-nil", src: "(not ())", exp: "T"},
	{name: "not not-nil", src: "(not (not ()))", exp: "()"},
	{name: "not-T", src: "(not 'T)", exp: "()"},
	{name: "not-not-T", src: "(not (not 1))", exp: "T"},
}
