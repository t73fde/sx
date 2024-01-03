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

// Package sxeval allows to evaluate s-expressions. Evaluation is splitted into
// parsing that s-expression and executing the result of the parsed expression.
// This is done to reduce syntax checks.
package sxeval

import "fmt"

// Engine is the collection of all relevant data element to execute / evaluate an object.
type Engine struct {
	root     *Binding
	toplevel *Binding
}

// MakeEngine creates a new engine.
func MakeEngine(root *Binding) *Engine {
	return &Engine{
		root:     root,
		toplevel: root,
	}
}

// Copy creates a shallow copy of the given engine.
func (eng *Engine) Copy() *Engine {
	if eng == nil {
		return nil
	}
	result := *eng
	return &result
}

// RootBinding returns the root binding of the engine.
func (eng *Engine) RootBinding() *Binding { return eng.root }

// SetToplevelBinding sets the given binding as the top-level binding.
// It must be the root binding or a child of it.
func (eng *Engine) SetToplevelBinding(bind *Binding) error {
	root := bind.rootBinding()
	if root != eng.root {
		return fmt.Errorf("root of %v is not root of engine %v: %v", bind, eng.root, root)
	}
	eng.toplevel = bind
	return nil
}

// GetToplevelBinding returns the current top-level binding.
func (eng *Engine) GetToplevelBinding() *Binding { return eng.toplevel }
