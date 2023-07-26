//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package cond provides some special/builtin functions to conditionally evaluate values.
package cond

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// IfS parses an if-statement: (if cond then else). If else is missing, a nil is assumed.
func IfS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("requires 2 or 3 arguments, got none")
	}
	testExpr, err := eng.Parse(env, args.Car())
	if err != nil {
		return nil, err
	}
	argTrue := args.Tail()
	if argTrue == nil {
		return nil, fmt.Errorf("requires 2 or 3 arguments, got one")
	}
	trueExpr, err := eng.Parse(env, argTrue.Car())
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
	falseExpr, err := eng.Parse(env, argFalse.Car())
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
	Test eval.Expr
	True eval.Expr
}

func (ife *If2Expr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	test, err := eng.Execute(env, ife.Test)
	if err != nil {
		return nil, err
	}
	if sxpf.IsTrue(test) {
		return eng.ExecuteTCO(env, ife.True)
	}
	return sxpf.Nil(), nil
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
func (ife *If2Expr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	testExpr := ife.Test.Rework(ro, env)
	trueExpr := ife.True.Rework(ro, env)
	if objectExpr, isObjectExpr := testExpr.(eval.ObjectExpr); isObjectExpr {
		if sxpf.IsTrue(objectExpr.Object()) {
			return trueExpr
		}
		return eval.NilExpr.Rework(ro, env)
	}
	ife.Test = testExpr
	ife.True = trueExpr
	return ife
}

// IfExpr represents the if-then-else form.
type If3Expr struct {
	Test  eval.Expr
	True  eval.Expr
	False eval.Expr
}

func (ife *If3Expr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	test, err := eng.Execute(env, ife.Test)
	if err != nil {
		return nil, err
	}
	if sxpf.IsTrue(test) {
		return eng.ExecuteTCO(env, ife.True)
	}
	return eng.ExecuteTCO(env, ife.False)
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
func (ife *If3Expr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	testExpr := ife.Test.Rework(ro, env)
	trueExpr := ife.True.Rework(ro, env)
	falseExpr := ife.False.Rework(ro, env)

	// Check for constant condition
	if objectExpr, isObjectExpr := testExpr.(eval.ObjExpr); isObjectExpr {
		if sxpf.IsTrue(objectExpr.Object()) {
			return trueExpr
		}
		return falseExpr
	}

	// A nil false expression will result in a If2Expr.
	if objectExpr, isObjectExpr := falseExpr.(eval.ObjectExpr); isObjectExpr {
		if sxpf.IsNil(objectExpr.Object()) {
			if2expr := &If2Expr{
				Test: testExpr,
				True: trueExpr,
			}
			return if2expr.Rework(ro, env)
		}
	}

	ife.Test = testExpr
	ife.True = trueExpr
	ife.False = falseExpr
	return ife
}
