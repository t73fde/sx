//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestQuasiQuote(t *testing.T) {
	t.Parallel()
	tcsQuasiQuote.Run(t)
}

var tcsQuasiQuote = tTestCases{
	{name: "simple-quasiquote-sym", src: "quasiquote", exp: "#<special:quasiquote>"},
	{name: "simple-quasiquote-zero", src: "(quasiquote 0)", exp: "0"},
	{name: "simple-quasiquote-nil", src: "(quasiquote ())", exp: "()"},
	{name: "simple-quasiquote-list", src: "(quasiquote (1 2 3))", exp: "(1 2 3)"},
	{name: "simple-quasiquote-list-improper", src: "(quasiquote (1 2 . 3))", exp: "(1 2 . 3)"},
	{name: "simple-quasiquote-zero-macro", src: "`0", exp: "0"},
	{name: "simple-quasiquote-nil-macro", src: "`()", exp: "()"},
	{name: "simple-quasiquote-list-macro", src: "`(1 2 3)", exp: "(1 2 3)"},
	{name: "simple-quasiquote-list-macro-improper", src: "`(1 2 . 3)", exp: "(1 2 . 3)"},
	{
		name:    "err-quasiquote-0",
		src:     "(quasiquote)",
		exp:     "{[{quasiquote: no arguments given}]}",
		withErr: true,
	},
	{
		name:    "err-quasiquote-2",
		src:     "(quasiquote 1 2)",
		exp:     "{[{quasiquote: more than one argument: (1 2)}]}",
		withErr: true,
	},
	{
		name:    "err-quasiquote-nested-0",
		src:     "(quasiquote (quasiquote))",
		exp:     "{[{quasiquote: missing argument for quasiquote}]}",
		withErr: true,
	},
	{name: "err-quasiquote-EOF", src: "`", exp: "{[{unexpected EOF}]}", withErr: true},

	{name: "err-unquote", src: "(unquote)", exp: "{[{unquote: not allowed outside quasiquote}]}", withErr: true},
	{
		name:    "err-unquote-macro",
		src:     ",1",
		exp:     "{[{unquote: not allowed outside quasiquote}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-qq",
		src:     "`(unquote)",
		exp:     "{[{quasiquote: missing argument for unquote}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-splicing",
		src:     "(unquote-splicing)",
		exp:     "{[{unquote-splicing: not allowed outside quasiquote}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-splicing-macro",
		src:     ",@1",
		exp:     "{[{unquote-splicing: not allowed outside quasiquote}]}",
		withErr: true,
	},
	{name: "err-unquote-EOF", src: "`,", exp: "{[{unexpected EOF}]}", withErr: true},
	{name: "err-unquote-splicing-EOF", src: "`,@", exp: "{[{unexpected EOF}]}", withErr: true},
	{
		name:    "err-unquote-0",
		src:     "(quasiquote (unquote))",
		exp:     "{[{quasiquote: missing argument for unquote}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-2",
		src:     "(quasiquote (unquote x 7))",
		exp:     "{[{quasiquote: additional arguments (7) for unquote}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-improper",
		src:     "(quasiquote (unquote . 7 ))",
		exp:     "{[{quasiquote: improper list: (unquote . 7)}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-splicing-0",
		src:     "(quasiquote (unquote-splicing))",
		exp:     "{[{quasiquote: (quasiquote (unquote-splicing)) is not allowed}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-splicing-1",
		src:     "(quasiquote (unquote-splicing 5))",
		exp:     "{[{quasiquote: (quasiquote (unquote-splicing 5)) is not allowed}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-splicing-2",
		src:     "(quasiquote (unquote-splicing x 7))",
		exp:     "{[{quasiquote: (quasiquote (unquote-splicing x 7)) is not allowed}]}",
		withErr: true,
	},
	{
		name:    "err-unquote-splicing-improper",
		src:     "(quasiquote (unquote-splicing . 7 ))",
		exp:     "{[{quasiquote: (quasiquote (unquote-splicing . 7)) is not allowed}]}",
		withErr: true,
	},
	{name: "unquote-immediate", src: "`,x", exp: "3"},
	{name: "unquote-nested", src: "`(((,x)))", exp: "(((3)))"},
	{name: "unquote-nested-2", src: "`(((html ,x)))", exp: "(((html 3)))"},
	{name: "unquote-list-nested", src: "`((,(list x)))", exp: "(((3)))"},
	{
		name:    "err-splicing-immediate-nil",
		src:     "`,@x",
		exp:     "{[{quasiquote: (quasiquote (unquote-splicing x)) is not allowed}]}",
		withErr: true,
	},
	{
		name:    "err-splicing-immediate-list",
		src:     "`,@(list x y)",
		exp:     "{[{quasiquote: (quasiquote (unquote-splicing (list x y))) is not allowed}]}",
		withErr: true,
	},
	{name: "splicing-immediate-num", src: "`(,@x)", exp: "3"},
	{name: "splicing-immediate", src: "`(,@(list x))", exp: "(3)"},
	{name: "splicing-immediate-2", src: "`(html ,@(list x))", exp: "(html 3)"},

	{name: "nested-qq", src: "(quasiquote (quasiquote 4))", exp: "(quasiquote 4)"},
	{name: "nested-qq-macro", src: "``7", exp: "(quasiquote 7)"},
	{name: "nested-qq-unquote", src: "(quasiquote (quasiquote (unquote 9)))", exp: "(quasiquote 9)"},
	{name: "nested-qq-ok-unquote", src: "`(1 ,x `4)", exp: "(1 3 (quasiquote 4))"},

	{name: "lang-true", src: "`(html ,@(if lang1 `((@ lang ,lang1))))", exp: "(html (@ lang \"de-DE\"))"},
	{name: "lang-true-alt", src: "`(html ,@(if lang1 (list lang1)))", exp: "(html \"de-DE\")"},
	{name: "lang-false", src: "`(html ,@(if lang0 `((@ lang ,lang0))))", exp: "(html)"},

	{name: "let-in-qq", src: "`(0 ,@(let ((a 1)) `(,a)))", exp: "(0 1)"},
}

