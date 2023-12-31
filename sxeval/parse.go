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

package sxeval

import "zettelstore.de/sx.fossil"

// Parser transform an object into an executable expression.
type Parser interface {
	Parse(*ParseFrame, sx.Object) (Expr, error)
}

// defaultParser is the parser for normal use.
type defaultParser struct{}

var myDefaultParser defaultParser

func (dp *defaultParser) Parse(pf *ParseFrame, form sx.Object) (Expr, error) {
restart:
	if sx.IsNil(form) {
		return NilExpr, nil
	}
	switch f := form.(type) {
	case *sx.Symbol:
		return ResolveSymbolExpr{Symbol: f}, nil
	case *sx.Pair:
		expr, err := dp.parsePair(pf, f)
		if err == nil {
			return expr, nil
		}
		if again, isAgain := err.(errParseAgain); isAgain {
			pf, form = again.pf, again.form
			goto restart
		}
		return nil, err
	}
	return ObjExpr{Obj: form}, nil
}

func (*defaultParser) parsePair(pf *ParseFrame, pair *sx.Pair) (Expr, error) {
	var proc Expr
	first := pair.Car()
	if sym, isSymbol := sx.GetSymbol(first); isSymbol {
		if val, found := pf.Resolve(sym); found {
			if sp, isSyntax := GetSyntax(val); isSyntax {
				return sp.Parse(pf, pair.Tail())
			}
		}
		proc = ResolveProcSymbolExpr{Symbol: sym}
	} else {
		p, err := pf.Parse(first)
		if err != nil {
			return nil, err
		}
		proc = p
	}

	var exprArgs []Expr
	arg := pair.Cdr()
	for {
		if sx.IsNil(arg) {
			break
		}
		elem, isPair := sx.GetPair(arg)
		if !isPair {
			return nil, sx.ErrImproper{Pair: pair}
		}
		expr, err2 := pf.Parse(elem.Car())
		if err2 != nil {
			return nil, err2
		}
		exprArgs = append(exprArgs, expr)
		arg = elem.Cdr()
	}

	ce := CallExpr{
		Proc: proc,
		Args: exprArgs,
	}
	return &ce, nil
}
