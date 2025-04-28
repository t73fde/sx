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

// Contains builtins to work with numbers.

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// NumberP returns true if the argument is a number.
var NumberP = sxeval.Builtin{
	Name:     "number?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		_, ok := sx.GetNumber(env.Top())
		env.Set(sx.MakeBoolean(ok))
		return nil
	},
}

// Add is the builtin that implements (+ n...)
var Add = sxeval.Builtin{
	Name:     "+",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Push(sx.Int64(0))
		return nil
	},
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		_, err := GetNumber(env.Top(), 0)
		return err
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		acc := sx.Number(sx.Int64(0))
		for i := range len(args) {
			num, err := GetNumber(args[i], i)
			if err != nil {
				env.Kill(numargs - 1)
				return err
			}
			acc = sx.NumAdd(acc, num)
		}
		env.Kill(numargs - 1)
		env.Set(acc)
		return nil
	},
}

// Sub is the builtin that implements (- n n...)
var Sub = sxeval.Builtin{
	Name:     "-",
	MinArity: 1,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		num, err := GetNumber(env.Top(), 0)
		env.Set(sx.NumNeg(num))
		return err
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		acc, err := GetNumber(args[0], 0)
		if err != nil {
			env.Kill(numargs - 1)
			return err
		}
		for i := 1; i < len(args); i++ {
			num, err2 := GetNumber(args[i], i)
			if err2 != nil {
				env.Kill(numargs - 1)
				return err2
			}
			acc = sx.NumSub(acc, num)
		}
		env.Kill(numargs - 1)
		env.Set(acc)
		return nil
	},
}

// Mul is the builtin that implements (* n...)
var Mul = sxeval.Builtin{
	Name:     "*",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Push(sx.Int64(1))
		return nil
	},
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		_, err := GetNumber(env.Top(), 0)
		return err
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		acc := sx.Number(sx.Int64(1))
		for i := range len(args) {
			num, err := GetNumber(args[i], i)
			if err != nil {
				env.Kill(numargs - 1)
				return err
			}
			acc = sx.NumMul(acc, num)
		}
		env.Kill(numargs - 1)
		env.Set(acc)
		return nil
	},
}

// Div is the builtin that implements (div n m).
var Div = sxeval.Builtin{
	Name:     "div",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		num0, err := GetNumber(env.Top(), 0)
		if err != nil {
			return err
		}
		num1, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		obj, err := sx.NumDiv(num0, num1)
		env.Set(obj)
		return err
	},
}

// Mod is the builtin that implements (mod n m)
var Mod = sxeval.Builtin{
	Name:     "mod",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		num0, err := GetNumber(env.Top(), 0)
		if err != nil {
			return err
		}
		num1, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		obj, err := sx.NumMod(num0, num1)
		env.Set(obj)
		return err
	},
}
