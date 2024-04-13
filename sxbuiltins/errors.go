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

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// Contains function to test for signalling errors.

// Error returns a generic error.
var Error = sxeval.Builtin{
	Name:     "error",
	MinArity: 0,
	MaxArity: -1,
	TestPure: nil,
	Fn0: func(_ *sxeval.Environment) (sx.Object, error) {
		return nil, fmt.Errorf("unspecified user error")
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return nil, fmt.Errorf("%s", arg.GoString())
	},
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		return nil, fmt.Errorf("%s %s", arg0.GoString(), arg1.GoString())
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		var sb strings.Builder
		for i, arg := range args {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(arg.GoString())
		}
		return nil, fmt.Errorf("%s", sb.String())
	},
	NoCallError: true,
}

// NotBoundError returns an error signalling that a symbol was not bound.
var NotBoundError = sxeval.Builtin{
	Name:     "not-bound-error",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		return nil, sxeval.NotBoundError{Binding: env.Binding(), Sym: sym}
	},
	Fn2: func(env *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		sym, err := GetSymbol(arg0, 0)
		if err != nil {
			return nil, err
		}
		bind, err := GetBinding(arg1, 1)
		if err != nil {
			return nil, err
		}
		return nil, sxeval.NotBoundError{Binding: bind, Sym: sym}
	},
	NoCallError: true,
}
