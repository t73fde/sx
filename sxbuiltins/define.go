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

// Contains all syntaxes and builtins to bind values to symbols.

import (
	"fmt"
	"io"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

const defvarName = "defvar"

// DefVarS parses a (defvar name value) form.
var DefVarS = sxeval.Special{
	Name: defvarName,
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, frame *sxeval.Frame) (sxeval.Expr, error) {
		sym, val, err := parseSymValue(pe, args, frame)
		if err != nil {
			return nil, err
		}
		return &DefineExpr{Sym: sym, Val: val}, nil
	},
}

func parseSymValue(pe *sxeval.ParseEnvironment, args *sx.Pair, frame *sxeval.Frame) (*sx.Symbol, sxeval.Expr, error) {
	if args == nil {
		return nil, nil, fmt.Errorf("need at least two arguments")
	}
	car := args.Car()
	sym, isSymbol := sx.GetSymbol(car)
	if !isSymbol {
		return nil, nil, fmt.Errorf("argument 1 must be a symbol, but is: %T/%v", car, car)
	}
	cdr := args.Cdr()
	if sx.IsNil(cdr) {
		return nil, nil, fmt.Errorf("argument 2 missing")
	}
	pair, isPair := sx.GetPair(cdr)
	if !isPair {
		return nil, nil, fmt.Errorf("argument 2 must be a proper list")
	}
	val, err := pe.Parse(pair.Car(), frame)
	return sym, val, err
}

// DefineExpr stores data for a define statement.
type DefineExpr struct {
	Sym *sx.Symbol
	Val sxeval.Expr
}

// IsPure signals an expression that has no side effects.
func (*DefineExpr) IsPure() bool { return false }

// Unparse the expression as an sx.Object
func (de *DefineExpr) Unparse() sx.Object {
	return sx.MakeList(sx.MakeSymbol(defvarName), de.Sym, de.Val.Unparse())
}

// Improve the expression into a possible simpler one.
func (de *DefineExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	imp.Bind(de.Sym)
	expr, err := imp.Improve(de.Val)
	if err == nil {
		de.Val = expr
	}
	return de, err
}

// Compute the expression in a frame and return the result.
func (de *DefineExpr) Compute(env *sxeval.Environment, frame *sxeval.Frame) (sx.Object, error) {
	val, err := env.Execute(de.Val, frame)
	if err == nil {
		frame.Bind(de.Sym, val)
		return val, nil
	}
	return nil, err
}

// Print the expression on the given writer.
func (de *DefineExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{DEFINE ")
	if err != nil {
		return length, err
	}
	l, err := sx.Print(w, de.Sym)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = de.Val.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

const setXName = "set!"

// SetXS parses a (set! name value) form.
var SetXS = sxeval.Special{
	Name: setXName,
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, frame *sxeval.Frame) (sxeval.Expr, error) {
		if args == nil {
			return nil, fmt.Errorf("need at least two arguments")
		}
		car := args.Car()
		sym, ok := sx.GetSymbol(car)
		if !ok {
			return nil, fmt.Errorf("argument 1 must be a symbol, but is: %T/%v", car, car)
		}
		cdr := args.Cdr()
		if sx.IsNil(cdr) {
			return nil, fmt.Errorf("argument 2 missing")
		}
		pair, isPair := sx.GetPair(cdr)
		if !isPair {
			return nil, fmt.Errorf("argument 2 must be a proper list, but is: %T/%v", cdr, cdr)
		}
		val, err := pe.Parse(pair.Car(), frame)
		if err != nil {
			return val, err
		}
		return &SetXExpr{Sym: sym, Val: val}, nil
	},
}

// SetXExpr stores data for a set! statement.
type SetXExpr struct {
	Sym *sx.Symbol
	Val sxeval.Expr
}

// IsPure signals an expression that has no side effects.
func (*SetXExpr) IsPure() bool { return false }

// Unparse the expression as an sx.Object
func (se *SetXExpr) Unparse() sx.Object {
	return sx.MakeList(sx.MakeSymbol(setXName), se.Sym, se.Val.Unparse())
}

// Improve the expression into a possible simpler one.
func (se *SetXExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	expr, err := imp.Improve(se.Val)
	if err == nil {
		se.Val = expr
	}
	return se, err
}

// Compute the expression in a frame and return the result.
func (se *SetXExpr) Compute(env *sxeval.Environment, frame *sxeval.Frame) (sx.Object, error) {
	sym := se.Sym
	fr := frame.FindFrame(sym)
	if fr == nil {
		return nil, frame.MakeNotBoundError(sym)
	}
	val, err := env.Execute(se.Val, frame)
	if err == nil {
		err = fr.Bind(sym, val)
	}
	return val, err
}

// Print the expression on the given writer.
func (se *SetXExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{SET! ")
	if err != nil {
		return length, err
	}
	l, err := sx.Print(w, se.Sym)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = se.Val.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
