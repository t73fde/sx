//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

// Contains builtins and syntax for boolean values.

import (
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Boolean returns the given value interpreted as a boolean.
func Boolean(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsTrue(args[0])), nil
}

// Not negates the given value interpreted as a boolean.
func Not(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsFalse(args[0])), nil
}

// AndS parses an and statement: (and expr...).
func AndS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	es, err := ParseExprSeq(pf, args)
	if err != nil {
		return nil, err
	}
	if ex, ok := es.ParseRework(sxeval.ObjExpr{Obj: sx.MakeBoolean(true)}); ok {
		return ex, nil
	}
	return &AndExpr{es}, nil
}

// AndExpr represents the and form.
type AndExpr struct{ ExprSeq }

func (ae *AndExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	for i, expr := range ae.Front {
		front := expr.Rework(rf)
		if objectExpr, isObjectExpr := front.(sxeval.ObjectExpr); isObjectExpr {
			if sx.IsFalse(objectExpr.Object()) {
				return front
			}
		}
		ae.Front[i] = front
	}
	last := ae.Last.Rework(rf)
	if objectExpr, isObjectExpr := last.(sxeval.ObjectExpr); isObjectExpr {
		if sx.IsFalse(objectExpr.Object()) {
			return last
		}
	}
	ae.Last = last
	return ae
}
func (ae *AndExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	for _, e := range ae.Front {
		obj, err := frame.Execute(e)
		if err != nil {
			return nil, err
		}
		if sx.IsFalse(obj) {
			return obj, nil
		}
	}
	return frame.ExecuteTCO(ae.Last)
}
func (ae *AndExpr) IsEqual(other sxeval.Expr) bool {
	if ae == other {
		return true
	}
	if otherA, ok := other.(*AndExpr); ok && otherA != nil {
		return ae.ExprSeq.IsEqual(&otherA.ExprSeq)
	}
	return false
}
func (ae *AndExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{AND")
	if err != nil {
		return length, err
	}
	l, err := ae.ExprSeq.Print(w)
	length += l
	return length, err

}

// OrS parses an or statement: (or expr...).
func OrS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	es, err := ParseExprSeq(pf, args)
	if err != nil {
		return nil, err
	}
	if ex, ok := es.ParseRework(sxeval.NilExpr); ok {
		return ex, nil
	}
	return &OrExpr{es}, nil
}

// OrExpr represents the and form.
type OrExpr struct{ ExprSeq }

func (oe *OrExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	for i, expr := range oe.Front {
		front := expr.Rework(rf)
		if objectExpr, isObjectExpr := front.(sxeval.ObjectExpr); isObjectExpr {
			if sx.IsTrue(objectExpr.Object()) {
				return front
			}
		}
		oe.Front[i] = front
	}
	last := oe.Last.Rework(rf)
	if objectExpr, isObjectExpr := last.(sxeval.ObjectExpr); isObjectExpr {
		if sx.IsTrue(objectExpr.Object()) {
			return last
		}
	}
	oe.Last = last
	return oe
}
func (oe *OrExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	for _, e := range oe.Front {
		obj, err := frame.Execute(e)
		if err != nil {
			return nil, err
		}
		if sx.IsTrue(obj) {
			return obj, nil
		}
	}
	return frame.ExecuteTCO(oe.Last)
}
func (oe *OrExpr) IsEqual(other sxeval.Expr) bool {
	if oe == other {
		return true
	}
	if otherO, ok := other.(*OrExpr); ok && otherO != nil {
		return oe.ExprSeq.IsEqual(&otherO.ExprSeq)
	}
	return false
}
func (oe *OrExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{OR")
	if err != nil {
		return length, err
	}
	l, err := oe.ExprSeq.Print(w)
	length += l
	return length, err

}
