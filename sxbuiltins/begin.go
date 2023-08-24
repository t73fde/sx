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
	front, last, err := ParseExprSeq(pf, args)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return sxeval.NilExpr, nil
	}
	if len(front) == 0 {
		return last, nil
	}
	return &BeginExpr{front, last}, nil
}

// BeginExpr represents the begin form.
type BeginExpr struct {
	Front []sxeval.Expr // all expressions, but the last
	Last  sxeval.Expr
}

func (be *BeginExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	for i, expr := range be.Front {
		be.Front[i] = expr.Rework(rf)
	}
	last := be.Last.Rework(rf)
	if len(be.Front) == 0 {
		return last
	}
	be.Last = last
	return be
}
func (be *BeginExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	for _, e := range be.Front {
		_, err := frame.Execute(e)
		if err != nil {
			return nil, err
		}
	}
	return frame.ExecuteTCO(be.Last)
}
func (be *BeginExpr) IsEqual(other sxeval.Expr) bool {
	if be == other {
		return true
	}
	if otherB, ok := other.(*LambdaExpr); ok && otherB != nil {
		return sxeval.EqualExprSlice(be.Front, otherB.Front) &&
			be.Last.IsEqual(otherB.Last)
	}
	return false
}
func (be *BeginExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{BEGIN")
	if err != nil {
		return length, err
	}
	l, err := sxeval.PrintFrontLast(w, be.Front, be.Last)
	length += l
	return length, err

}
