//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

// Contains builtins to work with numbers.

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// NumberP returns true if the argument is a number.
var NumberP = sxeval.Builtin{
	Name:     "number?",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		_, ok := sx.GetNumber(args[0])
		return sx.MakeBoolean(ok), nil
	},
}

// Add is the builtin that implements (+ n...)
var Add = sxeval.Builtin{
	Name:     "+",
	MinArity: 0,
	MaxArity: -1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		acc := sx.Number(sx.Int64(0))
		if len(args) == 0 {
			return acc, nil
		}

		for i := 0; i < len(args); i++ {
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
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
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
	IsPure:   false,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		acc := sx.Number(sx.Int64(1))
		for i := 0; i < len(args); i++ {
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
	IsPure:   false,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
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
	IsPure:   false,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
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
