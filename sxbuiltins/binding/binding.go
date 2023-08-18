//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package binding contains builtins and syntax to bind values.
package binding

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxbuiltins/callable"
	"zettelstore.de/sx.fossil/sxbuiltins/cond"
	"zettelstore.de/sx.fossil/sxeval"
)

// LetS parses the `(let (binding...) expr...)` syntax.`
func LetS(eng *sxeval.Engine, env sxeval.Environment, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("binding spec and body missing")
	}
	bindings, isPair := sx.GetPair(args.Car())
	if !isPair {
		return nil, fmt.Errorf("binding spec must be a list, but got: %t/%v", args.Car(), args.Car())
	}
	body, isPair := sx.GetPair(args.Cdr())
	if !isPair {
		return nil, sx.ErrImproper{Pair: args}
	}
	if bindings == nil {
		return cond.BeginS(eng, env, body)
	}
	letExpr := LetExpr{
		Symbols: nil,
		Expr:    nil,
		Front:   nil,
		Last:    nil,
	}
	letEnv := sxeval.MakeChildEnvironment(env, "let-def", 128)
	for node := bindings; node != nil; {
		sym, err := callable.GetParameterSymbol(letExpr.Symbols, node.Car())
		if err != nil {
			return nil, err
		}
		next, isPair2 := sx.GetPair(node.Cdr())
		if !isPair2 {
			return nil, sx.ErrImproper{Pair: bindings}
		}
		if next == nil {
			return nil, fmt.Errorf("binding missing for %v", sym)
		}
		expr, err := eng.Parse(env, next.Car())
		if err != nil {
			return nil, err
		}
		err = letEnv.Bind(sym, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
		letExpr.Symbols = append(letExpr.Symbols, sym)
		letExpr.Expr = append(letExpr.Expr, expr)
		next, isPair2 = sx.GetPair(next.Cdr())
		if !isPair2 {
			return nil, sx.ErrImproper{Pair: bindings}
		}
		node = next
	}

	front, last, err := sxbuiltins.ParseExprSeq(eng, letEnv, body)
	if err != nil {
		return nil, err
	}
	letExpr.Front = front
	letExpr.Last = last
	return &letExpr, nil
}

// LetExpr stores everything for a (let ...) expression.
type LetExpr struct {
	Symbols []*sx.Symbol
	Expr    []sxeval.Expr
	Front   []sxeval.Expr
	Last    sxeval.Expr
}

func (le *LetExpr) Compute(eng *sxeval.Engine, env sxeval.Environment) (sx.Object, error) {
	letEnv := sxeval.MakeChildEnvironment(env, "let", len(le.Symbols))
	for i, sym := range le.Symbols {
		obj, err := eng.Execute(env, le.Expr[i])
		if err != nil {
			return nil, err
		}
		err = letEnv.Bind(sym, obj)
		if err != nil {
			return nil, err
		}
	}
	for _, expr := range le.Front {
		_, err := eng.Execute(letEnv, expr)
		if err != nil {
			return nil, err
		}
	}
	return eng.ExecuteTCO(letEnv, le.Last)
}

func (le *LetExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{LET")
	if err != nil {
		return length, err
	}
	for i, sym := range le.Symbols {
		l, err2 := io.WriteString(w, " ")
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = io.WriteString(w, sym.Repr())
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = io.WriteString(w, " ")
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = le.Expr[i].Print(w)
		length += l
		if err2 != nil {
			return length, err2
		}
	}
	l, err := sxeval.PrintFrontLast(w, le.Front, le.Last)
	length += l
	return length, err
}

func (le *LetExpr) Rework(ro *sxeval.ReworkOptions, env sxeval.Environment) sxeval.Expr {
	for i, expr := range le.Expr {
		le.Expr[i] = expr.Rework(ro, env)
	}
	for i, expr := range le.Front {
		le.Front[i] = expr.Rework(ro, env)
	}
	le.Last = le.Last.Rework(ro, env)
	return le
}
