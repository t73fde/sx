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
	"math"

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
	base     *Improver
	binding  *Binding // Current binding
	height   int      // Height of current binding
	observer ImproveObserver
}

// ImproveObserver monitors the inner workings of the improve process.
type ImproveObserver interface {
	// BeforeImprove is called immediate before the given expression is improved.
	BeforeImprove(*Improver, Expr) Expr

	// AfterImprove is called after the given expression was improved to a
	// possibly simpler one.
	AfterImprove(*Improver, Expr, Expr, error)
}

// MakeChildImprover creates a subordinate improver with a new binding.
func (imp *Improver) MakeChildImprover(name string, baseSize int) *Improver {
	return &Improver{
		base:     imp.base,
		binding:  imp.binding.MakeChildBinding(name, baseSize),
		height:   imp.height + 1,
		observer: imp.observer,
	}
}

// Improve the given expression. Do not call `expr.Improve()` directly.
func (imp *Improver) Improve(expr Expr) (Expr, error) {
	if observer := imp.observer; observer != nil {
		expr2 := observer.BeforeImprove(imp, expr)
		if iexpr2, ok := expr2.(Improvable); ok {
			result, err := iexpr2.Improve(imp)
			observer.AfterImprove(imp, expr2, result, err)
			return result, err
		}
		observer.AfterImprove(imp, expr2, expr2, nil)
		return expr2, nil
	}
	if iexpr, ok := expr.(Improvable); ok {
		return iexpr.Improve(imp)
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
		if objExpr, isConstObject := arg.(ConstObjectExpr); isConstObject {
			vals[i] = objExpr.ConstObject()
		} else {
			return nil, nil
		}
	}
	if proc.IsPure(vals) {
		if result, err := imp.Call(proc, vals); err == nil {
			return imp.Improve(ObjExpr{Obj: result})
		}
	}
	return nil, nil
}

// Height returns the difference between the actual and the base height.
func (imp *Improver) Height() int { return imp.height - imp.base.height }

// Binding returns the binding of this environment.
func (imp *Improver) Binding() *Binding { return imp.binding }

// Resolve the symbol into an object, and return the binding depth plus an
// indication about the const-ness of the value. If the symbol could not be
// resolved, depth has the value of `math.MinInt`. If the symbol was found
// in the base environment, depth is set to -1, to indicate a possible unbound
// situation.
func (imp *Improver) Resolve(sym *sx.Symbol) (sx.Object, int, bool) {
	obj, b, depth := imp.binding.resolveFull(sym)
	if b == nil {
		return nil, math.MinInt, false
	}
	if depth >= imp.Height() {
		return obj, -1, b.IsFrozen()
	}
	return obj, depth, b.IsFrozen()
}

// Bind the undefined value to the symbol in the current environment.
func (imp *Improver) Bind(sym *sx.Symbol) error {
	return imp.binding.Bind(sym, sx.MakeUndefined())
}

// Call a function for constant folding.
//
// It is only called, if no full execution environment is needed, only a binding.
func (imp *Improver) Call(fn Callable, args sx.Vector) (sx.Object, error) {
	env := MakeComputeEnvironment(imp.binding)
	return fn.Call(env, args)
}
