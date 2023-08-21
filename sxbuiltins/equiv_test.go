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

func TestEquiv(t *testing.T) {
	t.Parallel()
	tcsEquiv.Run(t)
}

var tcsEquiv = tTestCases{
	{
		name:    "err-eq?-0",
		src:     "(eq?)",
		exp:     "{[{eq?: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-eq?-1",
		src:     "(eq? 1)",
		exp:     "{[{eq?: exactly 2 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},
	{name: "eq?-2-f", src: "(eq? 1 2)", exp: "False"},
	{name: "eq?-2-t", src: "(eq? 1 1)", exp: "True"},

	{
		name:    "err-eql?-0",
		src:     "(eql?)",
		exp:     "{[{eql?: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-eql?-1",
		src:     "(eql? 1)",
		exp:     "{[{eql?: exactly 2 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},
	{name: "eql?-2-f", src: "(eql? 1 2)", exp: "False"},
	{name: "eql?-2-t", src: "(eql? 1 1)", exp: "True"},

	{
		name:    "err-equal?-0",
		src:     "(equal?)",
		exp:     "{[{equal?: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-equal?-1",
		src:     "(equal? 1)",
		exp:     "{[{equal?: exactly 2 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},
	{name: "equal?-2-f", src: "(equal? 1 2)", exp: "False"},
	{name: "equal?-2-t", src: "(equal? 1 1)", exp: "True"}}
