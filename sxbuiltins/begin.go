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

package sxbuiltins

import (
	"io"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// ExprSeq is a TCO-optimized sequence of `sxeval.Expr`.
type ExprSeq struct {
	Front []sxeval.Expr
	Last  sxeval.Expr
}

// ParseExprSeq parses a sequence of expressions.
func ParseExprSeq(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return sxeval.NilExpr, nil
	}
	var be BeginExpr
	if err := doParseExprSeq(pf, args, &be.ExprSeq); err != nil {
		return nil, err
	}
	if len(be.Front) == 0 {
		return be.Last, nil
	}
	return &be, nil
}

// ParseExprSeq parses a sequence of expressions.
func doParseExprSeq(pf *sxeval.ParseEnvironment, args *sx.Pair, es *ExprSeq) error {
	var front []sxeval.Expr
	var last sxeval.Expr
	for node := args; ; {
		ex, err := pf.Parse(node.Car())
		if err != nil {
			return err
		}
		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			last = ex
			break
		}
		front = append(front, ex)
		if next, isPair := sx.GetPair(cdr); isPair {
			node = next
			continue
		}
		ex, err = pf.Parse(cdr)
		if err != nil {
			return err
		}
		last = ex
		break
	}
	*es = ExprSeq{Front: front, Last: last}
	return nil
}

// IsPure signals an expression that has no side effects.
func (es *ExprSeq) IsPure() bool {
	for _, expr := range es.Front {
		if !expr.IsPure() {
			return false
		}
	}
	return es.Last.IsPure()
}

// Unparse the expression sequence as an sx.Object
func (es *ExprSeq) Unparse(sym *sx.Symbol) sx.Object {
	obj := es.Last.Unparse()
	for i := len(es.Front) - 1; i >= 0; i-- {
		obj = sx.Cons(es.Front[i].Unparse(), obj)
	}
	return sx.Cons(sym, obj)
}

