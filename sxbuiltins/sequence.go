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
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		seq, err := GetSequence(arg, 0)
		if err != nil {
			return nil, err
		}
		if sx.IsNil(seq) {
			return sx.Int64(0), nil
		}
		return sx.Int64(int64(seq.Length())), nil
	},
}

// LengthLess returns true if the length of the sequence is less than the given number.
var LengthLess = sxeval.Builtin{
	Name:     "length<",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		seq, err := GetSequence(args[0], 0)
		if err != nil {
			return nil, err
		}
		n, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		return sx.MakeBoolean(seq.LengthLess(int(n.(sx.Int64)))), nil
	},
}

// LengthGreater returns true if the length of the sequence is greater than the given number.
var LengthGreater = sxeval.Builtin{
	Name:     "length>",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		seq, err := GetSequence(args[0], 0)
		if err != nil {
			return nil, err
		}
		n, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		return sx.MakeBoolean(seq.LengthGreater(int(n.(sx.Int64)))), nil
	},
}

// LengthEqual returns true if the length of the sequence is equal to the given number.
var LengthEqual = sxeval.Builtin{
	Name:     "length=",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		seq, err := GetSequence(args[0], 0)
		if err != nil {
			return nil, err
		}
		n, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		return sx.MakeBoolean(seq.LengthEqual(int(n.(sx.Int64)))), nil
	},
}

// Nth returns the n-th element of the sequence.
var Nth = sxeval.Builtin{
	Name:     "nth",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		seq, err := GetSequence(args[0], 0)
		if err != nil {
			return nil, err
		}
		n, err := GetNumber(args[1], 1)
		if err != nil {
			return nil, err
		}
		return seq.Nth(int(n.(sx.Int64)))
	},
}

// Sequence2List returns the sequence as a (pair) list.
var Sequence2List = sxeval.Builtin{
	Name:     "seq->list",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		seq, err := GetSequence(arg, 0)
		if err != nil {
			return nil, err
		}
		return seq.MakeList(), nil
	},
}
