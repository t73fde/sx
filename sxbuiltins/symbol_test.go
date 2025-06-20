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

import (
	"testing"

	"t73f.de/r/sx"
)

func TestSymbol(t *testing.T) {
	t.Parallel()
	tcsSymbol.Run(t)
}

var tcsSymbol = tTestCases{
	{name: "err-symbol?-0",
		src:     "(symbol?)",
		exp:     "{[{symbol?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "symbol?-nil", src: "(symbol? ())", exp: "()"},
	{name: "symbol?-1", src: "(symbol? 1)", exp: "()"},
	{name: "symbol?-cons", src: "(symbol? (cons 1 2))", exp: "()"},
	{name: "symbol?-list", src: "(symbol? 'sym)", exp: "T"},

	{name: "err-keyword?-0",
		src:     "(keyword?)",
		exp:     "{[{keyword?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-keyword?-1",
		src:     "(keyword? 1)",
		exp:     "{[{keyword?: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "keyword?-yes", src: "(keyword? :keyword)", exp: "T"},
	{name: "keyword?-yes-quote", src: "(keyword? ':keyword)", exp: "T"},
	{name: "keyword?-no", src: "(keyword? 'keyword?)", exp: "()"},

	{name: "err-symbol-package-0",
		src:     "(symbol-package)",
		exp:     "{[{symbol-package: exactly 1 arguments required, but none given}]}",
		withErr: true},
	{name: "err-symbol-package-int",
		src:     "(symbol-package 3)",
		exp:     "{[{symbol-package: argument 1 is not a symbol, but sx.Int64/3}]}",
		withErr: true},
	{name: "meta-symbol-package",
		src: "(symbol-package 'symbol-package)",
		exp: sx.CurrentPackage().String()},

	{name: "err-symbol-value-0",
		src:     "(symbol-value)",
		exp:     "{[{symbol-value: exactly 1 arguments required, but none given}]}",
		withErr: true},
	{name: "err-symbol-value-1",
		src:     "(symbol-value 3)",
		exp:     "{[{symbol-value: argument 1 is not a symbol, but sx.Int64/3}]}",
		withErr: true},
	{name: "err-symbol-value-2",
		src:     "(symbol-value 3 7)",
		exp:     "{[{symbol-value: exactly 1 arguments required, but 2 given: [3 7]}]}",
		withErr: true},
	{name: "symbol-value-T", src: "(symbol-value 'T)", exp: "T"},
	{name: "symbol-value-undef", src: "(defined? (symbol-value 'abc))", exp: "()"},

	{name: "err-set-symbol-value-0",
		src:     "(set-symbol-value)",
		exp:     "{[{set-symbol-value: exactly 2 arguments required, but none given}]}",
		withErr: true},
	{name: "err-set-symbol-value-1",
		src:     "(set-symbol-value 'a)",
		exp:     "{[{set-symbol-value: exactly 2 arguments required, but 1 given: [a]}]}",
		withErr: true},
	{name: "err-set-symbol-value-2",
		src:     "(set-symbol-value 3 7)",
		exp:     "{[{set-symbol-value: argument 1 is not a symbol, but sx.Int64/3}]}",
		withErr: true},
	{name: "err-set-symbol-value-3",
		src:     "(set-symbol-value 'a 'b 'c)",
		exp:     "{[{set-symbol-value: exactly 2 arguments required, but 3 given: [a b c]}]}",
		withErr: true},
	{name: "set-symbol-value", src: "(set-symbol-value 'abc 123)(symbol-value 'abc)", exp: "123 123"},

	{name: "err-freeze-symbol-value-0",
		src:     "(freeze-symbol-value)",
		exp:     "{[{freeze-symbol-value: exactly 1 arguments required, but none given}]}",
		withErr: true},
	{name: "err-freeze-symbol-value-1",
		src:     "(freeze-symbol-value 3)",
		exp:     "{[{freeze-symbol-value: argument 1 is not a symbol, but sx.Int64/3}]}",
		withErr: true},
	{name: "err-freeze-symbol-value-2",
		src:     "(freeze-symbol-value 3 7)",
		exp:     "{[{freeze-symbol-value: exactly 1 arguments required, but 2 given: [3 7]}]}",
		withErr: true},
	// {name: "freeze-symbol-value-undef",
	// 	src:     "(freeze-symbol-value 'undef-frozen)(set-symbol-value 'undef-frozen 19)",
	// 	exp:     "{[{symbol undef-frozen is frozen, without bound value}]}",
	// 	withErr: true},
	// {name: "freeze-symbol-value-def",
	// 	src:     "(set-symbol-value 'def-frozen 17)(freeze-symbol-value 'def-frozen)(set-symbol-value 'def-frozen 19)",
	// 	exp:     "{[{symbol def-frozen is frozen, with value: 17}]}",
	// 	withErr: true},
	{name: "freeze-symbol-value", src: "(freeze-symbol-value 'freeze-value)", exp: "()"},
	// Following depends on already frozen symbol freeze-value
	{name: "freeze-set",
		src:     "(set-symbol-value 'freeze-value 19)",
		exp:     "{[{set-symbol-value: symbol freeze-value is frozen, without bound value}]}",
		withErr: true},

	{name: "err-frozen-symbol-value-0",
		src:     "(frozen-symbol-value)",
		exp:     "{[{frozen-symbol-value: exactly 1 arguments required, but none given}]}",
		withErr: true},
	{name: "err-frozen-symbol-value-1",
		src:     "(frozen-symbol-value 3)",
		exp:     "{[{frozen-symbol-value: argument 1 is not a symbol, but sx.Int64/3}]}",
		withErr: true},
	{name: "err-frozen-symbol-value-2",
		src:     "(frozen-symbol-value 3 7)",
		exp:     "{[{frozen-symbol-value: exactly 1 arguments required, but 2 given: [3 7]}]}",
		withErr: true},
	{name: "frozen-symbol-value-T", src: "(frozen-symbol-value 'T)", exp: "T"},
	{name: "frozen-symbol-value-unfrozen", src: "(frozen-symbol-value 'unfrozen)", exp: "()"},
}
