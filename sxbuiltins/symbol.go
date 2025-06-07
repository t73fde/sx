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

package sxbuiltins

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// SymbolP returns true if the argument is a symbol.
var SymbolP = sxeval.Builtin{
	Name:     "symbol?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		_, isSymbol := sx.GetSymbol(arg)
		return sx.MakeBoolean(isSymbol), nil
	},
}

// SymbolPackage returns the package that defined the symbol.
var SymbolPackage = sxeval.Builtin{
	Name:     "symbol-package",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		return sym.Package(), nil
	},
}
