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
	"io"
	"strconv"

	"zettelstore.de/sx.fossil"
)

// Environment is a runtime object of the current computing environment.
type Environment struct {
	parent   *Environment // the lexical parent environment
	engine   *Engine
	executor Executor // most of the time: engine.exec, but could be updated for interactive debugging
	binding  *Binding
	caller   *Environment // the dynamic call stack
	name     string
}

// MakeExecutionEnvironment creates an environment for later execution of an expression.
func MakeExecutionEnvironment(eng *Engine, exec Executor, bind *Binding) Environment {
	if exec != nil {
		exec.Reset()
	}
	parent := createLexicalEnvs(eng, exec, bind.parent)
	return Environment{
		parent:   parent,
		engine:   eng,
		executor: exec,
		binding:  bind,
		caller:   nil,
		name:     bind.name,
	}
}

func createLexicalEnvs(eng *Engine, exec Executor, bind *Binding) *Environment {
	if bind == nil {
		return nil
	}
	parent := createLexicalEnvs(eng, exec, bind.parent)
	return &Environment{
		parent:   parent,
		engine:   eng,
		executor: exec,
		binding:  bind,
		caller:   nil,
		name:     bind.name,
	}
}

func (env *Environment) IsNil() bool  { return env == nil }
func (env *Environment) IsAtom() bool { return env == nil }
func (env *Environment) IsEqual(other sx.Object) bool {
	if env == other {
		return true
	}
	if env.IsNil() {
		return sx.IsNil(other)
	}
	if oenv, ok := other.(*Environment); ok {
		return env.engine == oenv.engine &&
			env.executor == oenv.executor &&
			env.binding.IsEqual(oenv.binding) &&
			env.caller == oenv.caller &&
			env.name == oenv.name
	}
	return false
}
func (env *Environment) Repr() string { return sx.Repr(env) }
func (env *Environment) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<environment:", env.name, "/", strconv.Itoa(len(env.binding.vars)), ">")
}

// String returns the local name of this binding.
func (env *Environment) String() string { return env.name }

// Parent returns the lexical parent environment.
//
// Please note that env.binding.parent may be different to env.parent.binding.
// This is because of the use of a parse frame binding in NewLexicalEnvironment
//
// Because of this, the builtin (parent-environment) is broken!
//
// We should probably merge Environment and ParseFrame somehow.
func (env *Environment) Parent() *Environment {
	if env == nil {
		return nil
	}
	return env.parent
}

// GetEnvironment returns the object as an environment, if possible.
func GetEnvironment(obj sx.Object) (*Environment, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	env, ok := obj.(*Environment)
	return env, ok
}

func (env *Environment) NewDynamicEnvironment() *Environment {
	return &Environment{
		parent:   env.parent,
		engine:   env.engine,
		executor: env.executor,
		binding:  env.binding,
		caller:   env,
		name:     env.name,
	}
}

func (env *Environment) MakeParseFrame() *ParseFrame {
	return &ParseFrame{
		sf:      env.engine.SymbolFactory(),
		binding: env.binding,
		parser:  env.engine.pars,
	}
}

func (env *Environment) MakeReworkFrame() *ReworkFrame {
	return &ReworkFrame{
		binding: env.binding,
	}
}

func (env *Environment) NewLexicalEnvironment(pf *ParseFrame, name string, numBindings int) *Environment {
	return &Environment{
		parent:   env,
		engine:   env.engine,
		executor: env.executor,
		binding:  MakeChildBinding(pf.binding, name, numBindings),
		caller:   env,
		name:     name,
	}
}

func (env *Environment) Execute(expr Expr) (sx.Object, error) {
	if exec := env.executor; exec != nil {
		for {
			res, err := exec.Execute(env, expr)
			if err == nil {
				return res, nil
			}
			if again, ok := err.(executeAgain); ok {
				env.binding = again.binding
				expr = again.expr
				continue
			}
			return res, err
		}
	}

	for {
		res, err := expr.Compute(env)
		if err == nil {
			return res, nil
		}
		if again, ok := err.(executeAgain); ok {
			env.binding = again.binding
			expr = again.expr
			continue
		}
		return res, err
	}
}
func (env *Environment) ExecuteTCO(expr Expr) (sx.Object, error) {
	// Uncomment this line to test for non-TCO
	// subEnv := env.NewDynamicEnvironment()
	// return subEnv.Execute(expr)

	// Just return relevant data for real TCO
	return nil, executeAgain{binding: env.binding, expr: expr}
}

func (env *Environment) Call(fn Callable, args []sx.Object) (sx.Object, error) {
	dynamicEnv := Environment{
		engine:   env.engine,
		executor: env.executor,
		binding:  env.binding,
		caller:   env,
		name:     env.name,
	}
	res, err := fn.Call(&dynamicEnv, args)
	if err == nil {
		return res, nil
	}
	if again, ok := err.(executeAgain); ok {
		dynamicEnv.binding = again.binding
		return dynamicEnv.Execute(again.expr)
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

func (env *Environment) CallResolveSymbol(sym *sx.Symbol) (sx.Object, error) {
	return env.callResolve(sym, env.engine.symResSym)
}
func (env *Environment) CallResolveCallable(sym *sx.Symbol) (sx.Object, error) {
	return env.callResolve(sym, env.engine.symResCall)
}
func (env *Environment) callResolve(sym *sx.Symbol, defSym *sx.Symbol) (sx.Object, error) {
	if obj, found := env.Resolve(defSym); found {
		if fn, isCallable := obj.(Callable); isCallable {
			return env.Call(fn, []sx.Object{sym, env})
		}
	}
	return nil, env.MakeNotBoundError(sym)
}

func (env *Environment) Bind(sym *sx.Symbol, obj sx.Object) error {
	return env.binding.Bind(sym, obj)
}
func (env *Environment) BindConst(sym *sx.Symbol, obj sx.Object) error {
	return env.binding.BindConst(sym, obj)
}
func (env *Environment) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	return env.binding.Lookup(sym)
}
func (env *Environment) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return resolve(env.binding, sym)
}
func (env *Environment) FindBinding(sym *sx.Symbol) *Binding {
	for curr := env.binding; curr != nil; curr = curr.parent {
		if _, found := curr.Lookup(sym); found {
			return curr
		}
	}
	return nil
}
func (env *Environment) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Env: env, Sym: sym}
}

// NotBoundError signals that a symbol was not found in a binding.
type NotBoundError struct {
	Env *Environment
	Sym *sx.Symbol
}

func (e NotBoundError) Error() string {
	return fmt.Sprintf("symbol %q not bound in %q", e.Sym.Name(), e.Env.String())
}
func (env *Environment) Bindings() *sx.Pair { return env.binding.Bindings() }
