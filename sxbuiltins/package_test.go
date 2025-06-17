//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import (
	"fmt"
	"testing"

	"t73f.de/r/sx"
)

func TestPackage(t *testing.T) {
	t.Parallel()
	tcsPackage.Run(t)
}

var tcsPackage = tTestCases{
	{name: "current-package", src: "(current-package)", exp: sx.CurrentPackage().String()},

	{name: "package-list", src: "(length> (package-list) 1)", exp: "T"},

	{name: "err-find-package-0",
		src:     "(find-package)",
		exp:     "{[{find-package: exactly 1 arguments required, but none given}]}",
		withErr: true},
	{name: "err-find-package-1",
		src:     "(find-package 7)",
		exp:     "{[{find-package: argument 1 is not a string, but sx.Int64/7}]}",
		withErr: true},
	{name: "err-find-package-2",
		src:     "(find-package 7 9)",
		exp:     "{[{find-package: exactly 1 arguments required, but 2 given: [7 9]}]}",
		withErr: true},
	{name: "find-package-INIT",
		src: fmt.Sprintf("(find-package \"%s\")", sx.InitName), exp: "#<package:INIT>"},

	{name: "err-package-symbols-1",
		src:     "(package-symbols 7)",
		exp:     "{[{package-symbols: argument 1 is not a package, but sx.Int64/7}]}",
		withErr: true},
	{name: "err-package-symbols-2",
		src:     "(package-symbols 7 11)",
		exp:     "{[{package-symbols: between 0 and 1 arguments required, but 2 given: [7 11]}]}",
		withErr: true},

	{name: "package-symbols-1-2-eq",
		src: "(= (length (package-symbols)) (length (package-symbols (current-package))))", exp: "T"},
}
