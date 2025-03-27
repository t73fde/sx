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
	Improve(*Improver) Expr
}

// Improver guides the improve operation.
type Improver struct {
	base     *Improver
	binding  *Binding // Current binding
	height   int      // Height of current binding.
	observer ImproveObserver
}

// ImproveObserver monitors the inner workings of the improve process.
type ImproveObserver interface {
	// BeforeImprove is called immediate before the given expression is improved.
	BeforeImprove(*Improver, Expr) Expr

	// AfterImprove is called after the given expression was improved to a
	// possibly simpler one.
	AfterImprove(*Improver, Expr, Expr)
}

// MakeChildImprover creates a subordinate improver with a new binding.
func (re *Improver) MakeChildImprover(name string, baseSize int) *Improver {
	return &Improver{
		base:     re.base,
		binding:  re.binding.MakeChildBinding(name, baseSize),
		height:   re.height + 1,
		observer: re.observer,
	}
}

// Improve the given expression. Do not call `expr.Improve()` directly.
func (re *Improver) Improve(expr Expr) Expr {
	if observer := re.observer; observer != nil {
		expr2 := observer.BeforeImprove(re, expr)
		if iexpr2, ok := expr2.(Improvable); ok {
			result := iexpr2.Improve(re)
			observer.AfterImprove(re, expr2, result)
			return result
		}
		observer.AfterImprove(re, expr2, expr2)
		return expr2
	}
	if iexpr, ok := expr.(Improvable); ok {
		return iexpr.Improve(re)
	}
	return expr
}

// Height returns the difference between the actual and the base height.
func (re *Improver) Height() int { return re.height - re.base.height }

// Binding returns the binding of this environment.
func (re *Improver) Binding() *Binding { return re.binding }

// Resolve the symbol into an object, and return the binding depth plus an
// indication about the const-ness of the value. If the symbol could not be
// resolved, depth has the value of `math.MinInt`. If the symbol was found
// in the base environment, depth is set to -1, to indicate a possible unbound
// situation.
func (re *Improver) Resolve(sym *sx.Symbol) (sx.Object, int, bool) {
	obj, b, depth := re.binding.resolveFull(sym)
	if b == nil {
		return nil, math.MinInt, false
	}
	if depth >= re.Height() {
		return obj, -1, b.IsFrozen()
	}
	return obj, depth, b.IsFrozen()
}

// Bind the undefined value to the symbol in the current environment.
func (re *Improver) Bind(sym *sx.Symbol) error {
	return re.binding.Bind(sym, sx.MakeUndefined())
}

// Call a function for constant folding.
//
// It is only called, if no full execution environment is needed, only a binding.
func (re *Improver) Call(fn Callable, args sx.Vector) (sx.Object, error) {
	env := MakeExecutionEnvironment(re.binding)
	switch len(args) {
	case 0:
		return fn.Call0(env)
	case 1:
		return fn.Call1(env, args[0])
	case 2:
		return fn.Call2(env, args[0], args[1])
	default:
		return fn.Call(env, args)
	}
}
