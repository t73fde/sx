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

	"zettelstore.de/sx.fossil"
)

// ParseEnvironment is a parsing environment.
type ParseEnvironment struct {
	binding  *Binding
	observer ParseObserver
}

// ParseObserver monitors the parsing process.
type ParseObserver interface {
	// BeforeParse is called immediate before the given form is parsed.
	// The observer may change the form.
	BeforeParse(*ParseEnvironment, sx.Object) sx.Object

	// AfterParse is called immediate after the given form was parsed to the expression.
	AfterParse(*ParseEnvironment, sx.Object, Expr, error)
}

func (pf *ParseEnvironment) Parse(form sx.Object) (Expr, error) {
	if observer := pf.observer; observer != nil {
		obj := observer.BeforeParse(pf, form)
		expr, err := pf.parseForm(obj)
		observer.AfterParse(pf, obj, expr, err)
		return expr, err
	}
	return pf.parseForm(form)
}

func (pf *ParseEnvironment) parseForm(form sx.Object) (Expr, error) {
restart:
	if sx.IsNil(form) {
		return NilExpr, nil
	}
	switch f := form.(type) {
	case *sx.Symbol:
		return ResolveSymbolExpr{Symbol: f}, nil
	case *sx.Pair:
		expr, err := pf.parsePair(f)
		if err == nil {
			return expr, nil
		}
		if again, isAgain := err.(errParseAgain); isAgain {
			pf, form = again.pf, again.form
			goto restart
		}
		return nil, err
	case *ExprObj:
		return f.expr, nil
	}
	return ObjExpr{Obj: form}, nil
}

func (pf *ParseEnvironment) parsePair(pair *sx.Pair) (Expr, error) {
	var proc Expr
	first := pair.Car()
	if sym, isSymbol := sx.GetSymbol(first); isSymbol {
		if val, found := pf.Resolve(sym); found {
			if sp, isSyntax := GetSyntax(val); isSyntax {
				return sp.Parse(pf, pair.Tail())
			}
		}
		proc = ResolveSymbolExpr{Symbol: sym}
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

func (pf *ParseEnvironment) ParseAgain(form sx.Object) error {
	return errParseAgain{pf: pf, form: form}
}

// errParseAgain is a non-error error signalling that the given form should be
// parsed again in the given environment.
type errParseAgain struct {
	pf   *ParseEnvironment
	form sx.Object
}

func (e errParseAgain) Error() string { return fmt.Sprintf("Again: %T/%v", e.form, e.form) }

func (pf *ParseEnvironment) MakeChildFrame(name string, baseSize int) *ParseEnvironment {
	return &ParseEnvironment{
		binding:  pf.binding.MakeChildBinding(name, baseSize),
		observer: pf.observer,
	}
}

func (pf *ParseEnvironment) Bind(sym *sx.Symbol, obj sx.Object) error {
	return pf.binding.Bind(sym, obj)
}

func (pf *ParseEnvironment) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return pf.binding.Resolve(sym)
}
func (pf *ParseEnvironment) Binding() *Binding { return pf.binding }
