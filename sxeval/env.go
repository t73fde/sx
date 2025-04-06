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
	stackp   *[]sx.Object
}

// tcodata contains everything to implement Tail Call Optimization (tco)
type tcodata struct {
	env  *Environment
	expr Expr
}

type observer struct {
	execute ComputeObserver
	parse   ParseObserver
	improve ImproveObserver
}

// ComputeObserver observes the execution of expressions.
type ComputeObserver interface {
	// BeforeCompute is called immediate before the given expression is computed.
	// The observer may change the expression or abort computation with an error.
	BeforeCompute(*Environment, Expr) (Expr, error)

	// AfterCompute is called immediate after the given expression was computed,
	// resulting in an `sx.Object` and an error.
	AfterCompute(*Environment, Expr, sx.Object, error)
}

func (env *Environment) String() string { return env.binding.name }

// MakeComputeEnvironment creates an environment for later computation of expressions.
func MakeComputeEnvironment(bind *Binding) *Environment {
	stack := make([]sx.Object, 0, 1024)
	return &Environment{
		binding:  bind,
		tco:      &tcodata{},
		observer: &observer{},
		stackp:   &stack,
	}
}

// SetComputeObserver sets the given compute observer.
func (env *Environment) SetComputeObserver(observe ComputeObserver) *Environment {
	env.newObserver().execute = observe
	return env
}

// SetParseObserver sets the given parsing observer.
func (env *Environment) SetParseObserver(observe ParseObserver) *Environment {
	env.newObserver().parse = observe
	return env
}

// SetImproveObserver sets the given improve observer.
func (env *Environment) SetImproveObserver(observe ImproveObserver) *Environment {
	env.newObserver().improve = observe
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
		return sx.Nil(), err
	}
	return env.Run(expr)
}

// Parse the given object.
func (env *Environment) Parse(obj sx.Object) (Expr, error) {
	pe := env.MakeParseEnvironment()
	expr, err := pe.Parse(obj)
	if err != nil {
		return expr, err
	}
	imp := &Improver{
		binding:  env.binding,
		height:   0,
		observer: env.observer.improve,
	}
	imp.base = imp
	return imp.Improve(expr)
}

// Run the given expression.
func (env *Environment) Run(expr Expr) (sx.Object, error) {
	return env.Execute(expr)
}

// MakeParseEnvironment builds a parsing environment to parse a form.
func (env *Environment) MakeParseEnvironment() *ParseEnvironment {
	return &ParseEnvironment{
		binding:  env.binding,
		observer: env.observer.parse,
	}
}

// NewLexicalEnvironment builds a new lexical environment with the given parent
// binding, environment name, and the number of bindings to store.
func (env *Environment) NewLexicalEnvironment(parent *Binding, name string, numBindings int) *Environment {
	result := *env
	result.binding = parent.MakeChildBinding(name, numBindings)
	return &result
}

// Execute the given expression.
func (env *Environment) Execute(expr Expr) (res sx.Object, err error) {
	if exec := env.observer.execute; exec != nil {
		for {
			expr, err = exec.BeforeCompute(env, expr)
			if err == nil {
				res, err = expr.Compute(env)
				if err == nil {
					exec.AfterCompute(env, expr, res, err)
					return res, nil
				}
			}
			exec.AfterCompute(env, expr, res, err)
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

// ApplyMacro executes the Callable in a macro environment.
func (env *Environment) ApplyMacro(name string, fn Callable, args sx.Vector) (sx.Object, error) {
	macroEnv := Environment{
		binding:  env.binding.MakeChildBinding(name, 0),
		tco:      &tcodata{},
		observer: env.observer,
		stackp:   env.stackp,
	}
	return macroEnv.Apply(fn, args)
}

// Apply the given Callable with the arguments.
func (env *Environment) Apply(fn Callable, args sx.Vector) (sx.Object, error) {
	res, err := fn.Call(env, args)
	if err == nil {
		return res, nil
	}
	if err == errExecuteAgain {
		return env.tco.env.Execute(env.tco.expr)
	}
	return nil, env.addExecuteError(&applyExpr{Proc: fn, Args: args}, err)
}

type applyExpr struct {
	Proc Callable
	Args sx.Vector
}

func (ce *applyExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }

func (ce *applyExpr) Unparse() sx.Object {
	args := sx.MakeList(ce.Args...)
	return args.Cons(ce.Proc.(sx.Object))
}

func (ce *applyExpr) Compute(env *Environment) (sx.Object, error) {
	return env.Apply(ce.Proc, ce.Args)
}

func (ce *applyExpr) Print(w io.Writer) (int, error) {
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

// ExecuteError is the error that may occur if an expression is computed.
// It contains the call stack.
type ExecuteError struct {
	Stack []EnvironmentExpr
	err   error
}

// EnvironmentExpr stores the curent environment and the expression to be
// computed, for better error messages.
type EnvironmentExpr struct {
	Env  *Environment
	Expr Expr
}

func (ee ExecuteError) Error() string { return ee.err.Error() }
func (ee ExecuteError) Unwrap() error { return ee.err }

// PrintStack prints the calling stack to an io.Writer.
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

// Bind the symbol to an object value in this environment.
func (env *Environment) Bind(sym *sx.Symbol, obj sx.Object) error {
	return env.binding.Bind(sym, obj)
}

// Lookup a symbol in this environment. If not found, false is returned.
func (env *Environment) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	return env.binding.Lookup(sym)
}

// Resolve a symbol in this environment or in its parent environments. If the
// symbol is not found, a false is returned.
func (env *Environment) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return env.binding.Resolve(sym)
}

// FindBinding returns the binding, where the symbol is bound to a value.
// If no binding was found, nil is returned.
func (env *Environment) FindBinding(sym *sx.Symbol) *Binding {
	for curr := env.binding; curr != nil; curr = curr.parent {
		if _, found := curr.Lookup(sym); found {
			return curr
		}
	}
	return nil
}

// Binding returns the binding of this environment.
func (env *Environment) Binding() *Binding { return env.binding }

// MakeNotBoundError builds an error to signal that a symbol was not bound in
// the environment.
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

// ----- Stack operations

// Stack returns the stack.
func (env *Environment) Stack() []sx.Object { return *env.stackp }

// Push a value to the stack
func (env *Environment) Push(val sx.Object) {
	*env.stackp = append(*env.stackp, val)
}

// Kill some elements on the stack
func (env *Environment) Kill(num int) {
	stack := *env.stackp
	*env.stackp = stack[:len(stack)-num]
}

// Args returns a given number of values on top of the stack as a slice.
func (env *Environment) Args(numargs int) []sx.Object {
	stack := *env.stackp
	sp := len(stack)
	return stack[sp-numargs : sp]
}
