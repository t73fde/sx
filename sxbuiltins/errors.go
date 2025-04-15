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
	TestPure: nil, // is not pure, because error must occur at runtime.
	Fn0: func(_ *sxeval.Environment, _ *sxeval.Binding) (sx.Object, error) {
		return nil, fmt.Errorf("unspecified user error")
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		return nil, fmt.Errorf("%s", arg.GoString())
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
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
	Fn1: func(_ *sxeval.Environment, arg sx.Object, bind *sxeval.Binding) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		return nil, bind.MakeNotBoundError(sym)
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		sym, err := GetSymbol(args[0], 0)
		if err != nil {
			return nil, err
		}
		bind, err := GetBinding(args[1], 1)
		if err != nil {
			return nil, err
		}
		return nil, bind.MakeNotBoundError(sym)
	},
	NoCallError: true,
}
