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

import "zettelstore.de/sx.fossil"

// NumberPold is the boolean that returns true if the argument is a number.
func NumberPold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sx.GetNumber(args[0])
	return sx.MakeBoolean(ok), nil
}

// AddOld is the builtin that implements (+ n...)
func AddOld(args []sx.Object) (sx.Object, error) {
	acc := sx.Number(sx.Int64(0))
	if len(args) == 0 {
		return acc, nil
	}

	for i := 0; i < len(args); i++ {
		num, err := GetNumber(nil, args, i)
		if err != nil {
			return nil, err
		}
		acc = sx.NumAdd(acc, num)
	}
	return acc, nil
}

// SubOld is the builtin that implements (- n n...)
func SubOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 0)
	acc, err := GetNumber(err, args, 0)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return sx.NumNeg(acc), nil
	}
	for i := 1; i < len(args); i++ {
		num, err2 := GetNumber(nil, args, i)
		if err2 != nil {
			return nil, err2
		}
		acc = sx.NumSub(acc, num)
	}
	return acc, nil
}

// MulOld is the builtin that implements (* n...)
func MulOld(args []sx.Object) (sx.Object, error) {
	acc := sx.Number(sx.Int64(1))
	if len(args) == 0 {
		return acc, nil
	}

	for i := 0; i < len(args); i++ {
		num, err := GetNumber(nil, args, i)
		if err != nil {
			return nil, err
		}
		acc = sx.NumMul(acc, num)
	}
	return acc, nil
}

// DivOld is the builtin that implements (div n m)
func DivOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 2, 2)
	acc, err := GetNumber(err, args, 0)
	num, err := GetNumber(err, args, 1)
	if err != nil {
		return nil, err
	}
	return sx.NumDiv(acc, num)
}

// ModOld is the builtin that implements (mod n m)
func ModOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 2, 2)
	acc, err := GetNumber(err, args, 0)
	num, err := GetNumber(err, args, 1)
	if err != nil {
		return nil, err
	}
	return sx.NumMod(acc, num)
}
