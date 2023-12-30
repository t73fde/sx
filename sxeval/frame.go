// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------

package sxeval

import (
	"fmt"

	"zettelstore.de/sx.fossil"
)

// Frame is a runtime object of the current computing environment.
type Frame struct {
	engine   *Engine
	executor Executor // most of the time: engine.exec, but could be updated for interactive debugging
	binding  *Binding
	caller   *Frame
}

func (frame *Frame) MakeCalleeFrame() *Frame {
	return &Frame{
		engine:   frame.engine,
		executor: frame.executor,
		binding:  frame.binding,
		caller:   frame,
	}
}

func (frame *Frame) MakeParseFrame() *ParseFrame {
	return &ParseFrame{
		sf:      frame.engine.SymbolFactory(),
		binding: frame.binding,
		parser:  frame.engine.pars,
	}
}

func (frame *Frame) MakeReworkFrame() *ReworkFrame {
	return &ReworkFrame{
		binding: frame.binding,
	}
}

func (frame *Frame) MakeLambdaFrame(pf *ParseFrame, name string, numBindings int) *Frame {
	return &Frame{
		engine:   frame.engine,
		executor: frame.executor,
		binding:  MakeChildBinding(pf.binding, name, numBindings),
		caller:   frame,
	}
}

func (frame *Frame) Execute(expr Expr) (sx.Object, error) {
	if exec := frame.executor; exec != nil {
		for {
			res, err := exec.Execute(frame, expr)
			if err == nil {
				return res, nil
			}
			if again, ok := err.(executeAgain); ok {
				frame.binding = again.binding
				expr = again.expr
				continue
			}
			return res, err
		}
	}

	for {
		res, err := expr.Compute(frame)
		if err == nil {
			return res, nil
		}
		if again, ok := err.(executeAgain); ok {
			frame.binding = again.binding
			expr = again.expr
			continue
		}
		return res, err
	}
}
func (frame *Frame) ExecuteTCO(expr Expr) (sx.Object, error) {
	// Uncomment this line to test for non-TCO
	// subFrame := frame.MakeCalleeFrame()
	// return subFrame.Execute(expr)

	// Just return relevant data for real TCO
	return nil, executeAgain{binding: frame.binding, expr: expr}
}

func (frame *Frame) Call(fn Callable, args []sx.Object) (sx.Object, error) {
	callFrame := Frame{
		engine:   frame.engine,
		executor: frame.executor,
		binding:  frame.binding,
		caller:   frame,
	}
	res, err := fn.Call(&callFrame, args)
	if err == nil {
		return res, nil
	}
	if again, ok := err.(executeAgain); ok {
		callFrame.binding = again.binding
		return callFrame.Execute(again.expr)
	}
	return nil, err
}

// executeAgain is a non-error error signalling that the given expression should be
// executed again in the given binding.
type executeAgain struct {
	binding *Binding
	expr    Expr
}

func (e executeAgain) Error() string { return fmt.Sprintf("Again: %v", e.expr) }

func (frame *Frame) CallResolveSymbol(sym *sx.Symbol) (sx.Object, error) {
	return frame.callResolve(sym, frame.engine.symResSym)
}
func (frame *Frame) CallResolveCallable(sym *sx.Symbol) (sx.Object, error) {
	return frame.callResolve(sym, frame.engine.symResCall)
}
func (frame *Frame) callResolve(sym *sx.Symbol, defSym *sx.Symbol) (sx.Object, error) {
	if obj, found := frame.Resolve(defSym); found {
		if fn, isCallable := obj.(Callable); isCallable {
			return frame.Call(fn, []sx.Object{sym, frame.binding})
		}
	}
	return nil, frame.MakeNotBoundError(sym)
}

func (frame *Frame) Bind(sym *sx.Symbol, obj sx.Object) error { return frame.binding.Bind(sym, obj) }
func (frame *Frame) BindConst(sym *sx.Symbol, obj sx.Object) error {
	return frame.binding.BindConst(sym, obj)
}
func (frame *Frame) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return Resolve(frame.binding, sym)
}
func (frame *Frame) FindBinding(sym *sx.Symbol) *Binding {
	bind := frame.binding
	for !sx.IsNil(bind) {
		if _, found := bind.Lookup(sym); found {
			return bind
		}
		bind = bind.Parent()
	}
	return bind
}
func (frame *Frame) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Binding: frame.binding, Sym: sym}
}

// NotBoundError signals that a symbol was not found in a binding.
type NotBoundError struct {
	Binding *Binding
	Sym     *sx.Symbol
}

func (e NotBoundError) Error() string {
	return fmt.Sprintf("symbol %q not bound in %q", e.Sym.Name(), e.Binding.String())
}
func (frame *Frame) Binding() *Binding { return frame.binding }
