//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// ToString transforms its argument into its string representation.
var ToString = sxeval.Builtin{
	Name:     "->string",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		if s, isString := sx.GetString(arg); isString {
			return s, nil
		}
		return sx.MakeString(arg.GoString()), nil
	},
}

// Concat appends all its string arguments.
var Concat = sxeval.Builtin{
	Name:     "concat",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(_ *sxeval.Environment, _ *sxeval.Frame) (sx.Object, error) {
		return sx.String{}, nil
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		return GetString(arg, 0)
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		s, err := GetString(args[0], 0)
		if err != nil {
			return nil, err
		}
		var sb strings.Builder
		sb.WriteString(s.GetValue())
		for i := 1; i < len(args); i++ {
			s, err = GetString(args[i], i)
			if err != nil {
				return nil, err
			}
			sb.WriteString(s.GetValue())
		}
		return sx.MakeString(sb.String()), nil
	},
}
