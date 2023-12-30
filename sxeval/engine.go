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

import (
	"fmt"

	"zettelstore.de/sx.fossil"
)

// Engine is the collection of all relevant data element to execute / evaluate an object.
type Engine struct {
	sf         sx.SymbolFactory
	root       Binding
	toplevel   Binding
	pars       Parser
	symResSym  *sx.Symbol
	symResCall *sx.Symbol
}

// MakeEngine creates a new engine.
func MakeEngine(sf sx.SymbolFactory, root Binding) *Engine {
	symResSym := sf.MustMake(resolveSymbolName)
	root.Bind(symResSym, simpleBuiltin(resolveNotBound))
	symResCall := sf.MustMake(resolveCallableName)
	root.Bind(symResCall, simpleBuiltin(resolveNotBound))
	return &Engine{
		sf:         sf,
		root:       root,
		toplevel:   root,
		pars:       &myDefaultParser,
		symResSym:  symResSym,
		symResCall: symResCall,
	}
}

const resolveSymbolName = "*RESOLVE-SYMBOL*"
const resolveCallableName = "*RESOLVE-CALLABLE*"

type simpleBuiltin func(*Frame, []sx.Object) (sx.Object, error)

func (sb simpleBuiltin) IsNil() bool  { return sb == nil }
func (sb simpleBuiltin) IsAtom() bool { return sb == nil }
func (sb simpleBuiltin) IsEqual(other sx.Object) bool {
	if sb == nil {
		return sx.IsNil(other)
	}
	return false
}
func (sb simpleBuiltin) Repr() string            { return "#<simple-builtin>" }
func (sb simpleBuiltin) String() string          { return "simple-builtin" }
func (sb simpleBuiltin) IsPure([]sx.Object) bool { return false }
func (sb simpleBuiltin) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	return sb(frame, args)
}

func resolveNotBound(frame *Frame, args []sx.Object) (sx.Object, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("at least one argument expected, but none given")
	}
	sym, isSymbol := sx.GetSymbol(args[0])
	if !isSymbol {
		return nil, fmt.Errorf("argument 1 is not a symbol, but %T/%v", args[0], args[0])
	}
	bind := frame.binding
	if len(args) > 1 {
		argBind, isBind := GetBinding(args[1])
		if !isBind {
			return nil, fmt.Errorf("argument 1 is not a binding, but %T/%v", args[1], args[1])
		}
		bind = argBind
	}
	return nil, NotBoundError{Binding: bind, Sym: sym}
}

// Copy creates a shallow copy of the given engine.
func (eng *Engine) Copy() *Engine {
	if eng == nil {
		return nil
	}
	result := *eng
	return &result
}

// SymbolFactory returns the symbol factory of the engine.
func (eng *Engine) SymbolFactory() sx.SymbolFactory { return eng.sf }

// RootBinding returns the root binding of the engine.
func (eng *Engine) RootBinding() Binding { return eng.root }

// SetToplevelBinding sets the given binding as the top-level binding.
// It must be the root binding or a child of it.
func (eng *Engine) SetToplevelBinding(bind Binding) error {
	root := RootBinding(bind)
	if root != eng.root {
		return fmt.Errorf("root of %v is not root of engine %v: %v", bind, eng.root, root)
	}
	eng.toplevel = bind
	return nil
}

// GetToplevelBinding returns the current top-level binding.
func (eng *Engine) GetToplevelBinding() Binding { return eng.toplevel }

// SetParser updates the current s-expression parser of the engine.
func (eng *Engine) SetParser(p Parser) Parser {
	orig := eng.pars
	if p == nil {
		p = &myDefaultParser
	}
	eng.pars = p
	return orig
}

// Eval parses the given object and executes it in the binding.
func (eng *Engine) Eval(obj sx.Object, bind Binding, exec Executor) (sx.Object, error) {
	expr, err := eng.Parse(obj, bind)
	if err != nil {
		return nil, err
	}
	expr = eng.Rework(expr, bind)
	return eng.Execute(expr, bind, exec)
}

// Parse the given object in the given binding.
func (eng *Engine) Parse(obj sx.Object, bind Binding) (Expr, error) {
	pf := ParseFrame{sf: eng.sf, binding: bind, parser: eng.pars}
	return pf.Parse(obj)
}

// Rework the given expression with the options stored in the engine.
func (eng *Engine) Rework(expr Expr, bind Binding) Expr {
	rf := ReworkFrame{binding: bind}
	return expr.Rework(&rf)
}

// Execute the given expression in the given binding.
func (eng *Engine) Execute(expr Expr, bind Binding, exec Executor) (sx.Object, error) {
	if exec != nil {
		exec.Reset()
	}
	frame := Frame{
		engine:   eng,
		executor: exec,
		binding:  bind,
		caller:   nil,
	}
	return frame.Execute(expr)
}

// BindSpecial binds a syntax parser to the its name in the engine's root binding.
func (eng *Engine) BindSpecial(syn *Special) error {
	return eng.BindConst(syn.Name, syn)
}

// BindBuiltin binds the given builtin with its given name in the engine's
// root binding.
func (eng *Engine) BindBuiltin(b *Builtin) error {
	return eng.BindConst(b.Name, b)
}

// BindConst a given object to a symbol of the given name as a constant in the
// engine's root binding.
func (eng *Engine) BindConst(name string, obj sx.Object) error {
	return eng.root.BindConst(eng.sf.MustMake(name), obj)
}

// Bind a given object to a symbol of the given name in the engine's root binding.
func (eng *Engine) Bind(name string, obj sx.Object) error {
	return eng.root.Bind(eng.sf.MustMake(name), obj)
}
