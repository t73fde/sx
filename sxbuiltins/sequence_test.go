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

func TestSequence(t *testing.T) {
	t.Parallel()
	tcsSequence.Run(t)
}

var tcsSequence = tTestCases{
	{
		name:    "err-length-0",
		src:     "(length)",
		exp:     "{[{length: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{
		name:    "err-length-1",
		src:     "(length 1)",
		exp:     "{[{length: argument 1 is not a sequence, but sx.Int64/1}]}",
		withErr: true},
	{name: "length-nil", src: "(length ())", exp: "0"},
	{name: "length-cons", src: "(length (cons 1 2))", exp: "1"},
	{name: "length-list-1", src: "(length (list 1))", exp: "1"},
	{name: "length-list-3", src: "(length (list 1 2 3))", exp: "3"},
	{name: "length-vector-0", src: "(length (vector))", exp: "0"},
	{name: "length-vector-1", src: "(length (vector 1))", exp: "1"},
	{name: "length-vector-3", src: "(length (vector 1 2 3))", exp: "3"},

	{name: "err-nth-0", src: "(nth)", exp: "{[{nth: exactly 2 arguments required, but none given}]}", withErr: true},
	{name: "err-nth-1", src: "(nth 1)", exp: "{[{nth: exactly 2 arguments required, but 1 given: [1]}]}", withErr: true},
	{name: "err-nth-2", src: "(nth 1 2)", exp: "{[{nth: argument 1 is not a sequence, but sx.Int64/1}]}", withErr: true},
	{name: "err-nth-nil-2", src: "(nth () 2)", exp: "{[{nth: sequence is nil}]}", withErr: true},
	{name: "err-nth-lst-2", src: "(nth '(1) ())", exp: "{[{nth: argument 2 is not a number, but *sx.Pair/()}]}", withErr: true},
	{name: "err-nth-lst-range", src: "(nth '(1) 1)", exp: "{[{nth: index too large: 1 for (1)}]}", withErr: true},
	{name: "err-nth-lst-neg", src: "(nth '(1) -1)", exp: "{[{nth: negative index -1}]}", withErr: true},
	{name: "err-nth-vector-2", src: "(nth (vector) 2)", exp: "{[{nth: sequence is nil}]}", withErr: true},
	{name: "err-nth-vector-range", src: "(nth (vector 1) 1)", exp: "{[{nth: index out of range: 1 (max: 0)}]}", withErr: true},
	{name: "err-nth-vector-neg", src: "(nth (vector 1) -1)", exp: "{[{nth: index out of range: -1 (max: 0)}]}", withErr: true},
	{name: "nth-list-1", src: "(nth '(1) 0)", exp: "1"},
	{name: "nth-list-3", src: "(nth '(1 2 3 4 5 6) 3)", exp: "4"},
	{name: "nth-vector-1", src: "(nth (vector 1) 0)", exp: "1"},
	{name: "nth-vector-4", src: "(nth (vector 1 2 3 4 5 6) 4)", exp: "5"},
}
