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

// VectorSetBang overwrites the object at the given position.
var VectorSetBang = sxeval.Builtin{
	Name:     "vset!",
	MinArity: 3,
	MaxArity: 3,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		v, err := GetVector(args, 0)
		if err != nil {
			return nil, err
		}
		num, err := GetNumber(args, 1)
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
