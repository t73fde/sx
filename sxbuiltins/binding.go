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
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

var CurrentBinding = sxeval.Builtin{
	Name:     "current-binding",
	MinArity: 0,
	MaxArity: 0,
	TestPure: nil,
	Fn:       func(env *sxeval.Environment, _ sx.Vector) (sx.Object, error) { return env.Binding(), nil },
}

var ParentBinding = sxeval.Builtin{
	Name:     "parent-binding",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		bind, err := GetBinding(args, 0)
		if err != nil {
			return nil, err
		}
		if parent := bind.Parent(); parent != nil {
			return parent, nil
		}
		return sx.MakeUndefined(), nil
	},
}

var Bindings = sxeval.Builtin{
	Name:     "bindings",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		bind, err := GetBinding(args, 0)
		if err != nil {
			return nil, err
		}
		return bind.Bindings(), nil
	},
}

// BoundP returns true, if the given symbol is bound in the current environment.
var BoundP = sxeval.Builtin{
	Name:     "bound?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		sym, err := GetSymbol(args, 0)
		if err != nil {
			return nil, err
		}
		_, found := env.Resolve(sym)
		return sx.MakeBoolean(found), nil
	},
}

// BindingLookup returns the symbol's bound value in the given
// binding, or Undefined if there is no such value.
var BindingLookup = sxeval.Builtin{
	Name:     "binding-lookup",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		sym, bind, err := envGetSymBinding(env, args)
		if err != nil {
			return nil, err
		}
		if obj, found := bind.Lookup(sym); found {
			return obj, nil
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
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		sym, bind, err := envGetSymBinding(env, args)
		if err != nil {
			return nil, err
		}
		if obj, found := bind.Resolve(sym); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
}

func envGetSymBinding(env *sxeval.Environment, args sx.Vector) (sx.Symbol, *sxeval.Binding, error) {
	sym, err := GetSymbol(args, 0)
	if err != nil {
		return "", nil, err
	}
	if len(args) == 1 {
		return sym, env.Binding(), err
	}
	bind, err := GetBinding(args, 1)
	return sym, bind, err
}
