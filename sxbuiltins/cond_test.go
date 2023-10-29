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

func TestCond(t *testing.T) {
	t.Parallel()
	tcsCond.Run(t)
}

var tcsCond = tTestCases{
	{name: "cond-0", src: "(cond)", exp: "()"},
	{name: "cond-nil", src: "(cond ())", exp: "()"},
	{name: "err-cond-1", src: "(cond 1)", exp: "{[{cond: clause must be a list, but got sx.Int64/1}]}", withErr: true},
	{name: "err-cond-improper", src: "(cond (2 3) . 1)", exp: "{[{cond: improper list: ((2 3) . 1)}]}", withErr: true},
	{name: "cond-if", src: "(cond ((= 9 9) 2 3))", exp: "3"},
	{name: "cond-if-false", src: "(cond ((= 0 9) 2 3))", exp: "()"},
	{name: "cond-if-else", src: "(cond ((= 0 1) 2 3) (5 6))", exp: "6"},
	{name: "cond-remove-nil", src: "(cond ((pp 2) 3) (() 4) ((pp 5) 6))", exp: "()"},
}
