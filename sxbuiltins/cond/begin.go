//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package cond

import (
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
)

// BeginS parses a begin-statement: (begin expr...).
func BeginS(frame *sxeval.Frame, args *sx.Pair) (sxeval.Expr, error) {
	front, last, err := sxbuiltins.ParseExprSeq(frame, args)
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

func (be *BeginExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	for _, e := range be.Front {
		_, err := frame.Execute(e)
		if err != nil {
			return nil, err
		}
	}
	return frame.ExecuteTCO(be.Last)
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
func (be *BeginExpr) Rework(ro *sxeval.ReworkOptions, env sxeval.Environment) sxeval.Expr {
	for i, expr := range be.Front {
		be.Front[i] = expr.Rework(ro, env)
	}
	be.Last = be.Last.Rework(ro, env)
	return be
}
