//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestQuote(t *testing.T) {
	t.Parallel()
	tcsQuote.Run(t)
}

var tcsQuote = tTestCases{
	{name: "quote-sym", src: "quote", exp: "#<syntax:quote>"},
	{name: "quote-zero", src: "(quote 0)", exp: "0"},
	{name: "quote-nil", src: "(quote ())", exp: "()"},
	{name: "quote-list", src: "(quote (1 2 3))", exp: "(1 2 3)"},
	{name: "quote-zero-macro", src: "'0", exp: "0"},
	{name: "quote-nil-macro", src: "'()", exp: "()"},
	{name: "quote-list-macro", src: "'(1 2 3)", exp: "(1 2 3)"},
	{name: "quote-quote-num", src: "(quote (quote 5))", exp: "(quote 5)"},
	{name: "err-quote-EOF", src: "'", exp: "{[{unexpected EOF}]}", withErr: true},
	{name: "err-quote-0", src: "(quote)", exp: "{[{quote: no arguments given}]}", withErr: true},
	{name: "err-quote-2", src: "(quote x 7)", exp: "{[{quote: more than one argument: (x 7)}]}", withErr: true},
	{name: "err-quote-improper", src: "(quote . 7 )", exp: "{[{quote: no arguments given}]}", withErr: true},
}
