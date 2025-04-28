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
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// Length returns the length of the given sequence.
var Length = sxeval.Builtin{
	Name:     "length",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		seq, err := GetSequence(env.Top(), 0)
		if err != nil {
			return err
		}
		if sx.IsNil(seq) {
			env.Set(sx.Int64(0))
			return nil
		}
		env.Set(sx.Int64(int64(seq.Length())))
		return nil
	},
}

// LengthLess returns true if the length of the sequence is less than the given number.
var LengthLess = sxeval.Builtin{
	Name:     "length<",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		seq, err := GetSequence(env.Top(), 0)
		if err != nil {
			return err
		}
		n, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		env.Set(sx.MakeBoolean(seq.LengthLess(int(n.(sx.Int64)))))
		return nil
	},
}

// LengthGreater returns true if the length of the sequence is greater than the given number.
var LengthGreater = sxeval.Builtin{
	Name:     "length>",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		seq, err := GetSequence(env.Top(), 0)
		if err != nil {
			return err
		}
		n, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		env.Set(sx.MakeBoolean(seq.LengthGreater(int(n.(sx.Int64)))))
		return nil
	},
}

// LengthEqual returns true if the length of the sequence is equal to the given number.
var LengthEqual = sxeval.Builtin{
	Name:     "length=",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		seq, err := GetSequence(env.Top(), 0)
		if err != nil {
			return err
		}
		n, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		env.Set(sx.MakeBoolean(seq.LengthEqual(int(n.(sx.Int64)))))
		return nil
	},
}

// Nth returns the n-th element of the sequence.
var Nth = sxeval.Builtin{
	Name:     "nth",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		seq, err := GetSequence(env.Top(), 0)
		if err != nil {
			return err
		}
		n, err := GetNumber(arg1, 1)
		if err != nil {
			return err
		}
		obj, err := seq.Nth(int(n.(sx.Int64)))
		env.Set(obj)
		return err
	},
}

// Sequence2List returns the sequence as a (pair) list.
var Sequence2List = sxeval.Builtin{
	Name:     "seq->list",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		seq, err := GetSequence(env.Top(), 0)
		if err != nil {
			return err
		}
		env.Set(seq.MakeList())
		return nil
	},
}
