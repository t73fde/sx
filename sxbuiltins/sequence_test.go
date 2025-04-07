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

	{name: "err-length<-0", src: "(length<)", exp: "{[{length<: exactly 2 arguments required, but none given}]}", withErr: true},
	{name: "err-length<-1", src: "(length< 1)", exp: "{[{length<: exactly 2 arguments required, but 1 given: [1]}]}", withErr: true},
	{name: "err-length<-2", src: "(length< 1 2)", exp: "{[{length<: argument 1 is not a sequence, but sx.Int64/1}]}", withErr: true},
	{name: "err-length<-nil-nil", src: "(length< () ())", exp: "{[{length<: argument 2 is not a number, but *sx.Pair/()}]}", withErr: true},
	{name: "length<-nil-0", src: "(length< () 0)", exp: "()"},
	{name: "length<-nil-1", src: "(length< () 1)", exp: "T"},
	{name: "length<-list-10-3", src: "(length< '(1 2 3 4 5 6 7 8 9 10) 3)", exp: "()"},
	{name: "length<-list-10-10", src: "(length< '(1 2 3 4 5 6 7 8 9 10) 10)", exp: "()"},
	{name: "length<-list-10-11", src: "(length< '(1 2 3 4 5 6 7 8 9 10) 11)", exp: "T"},
	{name: "length<-nil-vector-0", src: "(length< (vector) 0)", exp: "()"},
	{name: "length<-nil-vector-1", src: "(length< (vector) 1)", exp: "T"},

	{name: "err-length>-0", src: "(length>)", exp: "{[{length>: exactly 2 arguments required, but none given}]}", withErr: true},
	{name: "err-length>-1", src: "(length> 1)", exp: "{[{length>: exactly 2 arguments required, but 1 given: [1]}]}", withErr: true},
	{name: "err-length>-2", src: "(length> 1 2)", exp: "{[{length>: argument 1 is not a sequence, but sx.Int64/1}]}", withErr: true},
	{name: "err-length>-nil-nil", src: "(length> () ())", exp: "{[{length>: argument 2 is not a number, but *sx.Pair/()}]}", withErr: true},
	{name: "length>-nil-0", src: "(length> () -1)", exp: "T"},
	{name: "length>-nil-1", src: "(length> () 1)", exp: "()"},
	{name: "length>-list-10-3", src: "(length> '(1 2 3 4 5 6 7 8 9 10) 3)", exp: "T"},
	{name: "length>-list-10-10", src: "(length> '(1 2 3 4 5 6 7 8 9 10) 9)", exp: "T"},
	{name: "length>-list-10-11", src: "(length> '(1 2 3 4 5 6 7 8 9 10) 10)", exp: "()"},
	{name: "length>-nil-vector-0", src: "(length> (vector) -1)", exp: "T"},
	{name: "length>-nil-vector-1", src: "(length> (vector) 1)", exp: "()"},

	{name: "err-length=-0", src: "(length=)", exp: "{[{length=: exactly 2 arguments required, but none given}]}", withErr: true},
	{name: "err-length=-1", src: "(length= 1)", exp: "{[{length=: exactly 2 arguments required, but 1 given: [1]}]}", withErr: true},
	{name: "err-length=-2", src: "(length= 1 2)", exp: "{[{length=: argument 1 is not a sequence, but sx.Int64/1}]}", withErr: true},
	{name: "err-length=-nil-nil", src: "(length= () ())", exp: "{[{length=: argument 2 is not a number, but *sx.Pair/()}]}", withErr: true},
	{name: "length=-nil-0", src: "(length= () 0)", exp: "T"},
	{name: "length=-nil-1", src: "(length= () 1)", exp: "()"},
	{name: "length=-list-10-3", src: "(length= '(1 2 3 4 5 6 7 8 9 10) 3)", exp: "()"},
	{name: "length=-list-10-10", src: "(length= '(1 2 3 4 5 6 7 8 9 10) 10)", exp: "T"},
	{name: "length=-list-10-11", src: "(length= '(1 2 3 4 5 6 7 8 9 10) 11)", exp: "()"},
	{name: "length=-nil-vector-0", src: "(length= (vector) 0)", exp: "T"},
	{name: "length=-nil-vector-1", src: "(length= (vector) 1)", exp: "()"},

	{name: "err-nth-0", src: "(nth)", exp: "{[{nth: exactly 2 arguments required, but none given}]}", withErr: true},
	{name: "err-nth-1", src: "(nth 1)", exp: "{[{nth: exactly 2 arguments required, but 1 given: [1]}]}", withErr: true},
	{name: "err-nth-2", src: "(nth 1 2)", exp: "{[{nth: argument 1 is not a sequence, but sx.Int64/1}]}", withErr: true},
	{name: "err-nth-nil-2", src: "(nth () 2)", exp: "{[{nth: index too large: 2 for ()}]}", withErr: true},
	{name: "err-nth-lst-2", src: "(nth '(1) ())", exp: "{[{nth: argument 2 is not a number, but *sx.Pair/()}]}", withErr: true},
	{name: "err-nth-lst-range", src: "(nth '(1) 1)", exp: "{[{nth: index too large: 1 for (1)}]}", withErr: true},
	{name: "err-nth-lst-neg", src: "(nth '(1) -1)", exp: "{[{nth: negative index -1}]}", withErr: true},
	{name: "err-nth-vector-2", src: "(nth (vector) 2)", exp: "{[{nth: index too large: 2 for ()}]}", withErr: true},
	{name: "err-nth-vector-range", src: "(nth (vector 1) 1)", exp: "{[{nth: index out of range: 1 (max: 0)}]}", withErr: true},
	{name: "err-nth-vector-neg", src: "(nth (vector 1) -1)", exp: "{[{nth: index out of range: -1 (max: 0)}]}", withErr: true},
	{name: "nth-list-1", src: "(nth '(1) 0)", exp: "1"},
	{name: "nth-list-3", src: "(nth '(1 2 3 4 5 6) 3)", exp: "4"},
	{name: "nth-vector-1", src: "(nth (vector 1) 0)", exp: "1"},
	{name: "nth-vector-4", src: "(nth (vector 1 2 3 4 5 6) 4)", exp: "5"},

	{name: "err-seq->list-0", src: "(seq->list)", exp: "{[{seq->list: exactly 1 arguments required, but none given}]}", withErr: true},
	{name: "err-seq->list-1", src: "(seq->list 1)", exp: "{[{seq->list: argument 1 is not a sequence, but sx.Int64/1}]}", withErr: true},
	{name: "seq->list-nil", src: "(seq->list ())", exp: "()"},
	{name: "seq->list-some", src: "(seq->list (list 1 2 3 4))", exp: "(1 2 3 4)"},
	{name: "seq->list-nil-vector", src: "(seq->list (vector))", exp: "()"},
	{name: "seq->list-some-vector", src: "(seq->list (vector 1 2 3))", exp: "(1 2 3)"},
}
