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
func IfS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
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
		return &If2Expr{
			Test: testExpr,
			True: trueExpr,
		}, nil
	}
	if argFalse.Tail() != nil {
		return nil, fmt.Errorf("requires 2 or 3 arguments, got more")
	}
	falseExpr, err := pf.Parse(argFalse.Car())
	if err != nil {
		return nil, err
	}
	return &If3Expr{
		Test:  testExpr,
		True:  trueExpr,
		False: falseExpr,
	}, nil
}

// IfExpr represents the if-then-else form.
type If2Expr struct {
	Test sxeval.Expr
	True sxeval.Expr
}

func (ife *If2Expr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	testExpr := ife.Test.Rework(rf)
	trueExpr := ife.True.Rework(rf)
	if objectExpr, isObjectExpr := testExpr.(sxeval.ObjectExpr); isObjectExpr {
		if sx.IsTrue(objectExpr.Object()) {
			return trueExpr
		}
		return sxeval.NilExpr.Rework(rf)
	}
	ife.Test = testExpr
	ife.True = trueExpr
	return ife
}
func (ife *If2Expr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	test, err := frame.Execute(ife.Test)
	if err != nil {
		return nil, err
	}
	if sx.IsTrue(test) {
		return frame.ExecuteTCO(ife.True)
	}
	return sx.Nil(), nil
}
func (ife *If2Expr) IsEqual(other sxeval.Expr) bool {
	if ife == other {
		return true
	}
	if otherI, ok := other.(*If2Expr); ok && otherI != nil {
		return ife.Test.IsEqual(otherI.Test) && ife.True.IsEqual(otherI.True)
	}
	return false
}

func (ife *If2Expr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{IF2 ")
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
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

// IfExpr represents the if-then-else form.
type If3Expr struct {
	Test  sxeval.Expr
	True  sxeval.Expr
	False sxeval.Expr
}

func (ife *If3Expr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
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

	// A nil false expression will result in a If2Expr.
	if objectExpr, isObjectExpr := falseExpr.(sxeval.ObjectExpr); isObjectExpr {
		if sx.IsNil(objectExpr.Object()) {
			if2expr := &If2Expr{
				Test: testExpr,
				True: trueExpr,
			}
			return if2expr.Rework(rf)
		}
	}

	ife.Test = testExpr
	ife.True = trueExpr
	ife.False = falseExpr
	return ife
}
func (ife *If3Expr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	test, err := frame.Execute(ife.Test)
	if err != nil {
		return nil, err
	}
	if sx.IsTrue(test) {
		return frame.ExecuteTCO(ife.True)
	}
	return frame.ExecuteTCO(ife.False)
}
func (ife *If3Expr) IsEqual(other sxeval.Expr) bool {
	if ife == other {
		return true
	}
	if otherI, ok := other.(*If3Expr); ok && otherI != nil {
		return ife.Test.IsEqual(otherI.Test) && ife.True.IsEqual(otherI.True) && ife.False.IsEqual(otherI.False)
	}
	return false
}

func (ife *If3Expr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{IF3 ")
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
