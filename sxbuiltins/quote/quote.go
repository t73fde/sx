//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package quote contains functions to use quotations
// These are: quote, quasiquote, unquote, unquote-splicing.
//
// Quasiquote implementation is a little bit too simple as it does not support
// nested quasiquotes.
package quote

import (
	"fmt"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
	"zettelstore.de/sx.fossil/sxreader"
)

// InstallQuoteReader will install a quote symbol as a reader macro.
func InstallQuoteReader(rd *sxreader.Reader, quoteSym *sx.Symbol, initCh rune) {
	if rd.SymbolFactory() != quoteSym.Factory() {
		panic("reader symbol factory is not factory of symbol")
	}
	rd.SetMacro(initCh, makeQuotationMacro(quoteSym))
}

// InstallQuoteSyntax will setup the system to allow quoting values.
func InstallQuoteSyntax(env sxeval.Environment, symQuote *sx.Symbol) (sxeval.Environment, error) {
	return env.Bind(
		symQuote,
		sxeval.MakeSyntax(
			symQuote.Name(),
			func(_ *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
				if sx.IsNil(args) {
					return nil, sxeval.ErrNoArgs
				}
				if args.Tail() != nil {
					return nil, fmt.Errorf("more than one argument: %v", args)
				}
				return sxeval.ObjExpr{Obj: args.Car()}, nil
			}))
}

func makeQuotationMacro(sym *sx.Symbol) sxreader.Macro {
	return func(rd *sxreader.Reader, _ rune) (sx.Object, error) {
		obj, err := rd.Read()
		if err == nil {
			return sx.Nil().Cons(obj).Cons(sym), nil
		}
		return obj, err
	}
}
