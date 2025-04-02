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
	"slices"
	"strings"

	"t73f.de/r/sx"
)

// Environment is a runtime object of the current computing environment.
type Environment struct {
	binding  *Binding
	tco      *tcodata
	observer *observer
	stack    sx.Vector
}

// tcodata contains everything to implement Tail Call Optimization (tco)
type tcodata struct {
	env  *Environment
	expr Expr
}

type observer struct {
	execute ExecuteObserver
	parse   ParseObserver
	improve ImproveObserver
	compile CompileObserver
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
			improve: nil,
			compile: nil,
		},
		stack: nil,
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

// SetImproveObserver sets the given improve observer.
func (env *Environment) SetImproveObserver(observe ImproveObserver) *Environment {
	env.newObserver().improve = observe
	return env
}

// SetCompileObserver sets the given compile observer.
func (env *Environment) SetCompileObserver(observe CompileObserver) *Environment {
	env.newObserver().compile = observe
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

// Compile the expression
func (env *Environment) Compile(expr Expr) (*ProgramExpr, error) {
	sxc := Compiler{
		env:      env,
		observer: env.observer.compile,
		program:  nil,
		curStack: 0,
		maxStack: 0,
	}
	if err := sxc.Compile(expr, true); err != nil {
		return nil, err
	}
	if sxc.curStack != 1 {
		panic(fmt.Sprintf("wrong stack position: %d", sxc.curStack))
	}
	return &ProgramExpr{
		program:   slices.Clip(sxc.program),
		stacksize: sxc.maxStack,
		source:    expr,
	}, nil
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

// ApplyMacro executes the Callable in a macro environment.
func (env *Environment) ApplyMacro(name string, fn Callable, args sx.Vector) (res sx.Object, err error) {
	macroEnv := Environment{
		binding: env.binding.MakeChildBinding(name, 0),
		tco: &tcodata{
			env:  nil,
			expr: nil,
		},
		observer: env.observer,
	}
	return macroEnv.Apply(fn, args)
}

// Apply the given Callable with the arguments.
func (env *Environment) Apply(fn Callable, args sx.Vector) (res sx.Object, err error) {
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
		execError.Stack = append(execError.Stack, CallInfo{env, expr})
		return execError
	}
	return ExecuteError{
		Stack: []CallInfo{{env, expr}},
		err:   err,
	}
}

// ExecuteError is the error that may occur if an expression is computed.
// It contains the call stack.
type ExecuteError struct {
	Stack []CallInfo
	err   error
}

// CallInfo stores the curent environment and the expression to be
// computed, for better error messages.
type CallInfo struct {
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

// LookupNWithError tries to look up a symbol in a parent environment of this
// environment, which is n levels away. If this is not possible,
// a NotBoundError is returned.
func (env *Environment) LookupNWithError(sym *sx.Symbol, n int) (sx.Object, error) {
	if obj, found := env.binding.LookupN(sym, n); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(sym)
}

// ResolveNWithError tries to resolve the symbol in this environment, after
// skipping some parent levels. If it is not possible to resolve the symbol,
// a NotBoundError is returned.
func (env *Environment) ResolveNWithError(sym *sx.Symbol, skip int) (sx.Object, error) {
	if obj, found := env.binding.ResolveN(sym, skip); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(sym)
}

// ResolveUnboundWithError tries to resolve a symbol in this environment,
// or in its parent environments.
// If not possible, a NotBoundError is returned.
func (env *Environment) ResolveUnboundWithError(sym *sx.Symbol) (sx.Object, error) {
	if obj, found := env.binding.Resolve(sym); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(sym)
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

// ----- Threaded code infrastructure, mostly stack operations

// Reset the stack.
func (env *Environment) Reset() {
	if env.stack != nil {
		env.stack = env.stack[:0]
	}
}

// Stack returns the stack.
func (env *Environment) Stack() sx.Vector { return env.stack }

// Push a value to the stack
func (env *Environment) Push(val sx.Object) {
	env.stack = append(env.stack, val)
}

// Pop a value from the stack
func (env *Environment) Pop() sx.Object {
	sp := len(env.stack) - 1
	val := env.stack[sp]
	env.stack = env.stack[0:sp]
	return val
}

// Top returns the value on top of the stack
func (env *Environment) Top() sx.Object { return env.stack[len(env.stack)-1] }

// Set the value on top of the stack
func (env *Environment) Set(val sx.Object) { env.stack[len(env.stack)-1] = val }

// Kill1 removes the TOS.
func (env *Environment) Kill1() { env.stack = env.stack[:len(env.stack)-1] }

// Kill some elements on the stack
func (env *Environment) Kill(num int) { env.stack = env.stack[:len(env.stack)-num] }

// Args returns a given number of values on top of the stack as a slice.
func (env *Environment) Args(numargs int) sx.Vector {
	sp := len(env.stack)
	return env.stack[sp-numargs : sp]
}
