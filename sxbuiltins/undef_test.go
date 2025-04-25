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

func TestUndefined(t *testing.T) {
	t.Parallel()
	tcsUndefined.Run(t)
}

var tcsUndefined = tTestCases{
	{name: "err-defined-0",
		src:     "(defined?)",
		exp:     "{[{defined?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-defined-2",
		src:     "(defined? 1 2)",
		exp:     "{[{defined?: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "defined-a", src: "(defined? 'a)", exp: "T"},
	{name: "defined-lookup-xyz", src: "(defined? (binding-lookup 'xyz))", exp: "()"},
}
