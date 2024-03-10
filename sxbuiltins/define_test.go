//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestDefine(t *testing.T) {
	t.Parallel()
	tcsDefine.Run(t)
}

var tcsDefine = tTestCases{
	{
		name:    "err-defvar-0",
		src:     "(defvar)",
		exp:     "{[{defvar: need at least two arguments}]}",
		withErr: true,
	},
	{
		name:    "err-defvar-1",
		src:     "(defvar 1)",
		exp:     "{[{defvar: argument 1 must be a symbol, but is: sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "err-defvar-1a",
		src:     "(defvar a)",
		exp:     "{[{defvar: argument 2 missing}]}",
		withErr: true,
	},
	{
		name:    "err-defvar-a-1",
		src:     "(defvar a . 1)",
		exp:     "{[{defvar: argument 2 must be a proper list}]}",
		withErr: true,
	},
	{name: "defvar-a-1", src: "(defvar a 1)", exp: "1"},
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
		exp:     "{[{set!: argument 2 must be a proper list, but is: sx.Int64/1}]}",
		withErr: true,
	},
	{
		name:    "set!-unknown-1",
		src:     "(set! unknown 1)",
		exp:     `{[{symbol "unknown" not bound in "set!-unknown-1"->"vars"->"root"}]}`,
		withErr: true,
	},
	{name: "define-set", src: "(defvar a 1) (set! a 17)", exp: "1 17"},
}
