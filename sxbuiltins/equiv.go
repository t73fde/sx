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

// Contains function to test for equivalence of objects.

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

var Identical = sxeval.Builtin{
	Name:     "==",
	MinArity: 2,
	MaxArity: 0,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		for i := 1; i < len(args); i++ {
			if args[0] != args[i] {
				return sx.Nil(), nil
			}
		}
		return sx.MakeBoolean(true), nil
	},
}

var Equal = sxeval.Builtin{
	Name:     "=",
	MinArity: 2,
	MaxArity: 0,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		for i := 1; i < len(args); i++ {
			if !args[0].IsEqual(args[i]) {
				return sx.Nil(), nil
			}
		}
		return sx.MakeBoolean(true), nil
	},
}
