//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
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
	root       Environment
	toplevel   Environment
	pars       Parser
	symResSym  *sx.Symbol
	symResCall *sx.Symbol
}

// MakeEngine creates a new engine.
func MakeEngine(sf sx.SymbolFactory, root Environment) *Engine {
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
func (sb simpleBuiltin) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	return sb(frame, args)
}
func (sb simpleBuiltin) IsEqual(other sx.Object) bool {
	if sb == nil {
		return sx.IsNil(other)
	}
	return false
}
func (sb simpleBuiltin) Repr() string   { return sx.Repr(sb) }
func (sb simpleBuiltin) String() string { return "simple-builtin" }

func resolveNotBound(frame *Frame, args []sx.Object) (sx.Object, error) {
	if sym, isSymbol := sx.GetSymbol(args[0]); isSymbol {
		return nil, frame.MakeNotBoundError(sym)
	}
	return nil, fmt.Errorf("argument 1 is not a symbol, but %T/%v", args[0], args[0])
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

// RootEnvironment returns the root environment of the engine.
func (eng *Engine) RootEnvironment() Environment { return eng.root }

// SetToplevelEnv sets the given environment as the top-level environment.
// It must be the root environment or a child of it.
func (eng *Engine) SetToplevelEnv(env Environment) error {
	root := RootEnv(env)
	if root != eng.root {
		return fmt.Errorf("root of %v is not root of engine %v: %v", env, eng.root, root)
	}
	eng.toplevel = env
	return nil
}

// GetToplevelEnv returns the current top-level environment.
func (eng *Engine) GetToplevelEnv() Environment { return eng.toplevel }

// SetParser updates the current s-expression parser of the engine.
func (eng *Engine) SetParser(p Parser) Parser {
	orig := eng.pars
	if p == nil {
		p = &myDefaultParser
	}
	eng.pars = p
	return orig
}

// Eval parses the given object and executes it in the environment.
func (eng *Engine) Eval(obj sx.Object, env Environment, exec Executor) (sx.Object, error) {
	expr, err := eng.Parse(obj, env)
	if err != nil {
		return nil, err
	}
	expr = eng.Rework(expr, env)
	return eng.Execute(expr, env, exec)
}

// Parse the given object in the given environment.
func (eng *Engine) Parse(obj sx.Object, env Environment) (Expr, error) {
	pf := ParseFrame{sf: eng.sf, env: env, parser: eng.pars}
	return pf.Parse(obj)
}

// Rework the given expression with the options stored in the engine.
func (eng *Engine) Rework(expr Expr, env Environment) Expr {
	rf := ReworkFrame{env: env}
	return expr.Rework(&rf)
}

// Execute the given expression in the given environment.
func (eng *Engine) Execute(expr Expr, env Environment, exec Executor) (sx.Object, error) {
	if exec != nil {
		exec.Reset()
	}
	frame := Frame{
		engine:   eng,
		executor: exec,
		env:      env,
		caller:   nil,
	}
	return frame.Execute(expr)
}

// BindSpecial binds a syntax parser to the its name in the engine's root environment.
func (eng *Engine) BindSpecial(syn *Special) error {
	return eng.BindConst(syn.Name, syn)
}

// BindBuiltin binds the given builtin with its given name in the engine's
// root environment.
func (eng *Engine) BindBuiltin(b *Builtin) error {
	return eng.BindConst(b.Name, b)
}

// BindConst a given object to a symbol of the given name as a constant in the
// engine's root environment.
func (eng *Engine) BindConst(name string, obj sx.Object) error {
	return eng.root.BindConst(eng.sf.MustMake(name), obj)
}

// Bind a given object to a symbol of the given name in the engine's root environment.
func (eng *Engine) Bind(name string, obj sx.Object) error {
	return eng.root.Bind(eng.sf.MustMake(name), obj)
}
