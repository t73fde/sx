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

func TestList(t *testing.T) {
	t.Parallel()
	tcsList.Run(t)
}

var tcsList = tTestCases{
	{
		name:    "err-cons-0",
		src:     "(cons)",
		exp:     "{[{cons: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-cons-1",
		src:     "(cons 1)",
		exp:     "{[{cons: exactly 2 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},
	{
		name:    "err-cons-3",
		src:     "(cons 1 2 3)",
		exp:     "{[{cons: exactly 2 arguments required, but 3 given: [1 2 3]}]}",
		withErr: true},
	{name: "cons-2", src: "(cons 1 2)", exp: "(1 . 2)"},

	{
		name:    "err-pair?-0",
		src:     "(pair?)",
		exp:     "{[{pair?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{name: "pair?-nil", src: "(pair? ())", exp: "False"},
	{name: "pair?-1", src: "(pair? 1)", exp: "False"},
	{name: "pair?-cons", src: "(pair? (cons 1 2))", exp: "True"},
	{name: "pair?-list", src: "(pair? (list 1 2))", exp: "True"},

	{
		name:    "err-null?-0",
		src:     "(null?)",
		exp:     "{[{null?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{name: "null?-1", src: "(null? 1)", exp: "False"},
	{name: "null?-nil", src: "(null? ())", exp: "True"},
	{name: "null?-cons", src: "(null? (cons 1 2))", exp: "False"},

	{
		name:    "err-list?-0",
		src:     "(list?)",
		exp:     "{[{list?: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{name: "list?-nil", src: "(list? ())", exp: "True"},
	{name: "list?-1", src: "(list? 1)", exp: "False"},
	{name: "list?-cons", src: "(list? (cons 1 2))", exp: "False"},
	{name: "list?-list", src: "(list? (list 1 2))", exp: "True"},

	{
		name:    "err-car-0",
		src:     "(car)",
		exp:     "{[{car: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-car-1",
		src:     "(car 1)",
		exp:     "{[{car: argument 1 is not a pair, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "car-cons", src: "(car (cons 1 2))", exp: "1"},
	{name: "car-list", src: "(car (list 1 2))", exp: "1"},
	{
		name:    "err-car-nil",
		src:     "(car ())",
		exp:     "{[{car: argument 1 is not a pair, but *sx.Pair/()}]}",
		withErr: true,
	},

	{
		name:    "err-cdr-0",
		src:     "(cdr)",
		exp:     "{[{cdr: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name: "err-cdr-1",
		src:  "(cdr 1)", exp: "{[{cdr: argument 1 is not a pair, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "cdr-cons", src: "(cdr (cons 1 2))", exp: "2"},
	{name: "cdr-list", src: "(cdr (list 1 2))", exp: "(2)"},
	{
		name:    "err-cdr-nil",
		src:     "(cdr ())",
		exp:     "{[{cdr: argument 1 is not a pair, but *sx.Pair/()}]}",
		withErr: true,
	},

	{
		name:    "err-last-0",
		src:     "(last)",
		exp:     "{[{last: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-last-1",
		src:     "(last 1)",
		exp:     "{[{last: argument 1 is not a list, but sx.Int64/1}]}",
		withErr: true},
	{
		name:    "err-last-cons",
		src:     "(last (cons 1 2))",
		exp:     "{[{last: improper list: (1 . 2)}]}",
		withErr: true,
	},
	{name: "last-list", src: "(last (list 1 2))", exp: "2"},

	{name: "list-0", src: "(list)", exp: "()"},
	{name: "list-1", src: "(list 1)", exp: "(1)"},
	{name: "list-2", src: "(list 1 2)", exp: "(1 2)"},
	{
		name:    "err-list-2-improper",
		src:     "(list 1 . 2)",
		exp:     "{[{improper list: (list 1 . 2)}]}",
		withErr: true,
	},

	{
		name:    "err-list*-0",
		src:     "(list*)",
		exp:     "{[{list*: at least 1 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{name: "list*-1", src: "(list* 1)", exp: "1"},
	{name: "list*-2", src: "(list* 1 2)", exp: "(1 . 2)"},

	{name: "append-0", src: "(append)", exp: "()"},
	{name: "append-1", src: "(append 1)", exp: "1"},
	{name: "append-3", src: "(append (list 1 2) (list 3 4 5) '(6 . 7))", exp: "(1 2 3 4 5 6 . 7)"},
	{
		name:    "err-append-improper",
		src:     "(append (cons 1 2) (list 3 4))",
		exp:     "{[{append: improper list: (1 . 2)}]}",
		withErr: true,
	},

	{
		name:    "err-reverse-0",
		src:     "(reverse)",
		exp:     "{[{reverse: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-reverse-1",
		src:     "(reverse 1)",
		exp:     "{[{reverse: argument 1 is not a list, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "reverse-nil", src: "(reverse ())", exp: "()"},
	{
		name:    "err-reverse-cons",
		src:     "(reverse (cons 1 2))",
		exp:     "{[{reverse: improper list: (1 . 2)}]}",
		withErr: true,
	},
	{name: "reverse-list", src: "(reverse (list 1 2))", exp: "(2 1)"},

	{
		name:    "err-length-0",
		src:     "(length)",
		exp:     "{[{length: exactly 1 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-length-1",
		src:     "(length 1)",
		exp:     "{[{length: argument 1 is not a list, but sx.Int64/1}]}",
		withErr: true},
	{name: "length-cons", src: "(length (cons 1 2))", exp: "1"},
	{name: "length-list-1", src: "(length (list 1))", exp: "1"},
	{name: "length-list-3", src: "(length (list 1 2 3))", exp: "3"},

	{
		name:    "err-assoc-0",
		src:     "(assoc)",
		exp:     "{[{assoc: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-assoc-1",
		src:     "(assoc ())",
		exp:     "{[{assoc: exactly 2 arguments required, but 1 given: [()]}]}",
		withErr: true,
	},
	{
		name:    "err-assoc-2-nolist",
		src:     "(assoc 1 1)",
		exp:     "{[{assoc: argument 1 is not a list, but sx.Int64/1}]}",
		withErr: true,
	},
	{
		name: "assoc-nil-alist",
		src:  "(assoc () 1)",
		exp:  "()",
	},
	{name: "assoc-none", src: "(assoc '((1 . 2) (3 . 4)) 0)", exp: "()"},
	{name: "assoc-first", src: "(assoc '((1 . 2) (3 . 4)) 1)", exp: "(1 . 2)"},
	{name: "assoc-second", src: "(assoc '((1 . 2) (3 . 4)) 3)", exp: "(3 . 4)"},
}
