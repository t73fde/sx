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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// ToString transforms its argument into its string representation.
var ToString = sxeval.Builtin{
	Name:     "->string",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		obj := args[0]
		if s, isString := sx.GetString(obj); isString {
			return s, nil
		}
		return sx.String(obj.String()), nil
	},
}

// StringAppend append all its string arguments.
var Concat = sxeval.Builtin{
	Name:     "concat",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		if len(args) == 0 {
			return sx.String(""), nil
		}
		s, err := GetString(args, 0)
		if err != nil {
			return nil, err
		}
		if len(args) == 1 {
			return s, nil
		}
		var sb strings.Builder
		sb.WriteString(string(s))
		for i := 1; i < len(args); i++ {
			s, err = GetString(args, i)
			if err != nil {
				return nil, err
			}
			sb.WriteString(string(s))
		}
		return sx.String(sb.String()), nil
	},
}
