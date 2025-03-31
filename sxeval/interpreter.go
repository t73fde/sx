//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval

import (
	"t73f.de/r/sx"
)

// Interpreter is the runtime data used by the compiled code.
type Interpreter struct {
	stack []sx.Object
	sp    int // stack pointer
	env   *Environment
}

// NewInterpreter creates a new interpreter with a given minimum stack size.
func NewInterpreter(env *Environment, stacksize int) Interpreter {
	return Interpreter{
		stack: make([]sx.Object, 0, stacksize),
		sp:    0,
		env:   env,
	}
}

// Push a value to the stack
func (interp *Interpreter) Push(val sx.Object) {
	interp.stack = append(interp.stack, val)
	interp.sp++
}

// Pop a value from the stack
func (interp *Interpreter) Pop() sx.Object {
	interp.sp--
	return interp.stack[interp.sp]
}

// Top returns the value on top of the stack
func (interp *Interpreter) Top() sx.Object { return interp.stack[interp.sp-1] }

// Set the value on top of the stack
func (interp *Interpreter) Set(val sx.Object) { interp.stack[interp.sp-1] = val }

// Kill some elements on the stack
func (interp *Interpreter) Kill(num int) { interp.sp -= num }

// IFunc is a compiled command to be executed by an interpreter.
type IFunc func(*Interpreter) error
