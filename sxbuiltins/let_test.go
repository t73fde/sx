//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestLet(t *testing.T) {
	t.Parallel()
	tcsLet.Run(t)
}

var tcsLet = tTestCases{
	{name: "err-let-0", src: "(let)", exp: "{[{let: binding spec and body missing}]}", withErr: true},
	{name: "err-let-num",
		src:     "(let 1)",
		exp:     "{[{let: bindings must be a list, but is sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-let-improper", src: "(let () . 1)", exp: "{[{let: improper list: (() . 1)}]}", withErr: true},
	{name: "err-let-nobinding-list", src: "(let (a) 1)", exp: "{[{let: single binding must be a list, but is *sx.Symbol/a}]}", withErr: true},
	{name: "err-let-nobinding", src: "(let ((a)) 1)", exp: "{[{let: binding missing for symbol a}]}", withErr: true},
	{name: "err-let-improper-binding",
		src:     "(let ((a . 1)) a)",
		exp:     "{[{let: improper list: (a . 1)}]}",
		withErr: true,
	},
	{name: "err-let-improper-binding-2",
		src:     "(let ((a 1 . 2)) a)",
		exp:     "{[{let: too many bindings for symbol a: sx.Int64/2}]}",
		withErr: true,
	},
	{name: "let-nil-1", src: "(let () 1)", exp: "1"},
	{name: "let-a-b", src: "(let ((a 1) (b 2)) (lambda () a) b)", exp: "2"},
	{name: "let-nested-0", src: "(let ((a 1)) (let ((a 2)) a))", exp: "2"},
	{name: "let-no-nested", src: "(let ((a 1)) (let ((a 2)) (let ((a 3) (b a)) a)) a)", exp: "1"},
	{name: "err-let-double-sym",
		src:     "(let ((a 1) (a 2)) a)",
		exp:     "{[{let: symbol a already defined}]}",
		withErr: true,
	},
	{name: "err-let-improper-body",
		src:     "(let ((a 1)) . a)",
		exp:     "{[{let: improper list: (((a 1)) . a)}]}",
		withErr: true,
	},
	{name: "err-let-parse-body",
		src:     "(let ((a 1)) (set!))",
		exp:     "{[{set!: need at least two arguments}]}",
		withErr: true,
	},

	// (let* ...)

	{name: "err-let*-0", src: "(let*)", exp: "{[{let*: binding spec and body missing}]}", withErr: true},
	{name: "err-let*-num",
		src:     "(let* 1)",
		exp:     "{[{let*: bindings must be a list, but is sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-let*-improper", src: "(let* () . 1)", exp: "{[{let*: improper list: (() . 1)}]}", withErr: true},
	{name: "err-let*-nobinding-list", src: "(let* (a) 1)", exp: "{[{let*: single binding must be a list, but is *sx.Symbol/a}]}", withErr: true},
	{name: "err-let*-nobinding", src: "(let* ((a)) 1)", exp: "{[{let*: binding missing for symbol a}]}", withErr: true},
	{name: "err-let*-improper-binding",
		src:     "(let* ((a . 1)) a)",
		exp:     "{[{let*: improper list: (a . 1)}]}",
		withErr: true,
	},
	{name: "err-let*-improper-binding-2",
		src:     "(let* ((a 1 . 2)) a)",
		exp:     "{[{let*: too many bindings for symbol a: sx.Int64/2}]}",
		withErr: true,
	},
	{name: "let*-nil-1", src: "(let* () 1)", exp: "1"},
	{name: "let*-a-b", src: "(let* ((a 1) (b 2)) (lambda () a) b)", exp: "2"},
	{name: "let*-nested-0", src: "(let* ((a 1)) (let ((a 2)) a))", exp: "2"},
	{name: "let*-no-nested", src: "(let* ((a 1)) (let ((a 2)) a) a)", exp: "1"},
	{name: "let*-double-sym", src: "(let* ((a 1) (a 2)) a)", exp: "2"},
	{name: "let*-multieval", src: "(let* ((a 3) (b (+ a 4)) (a (+ a 5))) (+ a b))", exp: "15"},
}
