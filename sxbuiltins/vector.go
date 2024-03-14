//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"fmt"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Vector returns its arguments as a vector.
var Vector = sxeval.Builtin{
	Name:     sx.VectorName,
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0:      func(_ *sxeval.Environment) (sx.Object, error) { return sx.Nil(), nil },
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return sx.Vector{arg}, nil
	},
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		return sx.Vector{arg0, arg1}, nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		return args, nil
	},
}

// VectorP returns true if the argument is a vector.
var VectorP = sxeval.Builtin{
	Name:     "vector?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		_, isVector := sx.GetVector(arg)
		return sx.MakeBoolean(isVector), nil
	},
}

// VectorSetBang overwrites the object at the given position.
var VectorSetBang = sxeval.Builtin{
	Name:     "vset!",
	MinArity: 3,
	MaxArity: 3,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		v, err := GetVector(args[0], 0)
		if err != nil {
			return nil, err
		}
		num, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		pos := num.(sx.Int64)
		if pos < 0 {
			return nil, fmt.Errorf("negative vector index not allowed: %v", pos)
		}
		if sx.Int64(len(v)) <= pos {
			return nil, fmt.Errorf("vector index out of range: %v", pos)
		}

		v[pos] = args[2]
		return v, nil
	},
}
