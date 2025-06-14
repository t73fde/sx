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
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		_, ok := sx.GetNumber(arg)
		return sx.MakeBoolean(ok), nil
	},
}

// Add is the builtin that implements (+ n...)
var Add = sxeval.Builtin{
	Name:     "+",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(_ *sxeval.Environment, _ *sxeval.Frame) (sx.Object, error) {
		return sx.Int64(0), nil
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		num, err := GetNumber(arg, 0)
		return num, err
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		acc := sx.Number(sx.Int64(0))
		for i := range len(args) {
			num, err := GetNumber(args[i], i)
			if err != nil {
				return nil, err
			}
			acc = sx.NumAdd(acc, num)
		}
		return acc, nil
	},
}

// Sub is the builtin that implements (- n n...)
var Sub = sxeval.Builtin{
	Name:     "-",
	MinArity: 1,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		num, err := GetNumber(arg, 0)
		if err != nil {
			return nil, err
		}
		return sx.NumNeg(num), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		acc, err := GetNumber(args[0], 0)
		if err != nil {
			return nil, err
		}
		for i := 1; i < len(args); i++ {
			num, err2 := GetNumber(args[i], i)
			if err2 != nil {
				return nil, err2
			}
			acc = sx.NumSub(acc, num)
		}
		return acc, nil
	},
}

// Mul is the builtin that implements (* n...)
var Mul = sxeval.Builtin{
	Name:     "*",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(_ *sxeval.Environment, _ *sxeval.Frame) (sx.Object, error) {
		return sx.Int64(1), nil
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		num, err := GetNumber(arg, 0)
		return num, err
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		acc := sx.Number(sx.Int64(1))
		for i := range len(args) {
			num, err := GetNumber(args[i], i)
			if err != nil {
				return nil, err
			}
			acc = sx.NumMul(acc, num)
		}
		return acc, nil
	},
}

// Div is the builtin that implements (div n m).
var Div = sxeval.Builtin{
	Name:     "div",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		num0, err := GetNumber(args[0], 0)
		if err != nil {
			return nil, err
		}
		num1, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		return sx.NumDiv(num0, num1)
	},
}

// Mod is the builtin that implements (mod n m)
var Mod = sxeval.Builtin{
	Name:     "mod",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		num0, err := GetNumber(args[0], 0)
		if err != nil {
			return nil, err
		}
		num1, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		return sx.NumMod(num0, num1)
	},
}
