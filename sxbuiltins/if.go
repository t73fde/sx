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
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// IfS parses an if-statement: (if cond then else). If else is missing, a nil is assumed.
var IfS = sxeval.Syntax{
	Name: "if",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
		if args == nil {
			return nil, fmt.Errorf("requires 2 or 3 arguments, got none")
		}
		testExpr, err := pf.Parse(args.Car())
		if err != nil {
			return nil, err
		}
		argTrue := args.Tail()
		if argTrue == nil {
			return nil, fmt.Errorf("requires 2 or 3 arguments, got one")
		}
		trueExpr, err := pf.Parse(argTrue.Car())
		if err != nil {
			return nil, err
		}
		argFalse := argTrue.Tail()
		if argFalse == nil {
			return &IfExpr{
				Test:  testExpr,
				True:  trueExpr,
				False: sxeval.NilExpr,
			}, nil
		}
		if argFalse.Tail() != nil {
			return nil, fmt.Errorf("requires 2 or 3 arguments, got more")
		}
		falseExpr, err := pf.Parse(argFalse.Car())
		if err != nil {
			return nil, err
		}
		return &IfExpr{
			Test:  testExpr,
			True:  trueExpr,
			False: falseExpr,
		}, nil
	},
}

// IfExpr represents the if-then-else form.
type IfExpr struct {
	Test  sxeval.Expr
	True  sxeval.Expr
	False sxeval.Expr
}

func (ife *IfExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	testExpr := ife.Test.Rework(rf)
	trueExpr := ife.True.Rework(rf)
	falseExpr := ife.False.Rework(rf)

	// Check for constant condition
	if objectExpr, isObjectExpr := testExpr.(sxeval.ObjExpr); isObjectExpr {
		if sx.IsTrue(objectExpr.Object()) {
			return trueExpr
		}
		return falseExpr
	}

	ife.Test = testExpr
	ife.True = trueExpr
	ife.False = falseExpr
	return ife
}
func (ife *IfExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	subFrame := frame.MakeCalleeFrame()
	test, err := subFrame.Execute(ife.Test)
	if err != nil {
		return nil, err
	}
	if sx.IsTrue(test) {
		return frame.ExecuteTCO(ife.True)
	}
	return frame.ExecuteTCO(ife.False)
}
func (ife *IfExpr) IsEqual(other sxeval.Expr) bool {
	if ife == other {
		return true
	}
	if otherI, ok := other.(*IfExpr); ok && otherI != nil {
		return ife.Test.IsEqual(otherI.Test) && ife.True.IsEqual(otherI.True) && ife.False.IsEqual(otherI.False)
	}
	return false
}

func (ife *IfExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{IF ")
	if err != nil {
		return length, err
	}
	l, err := ife.Test.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = ife.True.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = ife.False.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
