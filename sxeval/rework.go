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

import "zettelstore.de/sx.fossil"

// ReworkEnvironment guides the Expr.Rework operation.
type ReworkEnvironment struct {
	binding *Binding // Current binding
}

// MakeChildFrame creates a subordinate rework environment with a new binding.
func (re *ReworkEnvironment) MakeChildFrame(name string, baseSize int) *ReworkEnvironment {
	return &ReworkEnvironment{
		binding: MakeChildBinding(re.binding, name, baseSize),
	}
}

// ResolveConst will resolve the symbol, which is assumed not
// to be changed afterwards.
func (re *ReworkEnvironment) ResolveConst(sym *sx.Symbol) (sx.Object, bool) {
	if bind := re.binding; bind.isConstantBind(sym) {
		return bind.Resolve(sym)
	}
	return nil, false
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
	return fn.Call(env, args)
}
