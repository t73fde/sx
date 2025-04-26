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

import (
	"fmt"

	"t73f.de/r/sx"
)

// ParseEnvironment is a parsing environment.
type ParseEnvironment struct {
	env *Environment
}

// ParseObserver monitors the parsing process.
type ParseObserver interface {
	// BeforeParse is called immediate before the given form is parsed.
	// The observer may change the form and abort the parse with an error.
	BeforeParse(*ParseEnvironment, sx.Object, *Binding) (sx.Object, error)

	// AfterParse is called immediate after the given form was parsed to the expression.
	AfterParse(*ParseEnvironment, sx.Object, Expr, *Binding, error)
}

// Parse the form into an expression.
func (pe *ParseEnvironment) Parse(form sx.Object, bind *Binding) (expr Expr, err error) {
	if observer := pe.env.obParse; observer != nil {
		form, err = observer.BeforeParse(pe, form, bind)
	}
	if err == nil {
		expr, err = pe.parseForm(form, bind)
	}
	if observer := pe.env.obParse; observer != nil {
		observer.AfterParse(pe, form, expr, bind, err)
	}
	return expr, err
}

func (pe *ParseEnvironment) parseForm(form sx.Object, bind *Binding) (Expr, error) {
restart:
	if sx.IsNil(form) {
		return NilExpr, nil
	}
	switch f := form.(type) {
	case *sx.Symbol:
		return UnboundSymbolExpr{sym: f}, nil
	case *sx.Pair:
		expr, err := pe.parsePair(f, bind)
		if err == nil {
			return expr, nil
		}
		if again, isAgain := err.(errParseAgain); isAgain {
			pe, form = again.pe, again.form
			goto restart
		}
		return nil, err
	case *ExprObj:
		return f.expr, nil
	}
	return ObjExpr{Obj: form}, nil
}

func (pe *ParseEnvironment) parsePair(pair *sx.Pair, bind *Binding) (Expr, error) {
	var proc Expr
	first := pair.Car()
	if sym, isSymbol := sx.GetSymbol(first); isSymbol {
		if val, found := bind.Resolve(sym); found {
			if sp, isSyntax := GetSyntax(val); isSyntax {
				return sp.Parse(pe, pair.Tail(), bind)
			}
		}
		proc = UnboundSymbolExpr{sym: sym}
	} else {
		p, err := pe.Parse(first, bind)
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
		expr, err2 := pe.Parse(elem.Car(), bind)
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

// ParseAgain signals the parser that the form must be parsed again, e.g. for macro expansion.
func (pe *ParseEnvironment) ParseAgain(form sx.Object) error {
	return errParseAgain{pe: pe, form: form}
}

// errParseAgain is a non-error error signalling that the given form should be
// parsed again in the given environment.
type errParseAgain struct {
	pe   *ParseEnvironment
	form sx.Object
}

func (e errParseAgain) Error() string { return fmt.Sprintf("Again: %T/%v", e.form, e.form) }
