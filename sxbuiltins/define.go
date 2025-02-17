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
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		sym, val, err := parseSymValue(pf, args)
		if err != nil {
			return nil, err
		}
		return &DefineExpr{Sym: sym, Val: val}, nil
	},
}

func parseSymValue(pf *sxeval.ParseEnvironment, args *sx.Pair) (*sx.Symbol, sxeval.Expr, error) {
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
	val, err := pf.Parse(pair.Car())
	return sym, val, err
}

// DefineExpr stores data for a define statement.
type DefineExpr struct {
	Sym *sx.Symbol
	Val sxeval.Expr
}

// Unparse the expression as an sx.Object
func (de *DefineExpr) Unparse() sx.Object {
	return sx.MakeList(sx.MakeSymbol(defvarName), de.Sym, de.Val.Unparse())
}

// Improve the expression into a possible simpler one.
func (de *DefineExpr) Improve(re *sxeval.ImproveEnvironment) sxeval.Expr {
	_ = re.Bind(de.Sym)
	de.Val = re.Rework(de.Val)
	return de
}

// Compute the expression in a frame and return the result.
func (de *DefineExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	val, err := env.Execute(de.Val)
	if err == nil {
		err = env.Bind(de.Sym, val)
	}
	return val, err
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
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
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
		val, err := pf.Parse(pair.Car())
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

// Unparse the expression as an sx.Object
func (se *SetXExpr) Unparse() sx.Object {
	return sx.MakeList(sx.MakeSymbol(setXName), se.Sym, se.Val.Unparse())
}

// Improve the expression into a possible simpler one.
func (se *SetXExpr) Improve(re *sxeval.ImproveEnvironment) sxeval.Expr {
	se.Val = re.Rework(se.Val)
	return se
}

// Compute the expression in a frame and return the result.
func (se *SetXExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	bind := env.FindBinding(se.Sym)
	if bind == nil {
		return nil, env.MakeNotBoundError(se.Sym)
	}
	val, err := env.Execute(se.Val)
	if err == nil {
		err = bind.Bind(se.Sym, val)
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
