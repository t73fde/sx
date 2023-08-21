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
}
