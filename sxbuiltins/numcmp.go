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

import "zettelstore.de/sx.fossil"

// Less implements a numeric comparision w.r.t to the less operation.
func Less(args []sx.Object) (sx.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes < 0 })
}

// LessEqual implements a numeric comparision w.r.t to the less-equal operation.
func LessEqual(args []sx.Object) (sx.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes <= 0 })
}

// Equal implements a numeric comparision w.r.t to the equal operation.
func Equal(args []sx.Object) (sx.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes == 0 })
}

// GreaterEqual implements a numeric comparision w.r.t to the greater-equal operation.
func GreaterEqual(args []sx.Object) (sx.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes >= 0 })
}

// Greater implements a numeric comparision w.r.t to the greater operation.
func Greater(args []sx.Object) (sx.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes > 0 })
}

func cmpBuiltin(args []sx.Object, cmpFn func(int) bool) (sx.Object, error) {
	err := CheckArgs(args, 2, 0)
	acc, err := GetNumber(err, args, 0)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(args); i++ {
		num, err2 := GetNumber(err, args, i)
		if err2 != nil {
			return nil, err2
		}
		cmpRes := sx.NumCmp(acc, num)
		if !cmpFn(cmpRes) {
			return sx.False, nil
		}
		acc = num
	}
	return sx.True, nil
}

// Min implements the minimum finding operation on numbers.
func Min(args []sx.Object) (sx.Object, error) {
	return minmaxBuiltin(args, func(cmpRes int) bool { return cmpRes <= 0 })
}

// Max implements the maximum finding operation on numbers.
func Max(args []sx.Object) (sx.Object, error) {
	return minmaxBuiltin(args, func(cmpRes int) bool { return cmpRes >= 0 })
}

func minmaxBuiltin(args []sx.Object, cmpFn func(int) bool) (sx.Object, error) {
	err := CheckArgs(args, 2, 0)
	acc, err := GetNumber(err, args, 0)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(args); i++ {
		num, err2 := GetNumber(err, args, i)
		if err2 != nil {
			return nil, err2
		}
		cmpRes := sx.NumCmp(acc, num)
		if !cmpFn(cmpRes) {
			acc = num
		}
	}
	return acc, nil
}
