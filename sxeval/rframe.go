//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval

import "zettelstore.de/sx.fossil"

// ReworkFrame guides the Expr.Rework operation.
type ReworkFrame struct {
	env      Environment // Current environment
	engine   *Engine
	executor Executor
}

// MakeChildFrame creates a subordinate rework frame with a new environment.
func (rf *ReworkFrame) MakeChildFrame(name string, baseSize int) *ReworkFrame {
	return &ReworkFrame{
		env:      MakeChildEnvironment(rf.env, name, baseSize),
		engine:   rf.engine,
		executor: rf.executor,
	}
}

// MakeFrame constructs a frame for calling external functions.
func (rf *ReworkFrame) MakeFrame() *Frame {
	return &Frame{
		engine:   rf.engine,
		executor: rf.executor,
		env:      rf.env,
		caller:   nil,
	}
}

// ResolveConst will resolve the symbol in an environment that is assumed not
// to b echanged afterwards.
func (rf *ReworkFrame) ResolveConst(sym *sx.Symbol) (sx.Object, bool) {
	if env := rf.env; IsConstantBinding(env, sym) {
		return Resolve(env, sym)
	}
	return nil, false
}

// Bind the undefined value to the symbol in the current environment.
func (rf *ReworkFrame) Bind(sym *sx.Symbol) error {
	return rf.env.Bind(sym, sx.MakeUndefined())
}
