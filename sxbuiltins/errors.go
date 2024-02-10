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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Contains function to test for signalling errors.

// Error returns a generic error.
var Error = sxeval.Builtin{
	Name:     "error",
	MinArity: 0,
	MaxArity: -1,
	TestPure: nil,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("unspecified user error")
		}
		var sb strings.Builder
		for i, arg := range args {
			if i > 0 {
				sb.WriteByte(' ')
			}
			switch o := arg.(type) {
			case sx.String:
				sb.WriteString(string(o))
			case *sx.Symbol:
				sb.WriteString(o.GoString())
			default:
				sb.WriteString(arg.String())
			}
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
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		sym, err := GetSymbol(args, 0)
		if err != nil {
			return nil, err
		}
		bind := env.Binding()
		if len(args) == 2 {
			bind, err = GetBinding(args, 1)
			if err != nil {
				return nil, err
			}
		}
		return nil, sxeval.NotBoundError{Binding: bind, Sym: sym}
	},
	NoCallError: true,
}
