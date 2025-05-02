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
	Name:     "current-binding",
	MinArity: 0,
	MaxArity: 0,
	TestPure: nil,
	Fn0: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		env.Push(bind)
		return nil
	},
}

// ParentBinding returns the parent binding of the given binding. For the
// top-most binding, an undefined value is returned.
var ParentBinding = sxeval.Builtin{
	Name:     "parent-binding",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		bind, err := GetBinding(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		if parent := bind.Parent(); parent != nil {
			env.Set(parent)
			return nil
		}
		env.Set(sx.MakeUndefined())
		return nil
	},
}

// Bindings returns the given bindings as an association list.
var Bindings = sxeval.Builtin{
	Name:     "bindings",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		bind, err := GetBinding(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(bind.Bindings())
		return nil
	},
}

// BoundP returns true, if the given symbol is bound in the current environment.
var BoundP = sxeval.Builtin{
	Name:     "bound?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		sym, err := GetSymbol(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		_, found := bind.Resolve(sym)
		env.Set(sx.MakeBoolean(found))
		return nil
	},
}

// BindingLookup returns the symbol's bound value in the given
// binding, or Undefined if there is no such value.
var BindingLookup = sxeval.Builtin{
	Name:     "binding-lookup",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		sym, err := GetSymbol(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		if obj, found := bind.Lookup(sym); found {
			env.Set(obj)
			return nil
		}
		env.Set(sx.MakeUndefined())
		return nil
	},
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		sym, err := GetSymbol(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		bind, err := GetBinding(arg1, 1)
		if err != nil {
			env.Kill(1)
			return err
		}
		if obj, found := bind.Lookup(sym); found {
			env.Set(obj)
			return nil
		}
		env.Set(sx.MakeUndefined())
		return nil
	},
}

// BindingResolve returns the symbol's bound value in the given environment
// and all its parent environment, or Undefined if there is no such value.
var BindingResolve = sxeval.Builtin{
	Name:     "binding-resolve",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		sym, err := GetSymbol(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		if obj, found := bind.Resolve(sym); found {
			env.Set(obj)
			return nil
		}
		env.Set(sx.MakeUndefined())
		return nil
	},
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		sym, err := GetSymbol(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		bind, err := GetBinding(arg1, 1)
		if err != nil {
			env.Kill(1)
			return err
		}
		if obj, found := bind.Resolve(sym); found {
			env.Set(obj)
			return nil
		}
		env.Set(sx.MakeUndefined())
		return nil
	},
}
