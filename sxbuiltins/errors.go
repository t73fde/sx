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

import (
	"fmt"
	"strings"

	"t73f.de/r/sx/sxeval"
)

// Contains function to test for signalling errors.

// Error returns a generic error.
var Error = sxeval.Builtin{
	Name:     "error",
	MinArity: 0,
	MaxArity: -1,
	TestPure: nil, // is not pure, because error must occur at runtime.
	Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		return fmt.Errorf("unspecified user error")
	},
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		return fmt.Errorf("%s", env.Pop().GoString())
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		var sb strings.Builder
		for i, arg := range env.Args(numargs) {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(arg.GoString())
		}
		env.Kill(numargs)
		return fmt.Errorf("%s", sb.String())
	},
	NoCallError: true,
}

// NotBoundError returns an error signalling that a symbol was not bound.
var NotBoundError = sxeval.Builtin{
	Name:     "not-bound-error",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		sym, err := GetSymbol(env.Pop(), 0)
		if err != nil {
			return err
		}
		return bind.MakeNotBoundError(sym)
	},
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		sym, err := GetSymbol(env.Pop(), 0)
		if err != nil {
			return err
		}
		bind, err := GetBinding(arg1, 1)
		if err != nil {
			return err
		}
		return bind.MakeNotBoundError(sym)
	},
	NoCallError: true,
}
