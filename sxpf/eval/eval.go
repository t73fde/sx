//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package eval allows to evaluate s-expressions. Evaluation is splitted into
// parsing that s-expression and executing the result of the parsed expression.
// This is done to reduce syntax checks.
package eval

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"zettelstore.de/sx.fossil/sxpf"
)

// Callable is a value that can be called for evaluation.
type Callable interface {
	// Call the value with the given args and environment.
	Call(*Engine, sxpf.Environment, []sxpf.Object) (sxpf.Object, error)
}

// GetCallable returns the object as a Callable, if possible.
func GetCallable(obj sxpf.Object) (Callable, bool) {
	res, ok := obj.(Callable)
	return res, ok
}

// CallError encapsulate an error that occured during a call.
type CallError struct {
	Name string
	Err  error
}

func (e CallError) Unwrap() error { return e.Err }
func (e CallError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Name)
	sb.WriteString(": ")
	sb.WriteString(e.Err.Error())
	return sb.String()
}

// Executor is about controlling the execution of expressions.
// Do not call the method `Expr.Execute` directly, call the executor to do this.
type Executor interface {
	// Execute the expression in an environment and return the result.
	// It may have side-effects, on the given environment, or on the
	// general environment of the system.
	Execute(*Engine, sxpf.Environment, Expr) (sxpf.Object, error)
}

type simpleExecutor struct{}

var mySimpleExecutor simpleExecutor

// Execute the given expression in the given environment of the given engine.
func (*simpleExecutor) Execute(eng *Engine, env sxpf.Environment, expr Expr) (sxpf.Object, error) {
	return expr.Compute(eng, env)
}

// Engine is the collection of all relevant data element to execute / evaluate an object.
type Engine struct {
	sf        sxpf.SymbolFactory
	root      sxpf.Environment
	toplevel  sxpf.Environment
	pars      Parser
	reworkOpt *ReworkOptions
	exec      Executor
	bNames    map[uintptr]string
}

// MakeEngine creates a new engine.
func MakeEngine(sf sxpf.SymbolFactory, root sxpf.Environment) *Engine {
	return &Engine{
		sf:        sf,
		root:      root,
		toplevel:  root,
		pars:      &myDefaultParser,
		reworkOpt: &ReworkOptions{ResolveEnv: root},
		exec:      nil,
		bNames:    make(map[uintptr]string, 128),
	}
}

// SymbolFactory returns the symbol factory of the engine.
func (eng *Engine) SymbolFactory() sxpf.SymbolFactory { return eng.sf }

// RootEnvironment returns the root environment of the engine.
func (eng *Engine) RootEnvironment() sxpf.Environment { return eng.root }

// SetToplevelEnv sets the given environment as the top-level environment.
// It must be the root environment or a child of it.
func (eng *Engine) SetToplevelEnv(env sxpf.Environment) error {
	root := sxpf.RootEnv(env)
	if root != eng.root {
		return fmt.Errorf("root of %v is not root of engine %v: %v", env, eng.root, root)
	}
	eng.toplevel = env
	return nil
}

// GetToplevelEnv returns the current top-level environment.
func (eng *Engine) GetToplevelEnv() sxpf.Environment { return eng.toplevel }

// SetParser updates the current s-expression parser of the engine.
func (eng *Engine) SetParser(p Parser) Parser {
	orig := eng.pars
	if p == nil {
		p = &myDefaultParser
	}
	eng.pars = p
	return orig
}

// SetExecutor updates the executor pf parsed s-expressions.
func (eng *Engine) SetExecutor(e Executor) Executor {
	orig := eng.exec
	eng.exec = e
	if orig != nil {
		return orig
	}
	return &mySimpleExecutor
}

// SetReworkOptions sets the rework options for this engine and returns the previous value.
//
// If nil is given, the current value is unchanged and it is just returned.
func (eng *Engine) SetReworkOptions(ro *ReworkOptions) *ReworkOptions {
	if ro != nil {
		old := eng.reworkOpt
		eng.reworkOpt = ro
		return old
	}
	return eng.reworkOpt
}

