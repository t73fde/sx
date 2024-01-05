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

func TestMacro(t *testing.T) {
	t.Parallel()
	tcsMacro.Run(t)
}

var tcsMacro = tTestCases{
	{name: "err-defmacro-0", src: "(defmacro)", exp: "{[{defmacro: no arguments given}]}", withErr: true},
	{name: "err-defmacro-a", src: "(defmacro a)", exp: "{[{defmacro: parameter spec and body missing}]}", withErr: true},
	{
		name: "defmacro-inc",
		src:  "(defmacro inc (var) `(set! ,var (+ ,var 1))) (defvar a 0) (inc a) (inc a)",
		exp:  "#<macro:inc> 0 1 2",
	},
	{
		name: "defmacro-inc-expand0",
		src:  "(defmacro inc (var) `(set! ,var (+ ,var 1))) (macroexpand-0 '(inc a))",
		exp:  "#<macro:inc> (set! a (+ a 1))",
	},
	{
		name: "defmacro-eq-same",
		src:  "(defmacro inc1 (var) `(set! ,var (+ ,var 1))) (== inc1 inc1) (= inc1 inc1)",
		exp:  "#<macro:inc1> 1 1",
	},
}
