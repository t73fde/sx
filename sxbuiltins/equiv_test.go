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

func TestEquiv(t *testing.T) {
	t.Parallel()
	tcsEquiv.Run(t)
}

var tcsEquiv = tTestCases{
	{name: "err-==-0",
		src:     "(==)",
		exp:     "{[{==: at least 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-==-1",
		src:     "(== 1)",
		exp:     "{[{==: at least 2 arguments required, but only 1 given: [1]}]}",
		withErr: true,
	},
	{name: "==-2-f", src: "(== 1 2)", exp: "()"},
	{name: "==-2-t", src: "(== 1 1)", exp: "T"},

	{name: "err-=-0",
		src:     "(=)",
		exp:     "{[{=: at least 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-=-1",
		src:     "(= 1)",
		exp:     "{[{=: at least 2 arguments required, but only 1 given: [1]}]}",
		withErr: true,
	},
	{name: "=-2-f", src: "(= 1 2)", exp: "()"},
	{name: "=-2-t", src: "(= 1 1)", exp: "T"}}
