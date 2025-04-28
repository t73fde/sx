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

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// Vector returns its arguments as a vector.
var Vector = sxeval.Builtin{
	Name:     sx.VectorName,
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0:      func(env *sxeval.Environment, _ *sxeval.Binding) error { env.Push(sx.Nil()); return nil },
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Set(sx.Vector{env.Top()})
		return nil
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		res := sx.Vector(env.CopyArgs(numargs))
		env.Kill(numargs - 1)
		env.Set(res)
		return nil
	},
}

// VectorP returns true if the argument is a vector.
var VectorP = sxeval.Builtin{
	Name:     "vector?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		_, isVector := sx.GetVector(env.Top())
		env.Set(sx.MakeBoolean(isVector))
		return nil
	},
}

// VectorSetBang overwrites the object at the given position.
var VectorSetBang = sxeval.Builtin{
	Name:     "vset!",
	MinArity: 3,
	MaxArity: 3,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		arg2 := env.Pop()
		arg1 := env.Pop()
		v, err := GetVector(env.Top(), 0)
		if err != nil {
			return err
		}
		num, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		pos := num.(sx.Int64)
		if pos < 0 {
			return fmt.Errorf("negative vector index not allowed: %v", pos)
		}
		if sx.Int64(len(v)) <= pos {
			return fmt.Errorf("vector index out of range: %v", pos)
		}

		v[pos] = arg2
		return nil
	},
}
