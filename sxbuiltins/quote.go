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

// Quasiquote implementation is a little bit too simple as it does not support
// nested quasiquotes.

import (
	"fmt"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// QuoteS parses the quote syntax.
var QuoteS = sxeval.Special{
	Name: sx.SymbolQuote.String(),
	Fn: func(_ *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		if sx.IsNil(args) {
			return nil, sxeval.ErrNoArgs
		}
		if args.Tail() != nil {
			return nil, fmt.Errorf("more than one argument: %v", args)
		}
		return sxeval.ObjExpr{Obj: args.Car()}, nil
	},
}
