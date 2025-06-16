//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval

import (
	"t73f.de/r/sx"
)

// Improvable is an additional interface for `Expr` that can be possibly
// improved to simple ones.
type Improvable interface {
	// Improve the expression into a possible simpler one.
	Improve(*Improver) (Expr, error)
}

// Improver guides the improve operation.
type Improver struct {
	frame    *Frame // Current frame
	parent   *Improver
	env      *Environment
	observer ImproveObserver
	dynamic  bool // symbol resolving by full search?
}

// ImproveObserver monitors the inner workings of the improve process.
type ImproveObserver interface {
	// BeforeImprove is called immediate before the given expression is improved.
	BeforeImprove(*Improver, Expr) Expr

	// AfterImprove is called after the given expression was improved to a
	// possibly simpler one.
	AfterImprove(*Improver, Expr, Expr, error)
}

// MakeChildImprover creates a subordinate improver with a new frame.
func (imp *Improver) MakeChildImprover(name string, baseSize int, dynamic bool) *Improver {
	return &Improver{
		parent:   imp.parent,
		frame:    imp.frame.MakeChildFrame(name, baseSize),
		env:      imp.env,
		observer: imp.observer,
		dynamic:  dynamic,
	}
}

// Improve the given expression. Do not call `expr.Improve()` directly.
func (imp *Improver) Improve(expr Expr) (Expr, error) {
	if observer := imp.observer; observer != nil {
		expr = observer.BeforeImprove(imp, expr)
	}
	if iexpr, ok := expr.(Improvable); ok {
		result, err := iexpr.Improve(imp)
		if observer := imp.observer; observer != nil {
			observer.AfterImprove(imp, expr, result, err)
		}
		return result, err
	}
	if observer := imp.observer; observer != nil {
		observer.AfterImprove(imp, expr, expr, nil)
	}
	return expr, nil
}

// ImproveSlice improves the given slice by updating it.
func (imp *Improver) ImproveSlice(exprs []Expr) error {
	for i, expr := range exprs {
		iexpr, err := imp.Improve(expr)
		if err != nil {
			return err
		}
		exprs[i] = iexpr
	}
	return nil
}

// ImproveFoldCall improves a call if all args are constants and the
// callable is pure. If successful, the new expression is returned.
// Otherwise the expression is nil.
func (imp *Improver) ImproveFoldCall(proc Callable, args []Expr) (Expr, error) {
	vals := make(sx.Vector, len(args))
	for i, arg := range args {
		if objExpr, isConstObject := GetConstExpr(arg); isConstObject {
			vals[i] = objExpr.ConstObject()
		} else {
			return nil, nil
		}
	}
	if proc.IsPure(vals) {
		env := MakeEnvironment(imp.env.globals)
		if result, err := proc.ExecuteCall(env, vals, imp.frame); err == nil {
			return imp.Improve(ObjExpr{Obj: result})
		}
	}
	return nil, nil
}

// Frame returns the frame of this environment.
func (imp *Improver) Frame() *Frame { return imp.frame }

// Bind the undefined value to the symbol in the current frame.
func (imp *Improver) Bind(sym *sx.Symbol) {
	imp.frame.Bind(sym, sx.MakeUndefined())
}
