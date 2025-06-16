//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

// Provides some special/builtin functions to work with bindings.

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// SymbolBoundP returns true, if the given symbol is bound in the current environment.
var SymbolBoundP = sxeval.Builtin{
	Name:     "symbol-bound?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, arg sx.Object, frame *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		_, found := env.Resolve(sym, frame)
		return sx.MakeBoolean(found), nil
	},
}

// ResolveSymbol returns the symbol's bound value in the given frame
// and all its parent frame and in the global binding, or Undefined if there is no such value.
var ResolveSymbol = sxeval.Builtin{
	Name:     "resolve-symbol",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, arg sx.Object, frame *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		if obj, found := env.Resolve(sym, frame); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
	Fn: func(env *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(args[0], 0)
		if err != nil {
			return nil, err
		}

		frame, err := GetFrame(args[1], 1)
		if err != nil {
			if !sx.IsNil(args[1]) {
				return nil, err
			}
			frame = nil
		}
		if obj, found := env.Resolve(sym, frame); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
}
