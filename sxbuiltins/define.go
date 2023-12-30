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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// DefVarS parses a (defvar name value) form.
var DefVarS = sxeval.Special{
	Name: "defvar",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
		sym, val, err := parseSymValue(pf, args)
		if err != nil {
			return nil, err
		}
		return &DefineExpr{Sym: sym, Val: val, Const: false}, nil
	},
}

// DefConstS parses a (defconst name value) form.
var DefConstS = sxeval.Special{
	Name: "defconst",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
		sym, val, err := parseSymValue(pf, args)
		if err != nil {
			return nil, err
		}
		return &DefineExpr{Sym: sym, Val: val, Const: true}, nil
	},
}

func parseSymValue(pf *sxeval.ParseFrame, args *sx.Pair) (*sx.Symbol, sxeval.Expr, error) {
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
	Sym   *sx.Symbol
	Val   sxeval.Expr
	Const bool
}

func (de *DefineExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	de.Val = de.Val.Rework(rf)
	return de
}
func (de *DefineExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	subEnv := env.NewDynamicEnvironment()
	val, err := subEnv.Execute(de.Val)
	if err == nil {
		if de.Const {
			err = env.BindConst(de.Sym, val)
		} else {
			err = env.Bind(de.Sym, val)
		}
	}
	return val, err
}
func (de *DefineExpr) IsEqual(other sxeval.Expr) bool {
	if de == other {
		return true
	}
	if otherD, ok := other.(*DefineExpr); ok && otherD != nil {
		return de.Sym.IsEqual(otherD.Sym) &&
			de.Val.IsEqual(otherD.Val) &&
			de.Const == otherD.Const
	}
	return false
}

func (de *DefineExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{DEFINE ")
	if err != nil {
		return length, err
	}
	var l int
	if de.Const {
		l, err = io.WriteString(w, "CONST ")
		length += l
		if err != nil {
			return length, err
		}
	}
	l, err = sx.Print(w, de.Sym)
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

// SetXS parses a (set! name value) form.
var SetXS = sxeval.Special{
	Name: "set!",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
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

func (se *SetXExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	se.Val = se.Val.Rework(rf)
	return se
}
func (se *SetXExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	bind := env.FindBinding(se.Sym)
	if sx.IsNil(bind) {
		return nil, env.MakeNotBoundError(se.Sym)
	}
	subEnv := env.NewDynamicEnvironment()
	val, err := subEnv.Execute(se.Val)
	if err == nil {
		err = bind.Bind(se.Sym, val)
	}
	return val, err
}
func (se *SetXExpr) IsEqual(other sxeval.Expr) bool {
	if se == other {
		return true
	}
	if otherM, ok := other.(*SetXExpr); ok && otherM != nil {
		return se.Sym.IsEqual(otherM.Sym) && se.Val.IsEqual(otherM.Val)
	}
	return false
}

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
