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

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/eval"
	"zettelstore.de/sx.fossil/sxpf/reader"
)

// InstallQuoteReader will install a quote symbol as a reader macro.
func InstallQuoteReader(rd *reader.Reader, quoteSym *sxpf.Symbol, initCh rune) {
	if rd.SymbolFactory() != quoteSym.Factory() {
		panic("reader symbol factory is not factory of symbol")
	}
	rd.SetMacro(initCh, makeQuotationMacro(quoteSym))
}

// InstallQuoteSyntax will setup the system to allow quoting values.
func InstallQuoteSyntax(env sxpf.Environment, symQuote *sxpf.Symbol) error {
	err := env.Bind(
		symQuote,
		eval.MakeSyntax(
			symQuote.Name(),
			func(_ *eval.Engine, _ sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
				if sxpf.IsNil(args) {
					return nil, eval.ErrNoArgs
				}
				if args.Tail() != nil {
					return nil, fmt.Errorf("more than one argument: %v", args)
				}
				return eval.ObjExpr{Obj: args.Car()}, nil
			}))
	return err
}

func makeQuotationMacro(sym *sxpf.Symbol) reader.Macro {
	return func(rd *reader.Reader, _ rune) (sxpf.Object, error) {
		obj, err := rd.Read()
		if err == nil {
			return sxpf.Nil().Cons(obj).Cons(sym), nil
		}
		return obj, err
	}
}
