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

// ReworkFrame guides the Expr.Rework operation.
type ReworkFrame struct {
	binding Binding // Current binding
}

// MakeChildFrame creates a subordinate rework frame with a new environment.
func (rf *ReworkFrame) MakeChildFrame(name string, baseSize int) *ReworkFrame {
	return &ReworkFrame{
		binding: MakeChildBinding(rf.binding, name, baseSize),
	}
}

// ResolveConst will resolve the symbol in an environment that is assumed not
// to be exchanged afterwards.
func (rf *ReworkFrame) ResolveConst(sym *sx.Symbol) (sx.Object, bool) {
	if bind := rf.binding; IsConstantBind(bind, sym) {
		return Resolve(bind, sym)
	}
	return nil, false
}

// Bind the undefined value to the symbol in the current environment.
func (rf *ReworkFrame) Bind(sym *sx.Symbol) error {
	return rf.binding.Bind(sym, sx.MakeUndefined())
}

// Call a function for constant folding.
//
// It is only called, if no full Frame is needed, only an environment.
func (rf *ReworkFrame) Call(fn Callable, args []sx.Object) (sx.Object, error) {
	frame := Frame{
		engine:   nil,
		executor: nil,
		binding:  rf.binding,
		caller:   nil,
	}
	return fn.Call(&frame, args)
}
