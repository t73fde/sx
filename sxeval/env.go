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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"t73f.de/r/sx"
)

// Environment is a runtime object of the current computing environment.
type Environment struct {
	binding  *Binding
	tco      *tcodata
	observer *observer
}

// tcodata contains everything to implement Tail Call Optimization (tco)
type tcodata struct {
	env  *Environment
	expr Expr
}

type observer struct {
	execute ExecuteObserver
	parse   ParseObserver
	rework  ReworkObserver
}

// ExecuteObserver observes the execution of expressions.
type ExecuteObserver interface {
	// BeforeExecution is called immediate before the given expression is executed.
	// The observer may change the expression or abort execution with an error.
	BeforeExecution(*Environment, Expr) (Expr, error)

	// AfterExecution is called immediate after the given expression was executed,
	// resulting in an `sx.Object` and an error.
	AfterExecution(*Environment, Expr, sx.Object, error)
}

func (env *Environment) String() string { return env.binding.name }

// MakeExecutionEnvironment creates an environment for later execution of an expression.
func MakeExecutionEnvironment(bind *Binding) *Environment {
	return &Environment{
		binding: bind,
		tco: &tcodata{
			env:  nil,
			expr: nil,
		},
		observer: &observer{
			execute: nil,
			parse:   nil,
			rework:  nil,
		},
	}
}

// SetExecutor sets the given executor.
func (env *Environment) SetExecutor(observe ExecuteObserver) *Environment {
	env.newObserver().execute = observe
	return env
}

// SetParseObserver sets the given parsing observer.
func (env *Environment) SetParseObserver(observe ParseObserver) *Environment {
	env.newObserver().parse = observe
	return env
}

// SetReworkObserver sets the given rework observer.
func (env *Environment) SetReworkObserver(observe ReworkObserver) *Environment {
	env.newObserver().rework = observe
	return env
}

func (env *Environment) newObserver() *observer {
	ob := *env.observer
	env.observer = &ob
	return env.observer
}

