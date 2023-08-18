//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval

import (
	"fmt"

	"zettelstore.de/sx.fossil"
)

// Parser transform an object into an executable expression.
type Parser interface {
	Parse(*Frame, sx.Object) (Expr, error)
}

// ErrParseAgain is a non-error error signalling that the given form should be
// parsed again in the given environment.
type ErrParseAgain struct {
	Frame *Frame
	Form  sx.Object
}

func (e ErrParseAgain) Error() string { return fmt.Sprintf("Again: %T/%v", e.Form, e.Form) }

// defaultParser is the parser for normal use.
type defaultParser struct{}

var myDefaultParser defaultParser

func (dp *defaultParser) Parse(frame *Frame, form sx.Object) (Expr, error) {
restart:
	if sx.IsNil(form) {
		return NilExpr, nil
	}
	switch f := form.(type) {
	case *sx.Symbol:
		return ResolveExpr{Symbol: f}, nil
	case *sx.Pair:
		expr, err := dp.parsePair(frame, f)
		if err == nil {
			return expr, nil
		}
		if again, isAgain := err.(ErrParseAgain); isAgain {
			frame, form = again.Frame, again.Form
			goto restart
		}
		return nil, err
	case sx.Boolean:
		if f == sx.False {
			return FalseExpr, nil
		}
		return TrueExpr, nil
	}
	return ObjExpr{Obj: form}, nil
}

func (*defaultParser) parsePair(frame *Frame, pair *sx.Pair) (Expr, error) {
	var proc Expr
	first := pair.Car()
	if sym, isSymbol := sx.GetSymbol(first); isSymbol {
		if val, found := frame.Resolve(sym); found {
			if sp, isSpecial := GetSpecial(val); isSpecial {
				return sp.Parse(frame, pair.Tail())
			}
		}
		proc = ResolveExpr{Symbol: sym}
	} else {
		p, err := frame.Parse(first)
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
		expr, err2 := frame.Parse(elem.Car())
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
