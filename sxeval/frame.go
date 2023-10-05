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
	env      Environment
	caller   *Frame
	executor Executor
}

func (frame *Frame) MakeCalleeFrame() *Frame {
	return &Frame{
		engine:   frame.engine,
		env:      frame.env,
		caller:   frame,
		executor: frame.executor,
	}
}
func (frame *Frame) MakeParseFrame() *ParseFrame {
	return &ParseFrame{
		engine: frame.engine,
		env:    frame.env,
		parser: frame.engine.pars,
	}
}

func (frame *Frame) MakeLetFrame(name string, numBindings int) *Frame {
	return &Frame{
		engine:   frame.engine,
		env:      MakeFixedEnvironment(frame.env, name, numBindings),
		caller:   frame,
		executor: frame.executor,
	}
}
func (frame *Frame) MakeLambdaFrame(pf *ParseFrame, name string, numBindings int) *Frame {
	return &Frame{
		engine:   frame.engine,
		env:      MakeFixedEnvironment(pf.env, name, numBindings),
		caller:   frame,
		executor: frame.executor,
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
		engine: frame.engine,
		env:    frame.env,
		caller: frame,
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

func (frame *Frame) Bind(sym *sx.Symbol, obj sx.Object) error { return frame.env.Bind(sym, obj) }
func (frame *Frame) Freeze()                                  { frame.env.Freeze() }
func (frame *Frame) Lookup(sym *sx.Symbol) (sx.Object, bool)  { return frame.env.Lookup(sym) }
func (frame *Frame) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return Resolve(frame.env, sym)
}
func (frame *Frame) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Env: frame.env, Sym: sym}
}

func (frame *Frame) Environment() Environment { return frame.env }
