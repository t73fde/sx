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

func TestBinding(t *testing.T) {
	t.Parallel()
	tcsBinding.Run(t)
}

var tcsBinding = tTestCases{
	{name: "err-let-0", src: "(let)", exp: "{[{let: binding spec and body missing}]}", withErr: true},
	{
		name:    "err-let-num",
		src:     "(let 1)",
		exp:     "{[{let: binding spec must be a list, but got: %!t(sx.Int64=1)/1}]}",
		withErr: true,
	},
	{
		name:    "err-let-num",
		src:     "(let 1)",
		exp:     "{[{let: binding spec must be a list, but got: %!t(sx.Int64=1)/1}]}",
		withErr: true,
	},
	{name: "err-let-improper", src: "(let () . 1)", exp: "{[{let: improper list: (() . 1)}]}", withErr: true},
	{name: "err-let-nobinding", src: "(let (a) 1)", exp: "{[{let: binding missing for a}]}", withErr: true},
	{
		name:    "err-let-improper-binding",
		src:     "(let (a . 1) a)",
		exp:     "{[{let: improper list: (a . 1)}]}",
		withErr: true,
	},
	{
		name:    "err-let-improper-binding-2",
		src:     "(let (a 1 . 2) a)",
		exp:     "{[{let: improper list: (a 1 . 2)}]}",
		withErr: true,
	},
	{name: "let-nil-1", src: "(let () 1)", exp: "1"},
	{name: "let-a-b", src: "(let (a 1 b 2) a b)", exp: "2"},
	{name: "let-nested-0", src: "(let (a 1) (let (a 2) a))", exp: "2"},
	{
		name:    "err-let-double-sym",
		src:     "(let (a 1 a 2) a)",
		exp:     "{[{let: symbol a already defined}]}",
		withErr: true,
	},
}
