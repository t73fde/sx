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

// Length returns the length of the given sequence.
var Length = sxeval.Builtin{
	Name:     "length",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		seq, err := GetSequence(args, 0)
		if err != nil {
			return nil, err
		}
		if sx.IsNil(seq) {
			return sx.Int64(0), nil
		}
		return sx.Int64(int64(seq.Length())), nil
	},
}

// Nth returns the n-th element of the sequence.
var Nth = sxeval.Builtin{
	Name:     "nth",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		seq, err := GetSequence(args, 0)
		if err != nil {
			return nil, err
		}
		if sx.IsNil(seq) {
			return nil, fmt.Errorf("sequence is nil")
		}
		n, err := GetNumber(args, 1)
		if err != nil {
			return nil, err
		}
		return seq.Nth(int(n.(sx.Int64)))
	},
}
