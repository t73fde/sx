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

// Environment is a runtime object of the current computing environment.
type Environment struct {
	binding  *Binding
	executor Executor
	observer ParseObserver
	caller   *Environment // the dynamic call stack
}

// MakeExecutionEnvironment creates an environment for later execution of an expression.
func MakeExecutionEnvironment(bind *Binding, options ...Option) *Environment {
	env := &Environment{
		binding:  bind,
		executor: nil,
		observer: nil,
		caller:   nil,
	}
	for _, opt := range options {
		opt(env)
	}
	return env
}

// Option is an option for creating environments.
type Option func(*Environment)

// WithExecutor sets the given executor.
func WithExecutor(exec Executor) Option {
	return func(env *Environment) {
		env.executor = exec
	}
}

// WithExecutor sets the given executor.
func WithParseObserver(observe ParseObserver) Option {
	return func(env *Environment) {
		env.observer = observe
	}
}

// Eval parses the given object and runs it in the environment.
func (env *Environment) Eval(obj sx.Object) (sx.Object, error) {
	expr, err := env.Parse(obj)
	if err != nil {
		return nil, err
	}
	expr = env.Rework(expr)
	return env.Run(expr)
}

// Parse the given object.
func (env *Environment) Parse(obj sx.Object) (Expr, error) {
	pf := env.MakeParseFrame()
	return pf.Parse(obj)
}

// Rework the given expression.
func (env *Environment) Rework(expr Expr) Expr {
	rf := env.MakeReworkFrame()
	return expr.Rework(rf)
}

// Run the given expression.
func (env *Environment) Run(expr Expr) (sx.Object, error) {
	if exec := env.executor; exec != nil {
		exec.Reset()
	}
	return env.Execute(expr)
}

func (env *Environment) MakeParseFrame() *ParseEnvironment {
	return &ParseEnvironment{
		binding:  env.binding,
		observer: env.observer,
	}
}

func (env *Environment) MakeReworkFrame() *ReworkFrame {
	return &ReworkFrame{
		binding: env.binding,
	}
}

func (env *Environment) NewDynamicEnvironment() *Environment {
	return &Environment{
		binding:  env.binding,
		executor: env.executor,
		observer: env.observer,
		caller:   env,
	}
}

func (env *Environment) NewLexicalEnvironment(parent *Binding, name string, numBindings int) *Environment {
	return &Environment{
		binding:  MakeChildBinding(parent, name, numBindings),
		executor: env.executor,
		observer: env.observer,
		caller:   env,
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
	dynamicEnv := env.NewDynamicEnvironment()
	res, err := fn.Call(dynamicEnv, args)
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

func (env *Environment) Bind(sym sx.Symbol, obj sx.Object) error {
	return env.binding.Bind(sym, obj)
}
func (env *Environment) BindConst(sym sx.Symbol, obj sx.Object) error {
	return env.binding.BindConst(sym, obj)
}
func (env *Environment) Lookup(sym sx.Symbol) (sx.Object, bool) {
	return env.binding.Lookup(sym)
}
func (env *Environment) Resolve(sym sx.Symbol) (sx.Object, bool) {
	return env.binding.Resolve(sym)
}
func (env *Environment) FindBinding(sym sx.Symbol) *Binding {
	for curr := env.binding; curr != nil; curr = curr.parent {
		if _, found := curr.Lookup(sym); found {
			return curr
		}
	}
	return nil
}
func (env *Environment) MakeNotBoundError(sym sx.Symbol) NotBoundError {
	return NotBoundError{Binding: env.binding, Sym: sym}
}

// NotBoundError signals that a symbol was not found in a binding.
type NotBoundError struct {
	Binding *Binding
	Sym     sx.Symbol
}

func (e NotBoundError) Error() string {
	return fmt.Sprintf("symbol %q not bound in %q", e.Sym.Name(), e.Binding.String())
}
func (env *Environment) Binding() *Binding  { return env.binding }
func (env *Environment) Bindings() *sx.Pair { return env.binding.Bindings() }
