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
	"zettelstore.de/sx.fossil"
)

// ParseFrame is a parsing environment.
type ParseFrame struct {
	engine *Engine
	env    Environment
}

func (frame *ParseFrame) IsEql(other *ParseFrame) bool {
	if frame == other {
		return true
	}
	if frame == nil || other == nil {
		return false
	}
	if frame.engine != other.engine {
		return false
	}
	return frame.env.IsEql(other.env)
}

func (pf *ParseFrame) Parse(obj sx.Object) (Expr, error) {
	return pf.engine.Parse(pf.env, obj)
}

func (pf *ParseFrame) Call(fn Callable, args []sx.Object) (sx.Object, error) {
	return pf.engine.Call(pf.env, fn, args)
}

func (pf *ParseFrame) MakeChildFrame(name string, baseSize int) *ParseFrame {
	return &ParseFrame{
		engine: pf.engine,
		env:    MakeChildEnvironment(pf.env, name, baseSize),
	}
}

func (pf *ParseFrame) Bind(sym *sx.Symbol, obj sx.Object) error {
	env, err := pf.env.Bind(sym, obj)
	pf.env = env
	return err
}

func (pf *ParseFrame) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return Resolve(pf.env, sym)
}
func (pf *ParseFrame) Environment() Environment { return pf.env }

// Frame is a runtime object of the current computing environment.
type Frame struct {
	engine *Engine
	env    Environment
}

func (frame *Frame) MakeParseFrame() *ParseFrame {
	return &ParseFrame{
		engine: frame.engine,
		env:    frame.env,
	}
}
func (frame *Frame) MakeChildEnvFrame(name string, baseSize int) *Frame {
	return &Frame{
		engine: frame.engine,
		env:    MakeChildEnvironment(frame.env, name, baseSize),
	}
}
func (frame *Frame) UpdateChildFrame(pf *ParseFrame, name string, baseSize int) *Frame {
	return &Frame{
		engine: frame.engine,
		env:    MakeChildEnvironment(pf.env, name, baseSize),
	}
}

func (frame *Frame) Execute(expr Expr) (sx.Object, error) {
	return frame.engine.Execute(frame.env, expr)
}
func (frame *Frame) ExecuteTCO(expr Expr) (sx.Object, error) {
	return frame.engine.ExecuteTCO(frame.env, expr)
}
func (frame *Frame) Call(fn Callable, args []sx.Object) (sx.Object, error) {
	return frame.engine.Call(frame.env, fn, args)
}

func (frame *Frame) Bind(sym *sx.Symbol, obj sx.Object) error {
	env, err := frame.env.Bind(sym, obj)
	frame.env = env
	return err
}
func (frame *Frame) Lookup(sym *sx.Symbol) (sx.Object, bool) { return frame.env.Lookup(sym) }
func (frame *Frame) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return Resolve(frame.env, sym)
}
func (frame *Frame) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Env: frame.env, Sym: sym}
}

func (frame *Frame) Environment() Environment { return frame.env }
