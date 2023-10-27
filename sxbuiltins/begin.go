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

// BeginS parses a sequence of expressions.
var BeginS = sxeval.Special{
	Name: "begin",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
		es, err := ParseExprSeq(pf, args)
		if err != nil {
			return nil, err
		}
		front, last := splitToFrontLast(es)
		if last == nil {
			return sxeval.NilExpr, nil
		}
		if len(front) == 0 {
			return last, nil
		}
		return &BeginExpr{Front: front, Last: last}, nil
	},
}

// ParseExprSeq parses a sequence of expressions.
func ParseExprSeq(pf *sxeval.ParseFrame, args *sx.Pair) ([]sxeval.Expr, error) {
	if args == nil {
		return nil, nil
	}
	var front []sxeval.Expr
	for node := args; ; {
		ex, err := pf.Parse(node.Car())
		if err != nil {
			return nil, err
		}
		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			return append(front, ex), nil
		}
		front = append(front, ex)
		if next, isPair := sx.GetPair(cdr); isPair {
			node = next
			continue
		}
		ex, err = pf.Parse(cdr)
		if err != nil {
			return nil, err
		}
		return append(front, ex), nil
	}
}

func splitToFrontLast(seq []sxeval.Expr) (front []sxeval.Expr, last sxeval.Expr) {
	switch l := len(seq); l {
	case 0:
		return nil, nil
	case 1:
		return nil, seq[0]
	default:
		return seq[0 : l-1], seq[l-1]
	}
}

// BeginExpr represents the begin form.
type BeginExpr struct {
	Front []sxeval.Expr
	Last  sxeval.Expr
}

func (be *BeginExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	for i, e := range be.Front {
		be.Front[i] = e.Rework(rf)
	}
	be.Last = be.Last.Rework(rf)
	return be
}

func (be *BeginExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	for _, e := range be.Front {
		subFrame := frame.MakeCalleeFrame()
		_, err := subFrame.Execute(e)
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
	if otherB, ok := other.(*BeginExpr); ok && otherB != nil {
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
	l, err := sxeval.PrintExprs(w, be.Front)
	length += l
	if err != nil {
		return length, err
	}
	l, err = be.Last.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