// Print the expression on the given writer.
func (es *ExprSeq) Print(w io.Writer, prefix string) (int, error) {
	length, err := io.WriteString(w, prefix)
	if err != nil {
		return length, err
	}
	var l int
	for _, expr := range es.Front {
		l, err = io.WriteString(w, " ")
		length += l
		if err != nil {
			return length, err
		}
		l, err = expr.Print(w)
		length += l
		if err != nil {
			return length, err
		}
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = es.Last.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

// ----- (begin ...)
const beginName = "begin"

// BeginS parses a sequence of expressions.
var BeginS = sxeval.Special{
	Name: beginName,
	Fn:   ParseExprSeq,
}

// BeginExpr represents the begin form.
type BeginExpr struct{ ExprSeq }

// IsPure signals an expression that has no side effects.
func (be *BeginExpr) IsPure() bool { return be.ExprSeq.IsPure() }

// Unparse the expression as an sx.Object
func (be *BeginExpr) Unparse() sx.Object { return be.ExprSeq.Unparse(sx.MakeSymbol(beginName)) }

// Improve the expression into a possible simpler one.
func (be *BeginExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	last, err := imp.Improve(be.Last)
	if err != nil {
		return be, err
	}
	frontLen := len(be.Front)
	if frontLen == 0 {
		return last, nil
	}
	if err = imp.ImproveSlice(be.Front); err != nil {
		return be, err
	}

	seq := make([]sxeval.Expr, 0, frontLen)
	for _, expr := range be.Front {
		if !expr.IsPure() {
			seq = append(seq, expr)
		}
	}
	if seqLen := len(seq); seqLen == 0 {
		return last, nil
	} else if seqLen == cap(be.Front) {
		copy(be.Front, seq)
	} else {
		newFront := make([]sxeval.Expr, seqLen)
		copy(newFront, seq)
		be.Front = newFront
	}
	be.Last = last
	return be, nil
}

// Compute the expression in a frame and return the result.
func (be *BeginExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	for _, e := range be.Front {
		if _, err := env.Execute(e, bind); err != nil {
			return nil, err
		}
	}
	return env.ExecuteTCO(be.Last, bind)
}

// Print the expression on the given writer.
func (be *BeginExpr) Print(w io.Writer) (int, error) { return be.ExprSeq.Print(w, "{BEGIN") }

// ----- (and ...)

const andName = "and"

// AndS parses a sequence of expressions that are reduced via logical and.
var AndS = sxeval.Special{
	Name: andName,
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		if args == nil {
			return sxeval.ObjExpr{Obj: sx.T}, nil
		}
		var ae AndExpr
		if err := doParseExprSeq(pe, args, &ae.ExprSeq); err != nil {
			return nil, err
		}
		return &ae, nil
	},
}

// AndExpr represents the and form.
type AndExpr struct{ ExprSeq }

// IsPure signals an expression that has no side effects.
func (ae *AndExpr) IsPure() bool { return ae.ExprSeq.IsPure() }

// Unparse the expression as an sx.Object
func (ae *AndExpr) Unparse() sx.Object { return ae.ExprSeq.Unparse(sx.MakeSymbol(andName)) }

// Improve the expression into a possible simpler one.
func (ae *AndExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	last, err := imp.Improve(ae.Last)
	if err != nil {
		return ae, err
	}
	frontLen := len(ae.Front)
	if frontLen == 0 {
		return last, nil
	}
	if err = imp.ImproveSlice(ae.Front); err != nil {
		return ae, err
	}

	for _, expr := range ae.Front {
		if objectExpr, isConstObject := sxeval.GetConstExpr(expr); isConstObject {
			if sx.IsFalse(objectExpr.ConstObject()) {
				return expr, nil
			}
			// TODO: ignore True values
		}
	}
	ae.Last = last
	return ae, nil
}

// Compute the expression in a frame and return the result.
func (ae *AndExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	for _, e := range ae.Front {
		obj, err := env.Execute(e, bind)
		if err != nil {
			return nil, err
		}
		if sx.IsFalse(obj) {
			return obj, nil
		}
	}
	return env.ExecuteTCO(ae.Last, bind)
}

// Print the expression on the given writer.
func (ae *AndExpr) Print(w io.Writer) (int, error) { return ae.ExprSeq.Print(w, "{AND") }

// ----- (or ...)

const orName = "or"

// OrS parses a sequence of expressions that are reduced via logical or.
var OrS = sxeval.Special{
	Name: orName,
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		if args == nil {
			return sxeval.NilExpr, nil
		}
		var oe OrExpr
		if err := doParseExprSeq(pe, args, &oe.ExprSeq); err != nil {
			return nil, err
		}
		return &oe, nil
	},
}

// OrExpr represents the or form.
type OrExpr struct{ ExprSeq }

// IsPure signals an expression that has no side effects.
func (oe *OrExpr) IsPure() bool { return oe.ExprSeq.IsPure() }

// Unparse the expression as an sx.Object
func (oe *OrExpr) Unparse() sx.Object { return oe.ExprSeq.Unparse(sx.MakeSymbol(orName)) }

// Improve the expression into a possible simpler one.
func (oe *OrExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	last, err := imp.Improve(oe.Last)
	if err != nil {
		return oe, err
	}
	frontLen := len(oe.Front)
	if frontLen == 0 {
		return last, nil
	}
	if err = imp.ImproveSlice(oe.Front); err != nil {
		return oe, err
	}

	for _, expr := range oe.Front {
		if objectExpr, isConstObject := sxeval.GetConstExpr(expr); isConstObject {
			if sx.IsTrue(objectExpr.ConstObject()) {
				return expr, nil
			}
			// TODO: ignore False values
		}
	}
	oe.Last = last
	return oe, nil
}

// Compute the expression in a frame and return the result.
func (oe *OrExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	for _, e := range oe.Front {
		obj, err := env.Execute(e, bind)
		if err != nil {
			return nil, err
		}
		if sx.IsTrue(obj) {
			return obj, nil
		}
	}
	return env.ExecuteTCO(oe.Last, bind)
}

// Print the expression on the given writer.
func (oe *OrExpr) Print(w io.Writer) (int, error) { return oe.ExprSeq.Print(w, "{OR") }
