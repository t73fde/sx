//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package number contains builtins to work with numbers.
package number

import (
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
)

// NumberP is the boolean that returns true if the argument is a number.
func NumberP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sxpf.GetNumber(args[0])
	return sxpf.MakeBoolean(ok), nil
}

// Add is the builtin that implements (+ n...)
func Add(args []sxpf.Object) (sxpf.Object, error) {
	acc := sxpf.Number(sxpf.Int64(0))
	if len(args) == 0 {
		return acc, nil
	}

	for i := 0; i < len(args); i++ {
		num, err := builtins.GetNumber(nil, args, i)
		if err != nil {
			return nil, err
		}
		acc = sxpf.NumAdd(acc, num)
	}
	return acc, nil
}

// Sub is the builtin that implements (- n n...)
func Sub(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 0)
	acc, err := builtins.GetNumber(err, args, 0)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return sxpf.NumNeg(acc), nil
	}
	for i := 1; i < len(args); i++ {
		num, err2 := builtins.GetNumber(nil, args, i)
		if err2 != nil {
			return nil, err2
		}
		acc = sxpf.NumSub(acc, num)
	}
	return acc, nil
}

// Mul is the builtin that implements (* n...)
func Mul(args []sxpf.Object) (sxpf.Object, error) {
	acc := sxpf.Number(sxpf.Int64(1))
	if len(args) == 0 {
		return acc, nil
	}

	for i := 0; i < len(args); i++ {
		num, err := builtins.GetNumber(nil, args, i)
		if err != nil {
			return nil, err
		}
		acc = sxpf.NumMul(acc, num)
	}
	return acc, nil
}

// Div is the builtin that implements (div n m)
func Div(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 2, 2)
	acc, err := builtins.GetNumber(err, args, 0)
	num, err := builtins.GetNumber(err, args, 1)
	if err != nil {
		return nil, err
	}
	return sxpf.NumDiv(acc, num)
}

// Mod is the builtin that implements (mod n m)
func Mod(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 2, 2)
	acc, err := builtins.GetNumber(err, args, 0)
	num, err := builtins.GetNumber(err, args, 1)
	if err != nil {
		return nil, err
	}
	return sxpf.NumDiv(acc, num)
}
