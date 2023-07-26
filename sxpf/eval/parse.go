//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package eval

import (
	"fmt"

	"zettelstore.de/sx.fossil/sxpf"
)

// Parser transform an object into an executable expression.
type Parser interface {
	Parse(*Engine, sxpf.Environment, sxpf.Object) (Expr, error)
}

// ErrParseAgain is a non-error error signalling that the given form should be
// parsed again in the given environment.
type ErrParseAgain struct {
	Env  sxpf.Environment
	Form sxpf.Object
}

func (e ErrParseAgain) Error() string { return fmt.Sprintf("Again: %T/%v", e.Form, e.Form) }

// defaultParser is the parser for normal use.
type defaultParser struct{}

var myDefaultParser defaultParser

func (dp *defaultParser) Parse(eng *Engine, env sxpf.Environment, form sxpf.Object) (Expr, error) {
restart:
	if sxpf.IsNil(form) {
		return NilExpr, nil
	}
	switch f := form.(type) {
	case *sxpf.Symbol:
		return ResolveExpr{Symbol: f}, nil
	case *sxpf.Pair:
		expr, err := dp.parsePair(eng, env, f)
		if err == nil {
			return expr, nil
		}
		if again, isAgain := err.(ErrParseAgain); isAgain {
			form, env = again.Form, again.Env
			goto restart
		}
		return nil, err
	case sxpf.Boolean:
		if f == sxpf.False {
			return FalseExpr, nil
		}
		return TrueExpr, nil
	}
	return ObjExpr{Obj: form}, nil
}

func (*defaultParser) parsePair(eng *Engine, env sxpf.Environment, pair *sxpf.Pair) (Expr, error) {
	var proc Expr
	first := pair.Car()
	if sym, isSymbol := sxpf.GetSymbol(first); isSymbol {
		if val, found := sxpf.Resolve(env, sym); found {
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
		if sxpf.IsNil(arg) {
			break
		}
		elem, isPair := sxpf.GetPair(arg)
		if !isPair {
			return nil, sxpf.ErrImproper{Pair: pair}
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
