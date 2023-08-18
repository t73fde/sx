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
	Parse(*Engine, Environment, sx.Object) (Expr, error)
}

// ErrParseAgain is a non-error error signalling that the given form should be
// parsed again in the given environment.
type ErrParseAgain struct {
	Env  Environment
	Form sx.Object
}

func (e ErrParseAgain) Error() string { return fmt.Sprintf("Again: %T/%v", e.Form, e.Form) }

// defaultParser is the parser for normal use.
type defaultParser struct{}

var myDefaultParser defaultParser

func (dp *defaultParser) Parse(eng *Engine, env Environment, form sx.Object) (Expr, error) {
restart:
	if sx.IsNil(form) {
		return NilExpr, nil
	}
	switch f := form.(type) {
	case *sx.Symbol:
		return ResolveExpr{Symbol: f}, nil
	case *sx.Pair:
		expr, err := dp.parsePair(eng, env, f)
		if err == nil {
			return expr, nil
		}
		if again, isAgain := err.(ErrParseAgain); isAgain {
			form, env = again.Form, again.Env
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

func (*defaultParser) parsePair(eng *Engine, env Environment, pair *sx.Pair) (Expr, error) {
	var proc Expr
	first := pair.Car()
	if sym, isSymbol := sx.GetSymbol(first); isSymbol {
		if val, found := Resolve(env, sym); found {
			if sp, isSpecial := GetSpecial(val); isSpecial {
				return sp.Parse(eng, env, pair.Tail())
			}
		}
		proc = ResolveExpr{Symbol: sym}
	} else {
		p, err := eng.Parse(env, first)
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
		expr, err2 := eng.Parse(env, elem.Car())
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
