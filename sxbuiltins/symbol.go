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
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		_, isSymbol := sx.GetSymbol(arg)
		return sx.MakeBoolean(isSymbol), nil
	},
}

// KeywordP returns true if the symbol is a keyword.
var KeywordP = sxeval.Builtin{
	Name:     "keyword?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		return sx.MakeBoolean(sym.IsKeyword()), nil
	},
}

// SymbolPackage returns the package that defined the symbol.
var SymbolPackage = sxeval.Builtin{
	Name:     "symbol-package",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		return sym.Package(), nil
	},
}

// SymbolValue returns the bound value of the symbol, or the undefined value.
var SymbolValue = sxeval.Builtin{
	Name:     "symbol-value",
	MinArity: 1,
	MaxArity: 1,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		if val, found := sym.Bound(); found {
			return val, nil
		}
		return sx.MakeUndefined(), nil
	},
}

// SetSymbolValue sets the bound value of the symbol.
//
// TODO: deprecated, since in the future, this function will probably be
// implemented by (define sym val) and (set! sym val).
var SetSymbolValue = sxeval.Builtin{
	Name:     "set-symbol-value",
	MinArity: 2,
	MaxArity: 2,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(args[0], 0)
		if err != nil {
			return nil, err
		}
		val := args[1]
		if err = sym.Bind(val); err != nil {
			return nil, err
		}
		return val, nil
	},
}

// FreezeSymbolValue forbits future update of the symbol value.
var FreezeSymbolValue = sxeval.Builtin{
	Name:     "freeze-symbol-value",
	MinArity: 1,
	MaxArity: 1,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		sym.Freeze()
		return sx.Nil(), nil
	},
}

// FrozenSymbolValue returns an indication whether the symbol value is frozen or not.
var FrozenSymbolValue = sxeval.Builtin{
	Name:     "frozen-symbol-value",
	MinArity: 1,
	MaxArity: 1,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		return sx.MakeBoolean(sym.IsFrozen()), nil
	},
}
