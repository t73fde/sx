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

func TestDefine(t *testing.T) {
	t.Parallel()
	tcsDefine.Run(t)
}

var tcsDefine = tTestCases{
	{
		name:    "err-define-0",
		src:     "(define)",
		exp:     "{[{define: need at least two arguments}]}",
		withErr: true,
	},
	{
		name:    "err-define-1",
		src:     "(define 1)",
		exp:     "{[{define: argument 1 must be a symbol or a list, but is: sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-define-1a",
		src:     "(define a)",
		exp:     "{[{define: argument 2 missing}]}",
		withErr: true,
	},
	{
		name:    "err-define-a-1",
		src:     "(define a . 1)",
		exp:     "{[{define: argument 2 must be a proper list}]}",
		withErr: true,
	},
	{name: "define-a-1", src: "(define a 1)", exp: "1"},

	{
		name:    "err-deffn-0",
		src:     "(define ())",
		exp:     "{[{define: empty function head}]}",
		withErr: true,
	},
	{name: "err-deffn-1", src: "(define (a))", exp: "{[{define: missing body}]}", withErr: true},
	{
		name:    "err-deffn-1-nosym",
		src:     "(define (1))",
		exp:     "{[{define: first element in function head is not a symbol, but: sx.Int64/1}]}",
		withErr: true,
	},
	{name: "deffn", src: "(define (a) 1)", exp: "#<lambda:a>"},
	{name: "deffn-eval", src: "((define (a) 1))", exp: "1"},
}

func TestSetX(t *testing.T) {
	t.Parallel()
	tcsSetX.Run(t)
}

var tcsSetX = tTestCases{
	{
		name:    "err-set!-0",
		src:     "(set!)",
		exp:     "{[{set!: need at least two arguments}]}",
		withErr: true,
	},
	{
		name:    "err-set!-1",
		src:     "(set! 1)",
		exp:     "{[{set!: argument 1 must be a symbol, but is: sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-set!-1a", src: "(set! a)", exp: "{[{set!: argument 2 missing}]}", withErr: true},
	{
		name:    "err-set!-a-1",
		src:     "(set! a . 1)",
		exp:     "{[{set!: argument 2 must be a proper list}]}",
		withErr: true,
	},
	{
		name:    "set!-a-1",
		src:     "(set! a 1)",
		exp:     `{[{symbol "a" not bound in environment "set!-a-1"}]}`,
		withErr: true,
	},
	{
		name:    "set!-b-1",
		src:     "(set! b 1)",
		exp:     `{[{symbol "b" not bound in environment "set!-b-1"}]}`,
		withErr: true,
	},
	{name: "define-set", src: "(define a 1) (set! a 17)", exp: "1 17"},
}