func TestQuasiQuoteExt(t *testing.T) {
	t.Parallel()
	tcsQuasiQuoteExt.Run(t)
}

var tcsQuasiQuoteExt = tTestCases{
	// Tests from https://github.com/fare/fare-quasiquote/blob/master/quasiquote-test.lisp
	{name: "simple-qq", src: "`a", exp: "a"},
	{name: "double-qq", src: "``a", exp: "(quasiquote a)"},
	{name: "simple-qq-unquote", src: "`(a ,b)", exp: "(a 11)"},
	// {name: "double-qq-unquote", src: "``(a ,b)", exp: "(list (quasiquote a) b)"},
	{name: "simple-qq-unquote-splicing", src: "`(a ,@c)", exp: "(a 22 33)"},
	{name: "simpler-qq-unquote-splicing", src: "`(,@c)", exp: "(22 33)"},
	{name: "unquote-qq-symbol", src: "`,`a", exp: "a"},
	{name: "simple-cons-unquote", src: "`(a . ,b)", exp: "(a . 11)"},
	{name: "simple-qq-unquote-and-splice", src: "`(a ,b ,@c)", exp: "(a 11 22 33)"},
	{src: "`(1 2 3)", exp: "(1 2 3)"},
	{src: "`(a ,@c . 4)", exp: "(a 22 33 . 4)"},
	{src: "`(a ,b ,@c . ,d)", exp: "(a 11 22 33 44 55)"},
	{src: "`(,@c . ,d)", exp: "(22 33 44 55)"},
	// {src: "```(,,a ,',',b)", exp: "(list (quote list) a (list (quote common-lisp:quote) '11))"},
	{
		src: "`(foobar a b ,c ,'(e f g) d ,@'(e f g) (h i j) ,@c)",
		exp: "(foobar a b (22 33) (e f g) d e f g (h i j) 22 33)",
	},
	// {src: "``(, @c)", exp: "(list @c)"},
	{src: "`(1 ,b)", exp: "(1 11)"},
	{src: "`(,'foo ,b)", exp: "(foo 11)"},

	// Tests from https://docs.racket-lang.org/reference/quasiquote.html
	{src: "(quasiquote (0 (unquote (+ 1 2)) 4))", exp: "(0 3 4)"},
	{src: "(quasiquote (0 (unquote-splicing (list 1 2)) 4))", exp: "(0 1 2 4)"},
	{
		src:     "(quasiquote (0 (unquote-splicing 1) 4))",
		exp:     "{[{append: argument 2 is not a list, but sx.Int64/1}]}",
		withErr: true,
	},
	{src: "(quasiquote (0 (unquote-splicing 1)))", exp: "(0 . 1)"},
	{src: "`(1 ,@(list 1 2) 4)", exp: "(1 1 2 4)"},
	// {src: "`(1 `,(+ 1 ,(+ 2 3)) 4)"},
}
