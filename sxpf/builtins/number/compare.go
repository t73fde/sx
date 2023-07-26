//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package number

import (
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
)

// Less implements a numeric comparision w.r.t to the less operation.
func Less(args []sxpf.Object) (sxpf.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes < 0 })
}

// LessEqual implements a numeric comparision w.r.t to the less-equal operation.
func LessEqual(args []sxpf.Object) (sxpf.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes <= 0 })
}

// Equal implements a numeric comparision w.r.t to the equal operation.
func Equal(args []sxpf.Object) (sxpf.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes == 0 })
}

// GreaterEqual implements a numeric comparision w.r.t to the greater-equal operation.
func GreaterEqual(args []sxpf.Object) (sxpf.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes >= 0 })
}

// Greater implements a numeric comparision w.r.t to the greater operation.
func Greater(args []sxpf.Object) (sxpf.Object, error) {
	return cmpBuiltin(args, func(cmpRes int) bool { return cmpRes > 0 })
}

func cmpBuiltin(args []sxpf.Object, cmpFn func(int) bool) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 2, 0)
	acc, err := builtins.GetNumber(err, args, 0)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(args); i++ {
		num, err2 := builtins.GetNumber(err, args, i)
		if err2 != nil {
			return nil, err2
		}
		cmpRes := sxpf.NumCmp(acc, num)
		if !cmpFn(cmpRes) {
			return sxpf.False, nil
		}
		acc = num
	}
	return sxpf.True, nil
}

// Min implements the minimum finding operation on numbers.
func Min(args []sxpf.Object) (sxpf.Object, error) {
	return minmaxBuiltin(args, func(cmpRes int) bool { return cmpRes <= 0 })
}

// Max implements the maximum finding operation on numbers.
func Max(args []sxpf.Object) (sxpf.Object, error) {
	return minmaxBuiltin(args, func(cmpRes int) bool { return cmpRes >= 0 })
}

func minmaxBuiltin(args []sxpf.Object, cmpFn func(int) bool) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 2, 0)
	acc, err := builtins.GetNumber(err, args, 0)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(args); i++ {
		num, err2 := builtins.GetNumber(err, args, i)
		if err2 != nil {
			return nil, err2
		}
		cmpRes := sxpf.NumCmp(acc, num)
		if !cmpFn(cmpRes) {
			acc = num
		}
	}
	return acc, nil
}
