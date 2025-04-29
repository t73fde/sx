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
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		arg := env.Top()
		if _, isString := sx.GetString(arg); isString {
			return nil
		}
		env.Set(sx.MakeString(arg.GoString()))
		return nil
	},
}

// Concat appends all its string arguments.
var Concat = sxeval.Builtin{
	Name:     "concat",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Push(sx.String{})
		return nil
	},
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		if _, err := GetString(env.Top(), 0); err != nil {
			env.Kill(1)
			return err
		}
		return nil
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		s, err := GetString(args[0], 0)
		if err != nil {
			env.Kill(numargs)
			return err
		}
		var sb strings.Builder
		sb.WriteString(s.GetValue())
		for i := 1; i < len(args); i++ {
			s, err = GetString(args[i], i)
			if err != nil {
				env.Kill(numargs)
				return err
			}
			sb.WriteString(s.GetValue())
		}
		env.Kill(numargs - 1)
		env.Set(sx.MakeString(sb.String()))
		return nil
	},
}
