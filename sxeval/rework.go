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

// ReworkEnvironment guides the Expr.Rework operation.
type ReworkEnvironment struct {
	base     *ReworkEnvironment
	binding  *Binding // Current binding
	height   int      // Height of current binding.
	observer ReworkObserver
}

// ReworkObserver monitors the inner workings of the rework process.
type ReworkObserver interface {
	// BeforeRework is called immediate before the given expression is reworked.
	BeforeRework(*ReworkEnvironment, Expr) Expr

	// AfterRework is called after the given expression was reworked to a
	// possibly simpler one.
	AfterRework(*ReworkEnvironment, Expr, Expr)
}

// MakeChildEnvironment creates a subordinate rework environment with a new binding.
func (re *ReworkEnvironment) MakeChildEnvironment(name string, baseSize int) *ReworkEnvironment {
	return &ReworkEnvironment{
		base:     re.base,
		binding:  re.binding.MakeChildBinding(name, baseSize),
		height:   re.height + 1,
		observer: re.observer,
	}
}

// Rework the given expression. Do not call `expr.Rework()` directly.
func (re *ReworkEnvironment) Rework(expr Expr) Expr {
	if observer := re.observer; observer != nil {
		expr2 := observer.BeforeRework(re, expr)
		result := expr2.Improve(re)
		observer.AfterRework(re, expr2, result)
		return result
	}
	return expr.Improve(re)
}

// Height returns the difference between the acual and the base height.
func (re *ReworkEnvironment) Height() int { return re.height - re.base.height }

// Binding returns the binding of this environment.
func (re *ReworkEnvironment) Binding() *Binding { return re.binding }

// Resolve the symbol into an object, and return the binding depth plus an
// indication about the const-ness of the value. If the symbol could not be
// resolved, depth has the value of `math.MinInt`. If the symbol was found
// in the base environment, depth is set to -1, to indicate a possible unbound
// situation.
func (re *ReworkEnvironment) Resolve(sym *sx.Symbol) (sx.Object, int, bool) {
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
func (re *ReworkEnvironment) Bind(sym *sx.Symbol) error {
	return re.binding.Bind(sym, sx.MakeUndefined())
}

// Call a function for constant folding.
//
// It is only called, if no full execution environment is needed, only a binding.
func (re *ReworkEnvironment) Call(fn Callable, args sx.Vector) (sx.Object, error) {
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
