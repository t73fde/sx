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
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		_, isVector := sx.GetVector(args[0])
		return sx.MakeBoolean(isVector), nil
	},
}

// VectorLength returns the length of the vector.
var VectorLength = sxeval.Builtin{
	Name:     "vector-length",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		v, err := GetVector(args, 0)
		if err != nil {
			return nil, err
		}
		return sx.Int64(len(v)), nil
	},
}

// VectorGet returns the object at the given position.
var VectorGet = sxeval.Builtin{
	Name:     "vget",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		v, pos, err := getVectorIndex(args, 0, 1)
		if err != nil {
			return nil, err
		}
		return v[pos], nil
	},
}

// VectorSetBang overwrites the object at the given position.
var VectorSetBang = sxeval.Builtin{
	Name:     "vset!",
	MinArity: 3,
	MaxArity: 3,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		v, pos, err := getVectorIndex(args, 0, 1)
		if err != nil {
			return nil, err
		}
		v[pos] = args[2]
		return v, nil
	},
}

func getVectorIndex(args sx.Vector, posV, posI int) (sx.Vector, sx.Int64, error) {
	v, err := GetVector(args, posV)
	if err != nil {
		return nil, 0, err
	}
	num, err := GetNumber(args, posI)
	if err != nil {
		return nil, 0, err
	}
	pos := num.(sx.Int64)
	if pos < 0 {
		return nil, 0, fmt.Errorf("negative vector index not allowed: %v", pos)
	}
	if sx.Int64(len(v)) <= pos {
		return nil, 0, fmt.Errorf("vector index out of range: %v", pos)
	}
	return v, pos, nil
}

// Vector2List returns the vector as a (pair) list.
var Vector2List = sxeval.Builtin{
	Name:     "vector->list",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		v, err := GetVector(args, 0)
		if err != nil {
			return nil, err
		}
		return sx.MakeList(v...), nil
	},
}
