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
	"fmt"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// NumberP returns true if the argument is a number.
var NumberP = sxeval.Builtin{
	Name:     "number?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		_, ok := sx.GetNumber(args[0])
		return sx.MakeBoolean(ok), nil
	},
}

// Add is the builtin that implements (+ n...)
var Add = sxeval.Builtin{
	Name:     "+",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		num0, isNumber := sx.GetNumber(arg0)
		if !isNumber {
			return nil, fmt.Errorf("number expected")
		}
		num1, isNumber := sx.GetNumber(arg1)
		if !isNumber {
			return nil, fmt.Errorf("number expected")
		}
		return sx.NumAdd(num0, num1), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		acc := sx.Number(sx.Int64(0))
		if len(args) == 0 {
			return acc, nil
		}

		for i := range len(args) {
			num, err := GetNumber(args, i)
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
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		num0, isNumber := sx.GetNumber(arg0)
		if !isNumber {
			return nil, fmt.Errorf("number expected")
		}
		num1, isNumber := sx.GetNumber(arg1)
		if !isNumber {
			return nil, fmt.Errorf("number expected")
		}
		return sx.NumSub(num0, num1), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		acc, err := GetNumber(args, 0)
		if err != nil {
			return nil, err
		}
		if len(args) == 1 {
			return sx.NumNeg(acc), nil
		}
		for i := 1; i < len(args); i++ {
			num, err2 := GetNumber(args, i)
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
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		num0, isNumber := sx.GetNumber(arg0)
		if !isNumber {
			return nil, fmt.Errorf("number expected")
		}
		num1, isNumber := sx.GetNumber(arg1)
		if !isNumber {
			return nil, fmt.Errorf("number expected")
		}
		return sx.NumMul(num0, num1), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		acc := sx.Number(sx.Int64(1))
		for i := range len(args) {
			num, err := GetNumber(args, i)
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
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		acc, err := GetNumber(args, 0)
		if err != nil {
			return nil, err
		}
		num, err := GetNumber(args, 1)
		if err != nil {
			return nil, err
		}
		return sx.NumDiv(acc, num)
	},
}

// Mod is the builtin that implements (mod n m)
var Mod = sxeval.Builtin{
	Name:     "mod",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		acc, err := GetNumber(args, 0)
		if err != nil {
			return nil, err
		}
		num, err := GetNumber(args, 1)
		if err != nil {
			return nil, err
		}
		return sx.NumMod(acc, num)
	},
}
