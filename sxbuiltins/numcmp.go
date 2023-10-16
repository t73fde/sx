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

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

var (
	// NumLess implements a numeric comparision w.r.t to the less operation.
	NumLess = cmpMakeBuiltin("<", func(cmpRes int) bool { return cmpRes < 0 })

	// NumLessEqual implements a numeric comparision w.r.t to the less-equal operation.
	NumLessEqual = cmpMakeBuiltin("<=", func(cmpRes int) bool { return cmpRes <= 0 })

	// NumGreaterEqual implements a numeric comparision w.r.t to the greater-equal operation.
	NumGreaterEqual = cmpMakeBuiltin(">=", func(cmpRes int) bool { return cmpRes >= 0 })

	// NumGreater implements a numeric comparision w.r.t to the greater operation.
	NumGreater = cmpMakeBuiltin(">", func(cmpRes int) bool { return cmpRes > 0 })
)

func cmpMakeBuiltin(name string, cmpFn func(int) bool) sxeval.Builtin {
	return sxeval.Builtin{
		Name:     name,
		MinArity: 2,
		MaxArity: -1,
		IsPure:   true,
		Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
			acc, err := GetNumber(args, 0)
			if err != nil {
				return nil, err
			}
			for i := 1; i < len(args); i++ {
				num, err2 := GetNumber(args, i)
				if err2 != nil {
					return nil, err2
				}
				cmpRes := sx.NumCmp(acc, num)
				if !cmpFn(cmpRes) {
					return sx.Nil(), nil
				}
				acc = num
			}
			return sx.MakeBoolean(true), nil
		},
	}
}
