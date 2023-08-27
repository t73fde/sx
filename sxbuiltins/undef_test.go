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

func TestUndefined(t *testing.T) {
	t.Parallel()
	tcsUndefined.Run(t)
}

var tcsUndefined = tTestCases{
	{
		name:    "err-undefined-0",
		src:     "(undefined?)",
		exp:     "{[{undefined?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-undefined-2",
		src:     "(undefined? 1 2)",
		exp:     "{[{undefined?: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "undefined-a", src: "(undefined? 'a)", exp: "False"},
	{name: "undefined-lookup-xyz", src: "(undefined? (lookup 'xyz))", exp: "True"},

	{
		name:    "err-defined-0",
		src:     "(defined?)",
		exp:     "{[{defined?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-defined-2",
		src:     "(defined? 1 2)",
		exp:     "{[{defined?: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "defined-a", src: "(defined? 'a)", exp: "True"},
	{name: "defined-lookup-xyz", src: "(defined? (lookup 'xyz))", exp: "False"},
}