// Eval parses the given object and executes it in the environment.
func (eng *Engine) Eval(env sxpf.Environment, obj sxpf.Object) (sxpf.Object, error) {
	expr, err := eng.Parse(env, obj)
	if err != nil {
		return nil, err
	}
	expr = expr.Rework(eng.reworkOpt, env)
	return eng.Execute(env, expr)
}

// Parse the given object in the given environment.
func (eng *Engine) Parse(env sxpf.Environment, obj sxpf.Object) (Expr, error) {
	return eng.pars.Parse(eng, env, obj)
}

// Rework the given expression with the options stored in the engine.
func (eng *Engine) Rework(env sxpf.Environment, expr Expr) Expr {
	return expr.Rework(eng.reworkOpt, env)
}

// Execute the given expression in the given environment.
func (eng *Engine) Execute(env sxpf.Environment, expr Expr) (sxpf.Object, error) {
	if exec := eng.exec; exec != nil {
		for {
			res, err := eng.exec.Execute(eng, env, expr)
			if err == nil {
				return res, nil
			}
			if again, ok := err.(executeAgain); ok {
				env, expr = again.Env, again.Expr
				continue
			}
			return res, err
		}
	}

	for {
		res, err := expr.Compute(eng, env)
		if err == nil {
			return res, nil
		}
		if again, ok := err.(executeAgain); ok {
			env, expr = again.Env, again.Expr
			continue
		}
		return res, err
	}
}

// ExecuteTCO the given expression in the given environment, but tail-call optimized.
func (eng *Engine) ExecuteTCO(env sxpf.Environment, expr Expr) (sxpf.Object, error) {
	return nil, executeAgain{Env: env, Expr: expr}
}

// executeAgain is a non-error error signalling that the given expression should be
// executed again in the given environment.
type executeAgain struct {
	Env  sxpf.Environment
	Expr Expr
}

func (e executeAgain) Error() string { return fmt.Sprintf("Again: %v", e.Expr) }

func (eng *Engine) Call(env sxpf.Environment, fn Callable, args []sxpf.Object) (sxpf.Object, error) {
	res, err := fn.Call(eng, env, args)
	if err == nil {
		return res, nil
	}
	if again, ok := err.(executeAgain); ok {
		return eng.Execute(again.Env, again.Expr)
	}
	return nil, err
}

// BindBuiltinA binds a standard builtin function to the given name in the engine's root environment.
func (eng *Engine) BindBuiltinA(name string, fn BuiltinA) error {
	eng.bNames[reflect.ValueOf(fn).Pointer()] = name
	return eng.Bind(name, fn)
}

// BindBuiltinEEA binds a special builtin function to the given name in the engine's root environment.
func (eng *Engine) BindBuiltinEEA(name string, fn BuiltinEEA) error {
	eng.bNames[reflect.ValueOf(fn).Pointer()] = name
	return eng.Bind(name, fn)
}

// BindSyntax binds a syntax parser to the given name in the engine's root environment.
// It also binds the parser to the symbol directly.
func (eng *Engine) BindSyntax(name string, fn SyntaxFn) error {
	return eng.Bind(name, MakeSyntax(name, fn))
}

// Bind a given object to a symbol of the given name in the engine's root environment.
func (eng *Engine) Bind(name string, obj sxpf.Object) error {
	return eng.root.Bind(eng.sf.MustMake(name), obj)
}

// BuiltinName returns the name of the given Builtin.
func (eng *Engine) BuiltinName(b Builtin) string {
	ptr := reflect.ValueOf(b).Pointer()
	if eng == nil {
		return strconv.FormatUint(uint64(ptr), 16)
	}
	if name, found := eng.bNames[ptr]; found {
		return name
	}
	return ""
}
