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
	var es ExprSeq
	if err := doParseExprSeq(pf, args, &es); err != nil {
		return nil, err
	}
	if len(es.Front) == 0 {
		return es.Last, nil
	}
	return &BeginExpr{ExprSeq: es}, nil
}

// ParseExprSeq parses a sequence of expressions.
func doParseExprSeq(pf *sxeval.ParseEnvironment, args *sx.Pair, es *ExprSeq) error {
	if args == nil {
		es.Front = nil
		es.Last = sxeval.NilExpr
		return nil
	}
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
		if _, isConstObject := expr.(sxeval.ConstObjectExpr); isConstObject {
			// A constant object has no side effect, it can be ignored in the sequence
			continue
		}
		if _, isSymbol := expr.(sxeval.SymbolExpr); isSymbol {
			// A symbol has no side effect, it can be ignored in the sequence
			continue
		}
		seq = append(seq, expr)
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
func (be *BeginExpr) Print(w io.Writer) (int, error) { return be.ExprSeq.Print(w, "{BEGIN}") }
