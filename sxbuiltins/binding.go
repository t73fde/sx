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

// Provides some special/builtin functions to work with environments.

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// CurrentBinding returns the current binding.
var CurrentBinding = sxeval.Builtin{
	Name:     "current-binding", // TODO: current-frame
	MinArity: 0,
	MaxArity: 0,
	TestPure: nil,
	Fn0: func(_ *sxeval.Environment, frame *sxeval.Frame) (sx.Object, error) {
		return frame, nil
	},
}

// ParentBinding returns the parent binding of the given binding. For the
// top-most binding, an undefined value is returned.
var ParentBinding = sxeval.Builtin{
	Name:     "parent-binding", // TODO
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		frame, err := GetFrame(arg, 0)
		if err != nil {
			return nil, err
		}
		return frame.Parent(), nil
	},
}

// Bindings returns the given bindings as an association list.
var Bindings = sxeval.Builtin{
	Name:     "bindings", // TODO: frame
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		frame, err := GetFrame(arg, 0)
		if err != nil {
			return nil, err
		}
		return frame.Bindings(), nil
	},
}

// BoundP returns true, if the given symbol is bound in the current environment.
var BoundP = sxeval.Builtin{
	Name:     "bound?",
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

// BindingLookup returns the symbol's bound value in the given
// binding, or Undefined if there is no such value.
var BindingLookup = sxeval.Builtin{
	Name:     "binding-lookup", // TODO
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, frame *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		if obj, found := frame.Lookup(sym); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(args[0], 0)
		if err != nil {
			return nil, err
		}
		if arg1 := args[1]; !sx.IsNil(arg1) {
			frame, err2 := GetFrame(arg1, 1)
			if err2 != nil {
				return nil, err2
			}
			if obj, found := frame.Lookup(sym); found {
				return obj, nil
			}
		}
		return sx.MakeUndefined(), nil
	},
}

// BindingResolve returns the symbol's bound value in the given environment
// and all its parent environment, or Undefined if there is no such value.
var BindingResolve = sxeval.Builtin{
	Name:     "binding-resolve",
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
			return nil, err
		}
		if obj, found := env.Resolve(sym, frame); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
}
