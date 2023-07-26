//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package boolean contains builtins and syntax for boolean values.
package boolean

import (
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// BooleanP is the boolean that returns true if the argument is a number.
func BooleanP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sxpf.GetBoolean(args[0])
	return sxpf.MakeBoolean(ok), nil
}

// Boolean negates the given value interpreted as a boolean.
func Boolean(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(sxpf.IsTrue(args[0])), nil
}

// Not negates the given value interpreted as a boolean.
func Not(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(sxpf.IsFalse(args[0])), nil
}

// AndS parses an and statement: (and expr...).
func AndS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	front, last, err := builtins.ParseExprSeq(eng, env, args)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return eval.TrueExpr, nil
	}
	if len(front) == 0 {
		return last, nil
	}
	return &AndExpr{front, last}, nil
}

// AndExpr represents the and form.
type AndExpr struct {
	Front []eval.Expr // all expressions, but the last
	Last  eval.Expr
}

func (ae *AndExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	for _, e := range ae.Front {
		obj, err := eng.Execute(env, e)
		if err != nil {
			return nil, err
		}
		if sxpf.IsFalse(obj) {
			return obj, nil
		}
	}
	return eng.ExecuteTCO(env, ae.Last)
}
func (ae *AndExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{AND")
	if err != nil {
		return length, err
	}
	l, err := eval.PrintFrontLast(w, ae.Front, ae.Last)
	length += l
	return length, err

}
func (ae *AndExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	for i, expr := range ae.Front {
		front := expr.Rework(ro, env)
		if objectExpr, isObjectExpr := front.(eval.ObjectExpr); isObjectExpr {
			if sxpf.IsFalse(objectExpr.Object()) {
				return front
			}
		}
		ae.Front[i] = front
	}
	last := ae.Last.Rework(ro, env)
	if objectExpr, isObjectExpr := last.(eval.ObjectExpr); isObjectExpr {
		if sxpf.IsFalse(objectExpr.Object()) {
			return last
		}
	}
	ae.Last = last
	return ae
}

// OrS parses an or statement: (or expr...).
func OrS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	front, last, err := builtins.ParseExprSeq(eng, env, args)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return eval.FalseExpr, nil
	}
	if len(front) == 0 {
		return last, nil
	}
	return &OrExpr{front, last}, nil
}

// OrExpr represents the and form.
type OrExpr struct {
	Front []eval.Expr // all expressions, but the last
	Last  eval.Expr
}

func (oe *OrExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	for _, e := range oe.Front {
		obj, err := eng.Execute(env, e)
		if err != nil {
			return nil, err
		}
		if sxpf.IsTrue(obj) {
			return obj, nil
		}
	}
	return eng.ExecuteTCO(env, oe.Last)
}
func (oe *OrExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{OR")
	if err != nil {
		return length, err
	}
	l, err := eval.PrintFrontLast(w, oe.Front, oe.Last)
	length += l
	return length, err

}
func (oe *OrExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	for i, expr := range oe.Front {
		front := expr.Rework(ro, env)
		if objectExpr, isObjectExpr := front.(eval.ObjectExpr); isObjectExpr {
			if sxpf.IsTrue(objectExpr.Object()) {
				return front
			}
		}
		oe.Front[i] = front
	}
	last := oe.Last.Rework(ro, env)
	if objectExpr, isObjectExpr := last.(eval.ObjectExpr); isObjectExpr {
		if sxpf.IsTrue(objectExpr.Object()) {
			return last
		}
	}
	oe.Last = last
	return oe
}
