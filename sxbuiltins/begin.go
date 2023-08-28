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

import (
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// BeginS parses a begin-statement: (begin expr...).
func BeginS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	es, err := ParseExprSeq(pf, args)
	if err != nil {
		return nil, err
	}
	if ex, ok := es.ParseRework(sxeval.NilExpr); ok {
		return ex, nil
	}
	return &BeginExpr{es}, nil
}

// BeginExpr represents the begin form.
type BeginExpr struct{ ExprSeq }

func (be *BeginExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	be.ExprSeq.Rework(rf)
	return be
}
func (be *BeginExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	return be.ExprSeq.Compute(frame)
}
func (be *BeginExpr) IsEqual(other sxeval.Expr) bool {
	if be == other {
		return true
	}
	if otherB, ok := other.(*BeginExpr); ok && otherB != nil {
		return be.ExprSeq.IsEqual(&otherB.ExprSeq)
	}
	return false
}
func (be *BeginExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{BEGIN")
	if err != nil {
		return length, err
	}
	l, err := be.ExprSeq.Print(w)
	length += l
	return length, err

}
