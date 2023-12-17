// -----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
	env      Environment
	caller   *Frame
}

func (frame *Frame) MakeCalleeFrame() *Frame {
	return &Frame{
		engine:   frame.engine,
		executor: frame.executor,
		env:      frame.env,
		caller:   frame,
	}
}

func (frame *Frame) MakeParseFrame() *ParseFrame {
	return &ParseFrame{
		sf:     frame.engine.SymbolFactory(),
		env:    frame.env,
		parser: frame.engine.pars,
	}
}

func (frame *Frame) MakeReworkFrame() *ReworkFrame {
	return &ReworkFrame{
		env: frame.env,
	}
}

func (frame *Frame) MakeLambdaFrame(pf *ParseFrame, name string, numBindings int) *Frame {
	return &Frame{
		engine:   frame.engine,
		executor: frame.executor,
		env:      MakeChildEnvironment(pf.env, name, numBindings),
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
				frame.env = again.env
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
			frame.env = again.env
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
	return nil, executeAgain{env: frame.env, expr: expr}
}

func (frame *Frame) Call(fn Callable, args []sx.Object) (sx.Object, error) {
	callFrame := Frame{
		engine:   frame.engine,
		executor: frame.executor,
		env:      frame.env,
		caller:   frame,
	}
	res, err := fn.Call(&callFrame, args)
	if err == nil {
		return res, nil
	}
	if again, ok := err.(executeAgain); ok {
		callFrame.env = again.env
		return callFrame.Execute(again.expr)
	}
	return nil, err
}

// executeAgain is a non-error error signalling that the given expression should be
// executed again in the given environment.
type executeAgain struct {
	env  Environment
	expr Expr
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
			return frame.Call(fn, []sx.Object{sym, frame.env})
		}
	}
	return nil, frame.MakeNotBoundError(sym)
}

func (frame *Frame) Bind(sym *sx.Symbol, obj sx.Object) error { return frame.env.Bind(sym, obj) }
func (frame *Frame) BindConst(sym *sx.Symbol, obj sx.Object) error {
	return frame.env.BindConst(sym, obj)
}
func (frame *Frame) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return Resolve(frame.env, sym)
}
func (frame *Frame) FindBindingEnv(sym *sx.Symbol) Environment {
	env := frame.env
	for !sx.IsNil(env) {
		if _, found := env.Lookup(sym); found {
			return env
		}
		env = env.Parent()
	}
	return env
}
func (frame *Frame) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Env: frame.env, Sym: sym}
}

// NotBoundError signals that a symbol was not found in an environment.
type NotBoundError struct {
	Env Environment
	Sym *sx.Symbol
}

func (e NotBoundError) Error() string {
	return fmt.Sprintf("symbol %q not bound in environment %q", e.Sym.Name(), e.Env.String())
}
func (frame *Frame) Environment() Environment { return frame.env }
