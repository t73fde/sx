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
	"fmt"
	"io"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

const ifName = "if"

// IfS parses an if-statement: (if cond then else). If else is missing, a nil is assumed.
var IfS = sxeval.Special{
	Name: ifName,
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
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

// IsPure signals an expression that has no side effects.
func (ife *IfExpr) IsPure() bool {
	return ife.Test.IsPure() && ife.True.IsPure() && ife.False.IsPure()
}

// Unparse the expression as an sx.Object
func (ife *IfExpr) Unparse() sx.Object {
	return sx.MakeList(sx.MakeSymbol(ifName), ife.Test.Unparse(), ife.True.Unparse(), ife.False.Unparse())
}

// Improve the expression into a possible simpler one.
func (ife *IfExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	testExpr, err := imp.Improve(ife.Test)
	if err != nil {
		return ife, err
	}
	trueExpr, err := imp.Improve(ife.True)
	if err != nil {
		return ife, err
	}
	falseExpr, err := imp.Improve(ife.False)
	if err != nil {
		return ife, err
	}

restart:
	// Check for nested IfExpr in testExpr
	//
	// This might occur when macros are used, e.g.
	// "(if (!= a b) c d)"
	// --> "(if (not (= a b)) c d)"
	// --> "(if (if (= a b) NIL T) c d)"
	// =HERE=> "(if (= a b) d c)"
	//
	// trueExpr and falseExpr of nested IfExpr must be checked:
	// * if one is not a constant, ignore this optimization
	// * if both are true, evaluate to c
	// * if both are false, evaluate do d
	// * if trueExpr is true (and falseExpr is false), lift embedded testExpr up
	// * id trueExpr is false (and falseExpr is true), lift embedded testExpr up and switch c and d
	if nestedIfe, isIfExpr := testExpr.(*IfExpr); isIfExpr {
		nestedTrueExpr, hasTrue := sxeval.GetConstExpr(nestedIfe.True)
		nestedFalseExpr, hasFalse := sxeval.GetConstExpr(nestedIfe.False)
		if hasTrue && hasFalse {
			trueIsTrue := sx.IsTrue(nestedTrueExpr.ConstObject())
			falseIsTrue := sx.IsTrue(nestedFalseExpr.ConstObject())
			switch {
			case trueIsTrue && falseIsTrue:
				return trueExpr, nil
			case !trueIsTrue && !falseIsTrue:
				return falseExpr, nil
			case trueIsTrue && !falseIsTrue:
				testExpr = nestedIfe.Test
			default /* !trueIsTrue && falseIsTrue */ :
				testExpr = nestedIfe.Test
				trueExpr, falseExpr = falseExpr, trueExpr
			}
		}
	}

	// Check for constant condition
	if objectExpr, isConstObject := sxeval.GetConstExpr(testExpr); isConstObject {
		if sx.IsTrue(objectExpr.ConstObject()) {
			return trueExpr, nil
		}
		return falseExpr, nil
	}

	// Optimize (if (not X) Y Z) ==> (if X Y Z) and restart (X may match (not E); may match nested if)
	if bce, isBCE := testExpr.(*sxeval.BuiltinCall1Expr); isBCE && bce.Proc == &Not {
		testExpr, trueExpr, falseExpr = bce.Arg, falseExpr, trueExpr
		goto restart
	}

	ife.Test = testExpr
	ife.True = trueExpr
	ife.False = falseExpr
	return ife, nil
}

// Compile the expression.
func (ife *IfExpr) Compile(sxc *sxeval.Compiler, tailPos bool) error {

	var condPatch func()
	andExpr, isIfAnd := ife.Test.(*AndExpr)
	if isIfAnd {
		// If ife.Test is an (and a b ...), the "jmpPatches" of the compilation of
		// (and) should point to ife.False. This removes execution of JumpPopFalse
		// if one but the last element of (and) is False.
		var err error
		condPatch, err = andExpr.compileForIf(sxc)
		if err != nil {
			return err
		}
	} else {
		if err := sxc.Compile(ife.Test, false); err != nil {
			return err
		}
		condPatch = sxc.EmitJumpPopFalse()
	}

	if err := sxc.Compile(ife.True, tailPos); err != nil {
		return err
	}
	jumpPatch := sxc.EmitJump()
	condPatch()
	if err := sxc.Compile(ife.False, tailPos); err != nil {
		return err
	}
	sxc.AdjustStack(-1) // Otherwise true and false branch will be counted double
	jumpPatch()
	return nil
}

// Compute the expression in a frame and return the result.
func (ife *IfExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	test, err := env.Execute(ife.Test, bind)
	if err != nil {
		return nil, err
	}
	if sx.IsTrue(test) {
		return env.ExecuteTCO(ife.True, bind)
	}
	return env.ExecuteTCO(ife.False, bind)
}

// Print the expression on the given writer.
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
