//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

// Contains function to test for equivalence of objects.

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// Identical returns true, if all arguments are identical objects.
var Identical = sxeval.Builtin{
	Name:     "==",
	MinArity: 2,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		for i := 1; i < len(args); i++ {
			if args[0] != args[i] {
				env.Kill(numargs - 1)
				env.Set(sx.Nil())
				return nil
			}
		}
		env.Kill(numargs - 1)
		env.Set(sx.T)
		return nil
	},
}

// Equal returns true if all arguments have the same value.
var Equal = sxeval.Builtin{
	Name:     "=",
	MinArity: 2,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		for i := 1; i < len(args); i++ {
			if !args[0].IsEqual(args[i]) {
				env.Kill(numargs - 1)
				env.Set(sx.Nil())
				return nil
			}
		}
		env.Kill(numargs - 1)
		env.Set(sx.T)
		return nil
	},
}
