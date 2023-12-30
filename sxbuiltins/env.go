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

var CurrentEnv = sxeval.Builtin{
	Name:     "current-environment",
	MinArity: 0,
	MaxArity: 0,
	TestPure: nil,
	Fn: func(frame *sxeval.Frame, _ []sx.Object) (sx.Object, error) {
		return frame.Binding(), nil
	},
}

var ParentEnv = sxeval.Builtin{
	Name:     "parent-environment",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		env, err := GetBinding(args, 0)
		if err != nil {
			return nil, err
		}
		if parent := env.Parent(); parent != nil {
			return parent, nil
		}
		return sx.MakeUndefined(), nil
	},
}

var EnvBindings = sxeval.Builtin{
	Name:     "environment-bindings",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		env, err := GetBinding(args, 0)
		if err != nil {
			return nil, err
		}
		return env.Bindings(), nil
	},
}

// BoundP returns true, if the given symbol is bound in the current environment.
var BoundP = sxeval.Builtin{
	Name:     "bound?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn: func(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		sym, err := GetSymbol(args, 0)
		if err != nil {
			return nil, err
		}
		_, found := frame.Resolve(sym)
		return sx.MakeBoolean(found), nil
	},
}

// EnvLookup returns the symbol's binding value in the given
// environment, or Undefined if there is no such value.
var EnvLookup = sxeval.Builtin{
	Name:     "environment-lookup",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		sym, env, err := envGetSymEnv(frame, args)
		if err != nil {
			return nil, err
		}
		if obj, found := env.Lookup(sym); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
}

// EnvResolve returns the symbol's binding value in the given environment
// and all its parent environment, or Undefined if there is no such value.
var EnvResolve = sxeval.Builtin{
	Name:     "environment-resolve",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		sym, env, err := envGetSymEnv(frame, args)
		if err != nil {
			return nil, err
		}
		if obj, found := sxeval.Resolve(env, sym); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
}

func envGetSymEnv(frame *sxeval.Frame, args []sx.Object) (*sx.Symbol, *sxeval.Binding, error) {
	sym, err := GetSymbol(args, 0)
	if err != nil {
		return nil, nil, err
	}
	if len(args) == 1 {
		return sym, frame.Binding(), err
	}
	env, err := GetBinding(args, 1)
	return sym, env, err
}
