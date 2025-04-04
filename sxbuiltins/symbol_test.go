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

func TestSymbol(t *testing.T) {
	t.Parallel()
	tcsSymbol.Run(t)
}

var tcsSymbol = tTestCases{
	{
		name:    "err-symbol?-0",
		src:     "(symbol?)",
		exp:     "{[{symbol?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "symbol?-nil", src: "(symbol? ())", exp: "()"},
	{name: "symbol?-1", src: "(symbol? 1)", exp: "()"},
	{name: "symbol?-cons", src: "(symbol? (cons 1 2))", exp: "()"},
	{name: "symbol?-list", src: "(symbol? 'sym)", exp: "T"},
}
