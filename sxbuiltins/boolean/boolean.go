//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package boolean contains builtins and syntax for boolean values.
package boolean

import (
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
)

// BooleanP is the boolean that returns true if the argument is a number.
func BooleanP(args []sx.Object) (sx.Object, error) {
	if err := sxbuiltins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sx.GetBoolean(args[0])
	return sx.MakeBoolean(ok), nil
}

// Boolean negates the given value interpreted as a boolean.
func Boolean(args []sx.Object) (sx.Object, error) {
	if err := sxbuiltins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsTrue(args[0])), nil
}

// Not negates the given value interpreted as a boolean.
func Not(args []sx.Object) (sx.Object, error) {
	if err := sxbuiltins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsFalse(args[0])), nil
}

// AndS parses an and statement: (and expr...).
func AndS(eng *sxeval.Engine, env sx.Environment, args *sx.Pair) (sxeval.Expr, error) {
	front, last, err := sxbuiltins.ParseExprSeq(eng, env, args)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return sxeval.TrueExpr, nil
	}
	if len(front) == 0 {
		return last, nil
	}
	return &AndExpr{front, last}, nil
}

// AndExpr represents the and form.
type AndExpr struct {
	Front []sxeval.Expr // all expressions, but the last
	Last  sxeval.Expr
}

func (ae *AndExpr) Compute(eng *sxeval.Engine, env sx.Environment) (sx.Object, error) {
	for _, e := range ae.Front {
		obj, err := eng.Execute(env, e)
		if err != nil {
			return nil, err
		}
		if sx.IsFalse(obj) {
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
	l, err := sxeval.PrintFrontLast(w, ae.Front, ae.Last)
	length += l
	return length, err

}
func (ae *AndExpr) Rework(ro *sxeval.ReworkOptions, env sx.Environment) sxeval.Expr {
	for i, expr := range ae.Front {
		front := expr.Rework(ro, env)
		if objectExpr, isObjectExpr := front.(sxeval.ObjectExpr); isObjectExpr {
			if sx.IsFalse(objectExpr.Object()) {
				return front
			}
		}
		ae.Front[i] = front
	}
	last := ae.Last.Rework(ro, env)
	if objectExpr, isObjectExpr := last.(sxeval.ObjectExpr); isObjectExpr {
		if sx.IsFalse(objectExpr.Object()) {
			return last
		}
	}
	ae.Last = last
	return ae
}

// OrS parses an or statement: (or expr...).
func OrS(eng *sxeval.Engine, env sx.Environment, args *sx.Pair) (sxeval.Expr, error) {
	front, last, err := sxbuiltins.ParseExprSeq(eng, env, args)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return sxeval.FalseExpr, nil
	}
	if len(front) == 0 {
		return last, nil
	}
	return &OrExpr{front, last}, nil
}

// OrExpr represents the and form.
type OrExpr struct {
	Front []sxeval.Expr // all expressions, but the last
	Last  sxeval.Expr
}

func (oe *OrExpr) Compute(eng *sxeval.Engine, env sx.Environment) (sx.Object, error) {
	for _, e := range oe.Front {
		obj, err := eng.Execute(env, e)
		if err != nil {
			return nil, err
		}
		if sx.IsTrue(obj) {
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
	l, err := sxeval.PrintFrontLast(w, oe.Front, oe.Last)
	length += l
	return length, err

}
func (oe *OrExpr) Rework(ro *sxeval.ReworkOptions, env sx.Environment) sxeval.Expr {
	for i, expr := range oe.Front {
		front := expr.Rework(ro, env)
		if objectExpr, isObjectExpr := front.(sxeval.ObjectExpr); isObjectExpr {
			if sx.IsTrue(objectExpr.Object()) {
				return front
			}
		}
		oe.Front[i] = front
	}
	last := oe.Last.Rework(ro, env)
	if objectExpr, isObjectExpr := last.(sxeval.ObjectExpr); isObjectExpr {
		if sx.IsTrue(objectExpr.Object()) {
			return last
		}
	}
	oe.Last = last
	return oe
}
