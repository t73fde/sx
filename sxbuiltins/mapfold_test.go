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

func TestMapFold(t *testing.T) {
	t.Parallel()
	tcsMapFold.Run(t)
}

var tcsMapFold = tTestCases{
	{
		name:    "err-map-0",
		src:     "(map)",
		exp:     "{[{map: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-map-nofn",
		src:     "(map () ())",
		exp:     "{[{map: argument 1 is not a function, but *sx.Pair/()}]}",
		withErr: true,
	},
	{
		name:    "err-map-nolist",
		src:     "(map + 1)",
		exp:     "{[{map: argument 2 is not a list, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "map-nil", src: "(map (lambda (x) (if x () 1)) ())", exp: "()"},
	{
		name: "map-list",
		src:  "(map (lambda (x) (if x () 1)) (list 1 () \"\" \"1\" 3))",
		exp:  "(() 1 1 () ())",
	},
	{name: "map-cons", src: "(map (lambda (x) (if x () 1)) (cons 1 \"\"))", exp: "(() . 1)"},
	{name: "map-list*", src: "(map (lambda (x) (if x () 1)) (list* () 1 \"\"))", exp: "(1 () . 1)"},
	{name: "map-bcall", src: "(map list '(1 2 3 4))", exp: "((1) (2) (3) (4))"},

	{
		name:    "err-apply-0",
		src:     "(apply)",
		exp:     "{[{apply: exactly 2 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{name: "apply-nil", src: "(apply + ())", exp: "0"},
	{
		name:    "err-apply-cons",
		src:     "(apply + (cons 1 2))",
		exp:     "{[{apply: improper list: (1 . 2)}]}",
		withErr: true,
	},
	{name: "apply-bcall", src: "(apply + '(1 2 3 4))", exp: "10"},

	{
		name:    "err-fold-0",
		src:     "(fold)",
		exp:     "{[{fold: exactly 3 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-fold-nofn",
		src:     "(fold 0 1 (list 1))",
		exp:     "{[{fold: argument 1 is not a function, but sx.Int64/0}]}",
		withErr: true,
	},
	{name: "fold", src: "(fold + 0 (list 1 2 3 4 5 6 7 8 9))", exp: "45"},
	{name: "fold-nil", src: "(fold + 7 ())", exp: "7"},
	{name: "fold-1", src: "(fold + 0 (list 1))", exp: "1"},
	{name: "fold-3", src: "(fold + 0 (list 1 2 3))", exp: "6"},
	{
		name:    "err-fold-improper",
		src:     "(fold + 0 (list* 1 2 3))",
		exp:     "{[{fold: improper list: (1 2 . 3)}]}",
		withErr: true,
	},
	{name: "fold-3-cons", src: "(fold cons () (list 1 2 3))", exp: "(3 2 1)"},

	{
		name:    "err-fold-reverse-0",
		src:     "(fold-reverse)",
		exp:     "{[{fold-reverse: exactly 3 arguments required, but 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-fold-reverse-nofn",
		src:     "(fold-reverse 0 1 (list 1))",
		exp:     "{[{fold-reverse: argument 1 is not a function, but sx.Int64/0}]}",
		withErr: true,
	},
	{name: "fold-reverse", src: "(fold-reverse + 0 (list 1 2 3 4 5 6 7 8 9))", exp: "45"},
	{name: "fold-reverse-nil", src: "(fold-reverse + 7 ())", exp: "7"},
	{name: "fold-reverse-1", src: "(fold-reverse + 0 (list 1))", exp: "1"},
	{name: "fold-reverse-3", src: "(fold-reverse + 0 (list 1 2 3))", exp: "6"},
	{
		name:    "err-fold-reverse-improper",
		src:     "(fold-reverse + 0 (list* 1 2 3))",
		exp:     "{[{fold-reverse: improper list: (1 2 . 3)}]}",
		withErr: true,
	},
	{name: "fold-reverse-3-cons", src: "(fold-reverse cons () (list 1 2 3))", exp: "(1 2 3)"},
}
