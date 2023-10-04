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

// Provides functions to measure evaluation.

import (
	"fmt"
	"io"
	"time"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// TimeitS is a syntax to measure code execution time.
func TimeitS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("one argument expected")
	}
	expr, err := pf.Parse(args.Car())
	if err != nil {
		return nil, err
	}
	return &TimeitExpr{expr}, nil
}

// TimeitExpr stores information to measure execution time.
type TimeitExpr struct {
	expr sxeval.Expr
}

func (te *TimeitExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	te.expr = te.expr.Rework(rf)
	return te
}
func (te *TimeitExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	start := time.Now()
	subFrame := frame.MakeCalleeFrame()
	obj, err := subFrame.Execute(te.expr)
	duration := sx.String(time.Since(start).String())
	if err == nil {
		return sx.MakeList(duration, obj), nil
	}
	return sx.MakeList(duration, obj, sx.String(err.Error())), nil
}
func (te *TimeitExpr) IsEqual(other sxeval.Expr) bool {
	if te == other {
		return true
	}
	if otherT, ok := other.(*TimeitExpr); ok && otherT != nil {
		return te.expr.IsEqual(otherT.expr)
	}
	return false
}
func (te *TimeitExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{TIMEIT ")
	if err != nil {
		return length, err
	}
	l, err := te.expr.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