// RebindExecutionEnvironment clones the original environment, but uses the
// given binding.
func (env *Environment) RebindExecutionEnvironment(bind *Binding) *Environment {
	result := *env
	result.binding = bind
	return &result
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

// Compile the given object and return the reworked expression.
func (env *Environment) Compile(obj sx.Object) (Expr, error) {
	pe := env.MakeParseEnvironment()
	expr, err := pe.Parse(obj)
	if err != nil {
		return nil, err
	}
	re := env.MakeReworkEnvironment()
	return re.Rework(expr), nil
}

// Parse the given object.
func (env *Environment) Parse(obj sx.Object) (Expr, error) {
	pe := env.MakeParseEnvironment()
	return pe.Parse(obj)
}

// Rework the given expression.
func (env *Environment) Rework(expr Expr) Expr {
	re := env.MakeReworkEnvironment()
	return re.Rework(expr)
}

// Run the given expression.
func (env *Environment) Run(expr Expr) (sx.Object, error) {
	return env.Execute(expr)
}

func (env *Environment) MakeParseEnvironment() *ParseEnvironment {
	return &ParseEnvironment{
		binding:  env.binding,
		observer: env.observer.parse,
	}
}

func (env *Environment) MakeReworkEnvironment() *ReworkEnvironment {
	re := &ReworkEnvironment{
		binding:  env.binding,
		height:   0,
		observer: env.observer.rework,
	}
	re.base = re
	return re
}

func (env *Environment) NewLexicalEnvironment(parent *Binding, name string, numBindings int) *Environment {
	result := *env
	result.binding = parent.MakeChildBinding(name, numBindings)
	return &result
}

// Execute the given expression.
func (env *Environment) Execute(expr Expr) (res sx.Object, err error) {
	if exec := env.observer.execute; exec != nil {
		for {
			expr, err = exec.BeforeExecution(env, expr)
			if err == nil {
				res, err = expr.Compute(env)
				if err == nil {
					exec.AfterExecution(env, expr, res, err)
					return res, nil
				}
			}
			exec.AfterExecution(env, expr, res, err)
			if err == errExecuteAgain {
				env = env.tco.env
				expr = env.tco.expr
				continue
			}
			return res, env.addExecuteError(expr, err)
		}
	}

	for {
		res, err = expr.Compute(env)
		if err == nil {
			return res, nil
		}
		if err == errExecuteAgain {
			env = env.tco.env
			expr = env.tco.expr
			continue
		}
		return res, env.addExecuteError(expr, err)
	}
}

// ExecuteTCO is called when the expression should be executed at last
// position, aka as tail call order.
func (env *Environment) ExecuteTCO(expr Expr) (sx.Object, error) {
	// Uncomment this line to test for non-TCO
	// return env.Execute(expr)

	// Just return relevant data for real TCO
	env.tco.env = env
	env.tco.expr = expr
	return nil, errExecuteAgain
}

// MacroCall executes the Callable in a macro environment.
func (env *Environment) MacroCall(name string, fn Callable, args sx.Vector) (res sx.Object, err error) {
	macroEnv := Environment{
		binding: env.binding.MakeChildBinding(name, 0),
		tco: &tcodata{
			env:  nil,
			expr: nil,
		},
		observer: env.observer,
	}
	return macroEnv.Call(fn, args)
}

// Call the given Callable with the arguments.
func (env *Environment) Call(fn Callable, args sx.Vector) (res sx.Object, err error) {
	switch len(args) {
	case 0:
		res, err = fn.Call0(env)
	case 1:
		res, err = fn.Call1(env, args[0])
	case 2:
		res, err = fn.Call2(env, args[0], args[1])
	default:
		res, err = fn.Call(env, args)
	}
	if err == nil {
		return res, nil
	}
	if err == errExecuteAgain {
		return env.tco.env.Execute(env.tco.expr)
	}
	return nil, env.addExecuteError(&callableExpr{Proc: fn, Args: args}, err)
}

type callableExpr struct {
	Proc Callable
	Args sx.Vector
}

func (ce *callableExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }

func (ce *callableExpr) Unparse() sx.Object {
	args := sx.MakeList(ce.Args...)
	return args.Cons(ce.Proc.(sx.Object))
}

func (ce *callableExpr) Improve(*ReworkEnvironment) Expr { return ce }

func (ce *callableExpr) Compute(env *Environment) (sx.Object, error) {
	return env.Call(ce.Proc, ce.Args)
}

func (ce *callableExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{call %v %v}", ce.Proc, ce.Args)
}

// errExecuteAgain is a non-error error signalling that the given expression should be
// executed again in the given binding.
var errExecuteAgain = errors.New("TCO trampoline")

func (env *Environment) addExecuteError(expr Expr, err error) error {
	var execError ExecuteError
	if errors.As(err, &execError) {
		execError.Stack = append(execError.Stack, EnvironmentExpr{env, expr})
		return execError
	}
	return ExecuteError{
		Stack: []EnvironmentExpr{{env, expr}},
		err:   err,
	}
}

type ExecuteError struct {
	Stack []EnvironmentExpr
	err   error
}
type EnvironmentExpr struct {
	Env  *Environment
	Expr Expr
}

func (ee ExecuteError) Error() string { return ee.err.Error() }
func (ee ExecuteError) Unwrap() error { return ee.err }
func (ee ExecuteError) PrintStack(w io.Writer, prefix string, logger *slog.Logger, logmsg string) {
	for i, elem := range ee.Stack {
		val := elem.Expr.Unparse()
		if logger != nil {
			logger.Debug(logmsg, "env", elem.Env, "expr", val)
		}
		_, _ = fmt.Fprintf(w, "%v%d: env: %v, expr: %T/%v\n", prefix, i, elem.Env, val, val)
		_, _ = io.WriteString(w, "   exp: ")
		_, _ = elem.Expr.Print(w)
		_, _ = io.WriteString(w, "\n")
	}

}

func (env *Environment) Bind(sym *sx.Symbol, obj sx.Object) error {
	return env.binding.Bind(sym, obj)
}
func (env *Environment) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	return env.binding.Lookup(sym)
}
func (env *Environment) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return env.binding.Resolve(sym)
}
func (env *Environment) LookupNWithError(sym *sx.Symbol, n int) (sx.Object, error) {
	if obj, found := env.binding.LookupN(sym, n); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(sym)
}
func (env *Environment) ResolveNWithError(sym *sx.Symbol, skip int) (sx.Object, error) {
	if obj, found := env.binding.ResolveN(sym, skip); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(sym)
}
func (env *Environment) ResolveUnboundWithError(sym *sx.Symbol) (sx.Object, error) {
	if obj, found := env.binding.Resolve(sym); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(sym)
}

func (env *Environment) FindBinding(sym *sx.Symbol) *Binding {
	for curr := env.binding; curr != nil; curr = curr.parent {
		if _, found := curr.Lookup(sym); found {
			return curr
		}
	}
	return nil
}

func (env *Environment) Binding() *Binding { return env.binding }

func (env *Environment) MakeNotBoundError(sym *sx.Symbol) NotBoundError {
	return NotBoundError{Binding: env.binding, Sym: sym}
}

// NotBoundError signals that a symbol was not found in a binding.
type NotBoundError struct {
	Binding *Binding
	Sym     *sx.Symbol
}

func (e NotBoundError) Error() string {
	var sb strings.Builder
	if e.Sym == nil {
		sb.WriteString("symbol == nil, not bound in ")
	} else {
		fmt.Fprintf(&sb, "symbol %q not bound in ", e.Sym.String())
	}
	second := false
	for binding := e.Binding; binding != nil; binding = binding.Parent() {
		if second {
			sb.WriteString("->")
		}
		fmt.Fprintf(&sb, "%q", binding.Name())
		second = true
	}
	return sb.String()
}
