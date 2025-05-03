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
	"iter"
	"log/slog"
	"slices"

	"t73f.de/r/sx"
)

// Environment is a runtime object of the current computing environment.
type Environment struct {
	stack  []sx.Object
	bstack []*Binding // saves binding for later restore

	newExpr Expr
	newBind *Binding
	newProg *ProgramExpr

	obCompute   ComputeObserver
	obParse     ParseObserver
	obImprove   ImproveObserver
	obCompile   CompileObserver
	obInterpret InterpretObserver
}

// ComputeObserver observes the execution of expressions.
type ComputeObserver interface {
	// BeforeCompute is called immediate before the given expression is computed.
	// The observer may change the expression or abort computation with an error.
	BeforeCompute(*Environment, Expr, *Binding) (Expr, error)

	// AfterCompute is called immediate after the given expression was computed,
	// resulting in an `sx.Object` and an error.
	AfterCompute(*Environment, Expr, *Binding, sx.Object, error)
}

// MakeEnvironment creates an environment for later parsing, improving, and
// computation of expressions.
func MakeEnvironment() *Environment {
	return &Environment{
		stack:  make([]sx.Object, 0, 1024),
		bstack: make([]*Binding, 0, 256),
	}
}

// SetComputeObserver sets the given compute observer.
func (env *Environment) SetComputeObserver(observe ComputeObserver) *Environment {
	env.obCompute = observe
	return env
}

// SetParseObserver sets the given parsing observer.
func (env *Environment) SetParseObserver(observe ParseObserver) *Environment {
	env.obParse = observe
	return env
}

// SetImproveObserver sets the given improve observer.
func (env *Environment) SetImproveObserver(observe ImproveObserver) *Environment {
	env.obImprove = observe
	return env
}

// SetCompileObserver sets the given compile observer.
func (env *Environment) SetCompileObserver(observe CompileObserver) *Environment {
	env.obCompile = observe
	return env
}

// SetInterpretObserver sets the given compile observer.
func (env *Environment) SetInterpretObserver(observe InterpretObserver) *Environment {
	env.obInterpret = observe
	return env
}

// Eval parses the given object and runs it in the environment.
func (env *Environment) Eval(obj sx.Object, bind *Binding) (sx.Object, error) {
	expr, err := env.Parse(obj, bind)
	if err != nil {
		return sx.Nil(), err
	}
	cexpr, err := env.Compile(expr)
	if err == nil {
		expr = cexpr
	}
	return env.Run(expr, bind)
}

// Parse the given object.
func (env *Environment) Parse(obj sx.Object, bind *Binding) (Expr, error) {
	pe := env.MakeParseEnvironment()
	expr, err := pe.Parse(obj, bind)
	if err != nil {
		return expr, err
	}
	imp := &Improver{
		binding:  bind,
		height:   0,
		observer: env.obImprove,
	}
	imp.base = imp
	return imp.Improve(expr)
}

// Compile the expression
func (env *Environment) Compile(expr Expr) (*ProgramExpr, error) {
	sxc := env.MakeCompiler()
	return sxc.CompileProgram(expr)
}

// MakeParseEnvironment builds a parsing environment to parse a form.
func (env *Environment) MakeParseEnvironment() *ParseEnvironment {
	return &ParseEnvironment{env: env}
}

// MakeCompiler builds a new compiler for the given environment.
func (env *Environment) MakeCompiler() *Compiler {
	sxc := &Compiler{
		level: 0,
		env:   env,
	}
	sxc.resetState()
	return sxc
}

// Run the given expression.
func (env *Environment) Run(expr Expr, bind *Binding) (obj sx.Object, err error) {
	if err = env.Execute(expr, bind); err == nil {
		obj = env.Pop()
	}
	return obj, err
}

// Execute the given expression.
func (env *Environment) Execute(expr Expr, bind *Binding) (err error) {
	for {
		if exec := env.obCompute; exec != nil {
			if expr, err = exec.BeforeCompute(env, expr, bind); err != nil {
				break
			}
		}

		if err = expr.Compute(env, bind); err == nil {
			if exec := env.obCompute; exec != nil {
				exec.AfterCompute(env, expr, bind, env.Top(), nil)
			}
			return nil
		}

		if exec := env.obCompute; exec != nil {
			exec.AfterCompute(env, expr, bind, nil, err)
		}
		if err != errExecuteAgain {
			break
		}
		expr = env.newExpr
		bind = env.newBind
	}
	return env.addExecuteError(expr, err)
}

// ExecuteTCO is called when the expression should be executed at last
// position, aka as tail call order.
func (env *Environment) ExecuteTCO(expr Expr, bind *Binding) error {
	// Just return relevant data for real TCO
	env.newExpr = expr
	env.newBind = bind
	return errExecuteAgain
}

// ApplyMacro executes the Callable in a macro environment, with given number
// of args (which are placed on the stack).
func (env *Environment) ApplyMacro(name string, fn Callable, numargs int, bind *Binding) error {
	macroBind := bind.MakeChildBinding(name, 0)
	return env.Apply(fn, numargs, macroBind)
}

// Apply the given Callable with the given number of arguments (which are on the stack).
func (env *Environment) Apply(fn Callable, numargs int, bind *Binding) error {
	stack := env.stack
	if err := fn.ExecuteCall(env, numargs, bind); err != nil {
		if err == errExecuteAgain {
			return env.Execute(env.newExpr, env.newBind)
		}
		sp := len(stack)
		args := slices.Clone(stack[sp-numargs : sp])
		return env.addExecuteError(&applyErrExpr{Proc: fn, Args: args}, err)
	}
	return nil
}

// applyErrExpr is needed, when an error occurs during `env.Apply`, to give a better
// error message.
type applyErrExpr struct {
	Proc Callable
	Args sx.Vector
}

func (ce *applyErrExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }

// IsPure signals an expression that has no side effects.
func (*applyErrExpr) IsPure() bool { return false }

func (ce *applyErrExpr) Unparse() sx.Object {
	args := sx.MakeList(ce.Args...)
	return args.Cons(ce.Proc.(sx.Object))
}

func (ce *applyErrExpr) Compile(*Compiler, bool) error { return MissingCompileError{Expr: ce} }

func (ce *applyErrExpr) Compute(env *Environment, bind *Binding) error {
	env.PushArgs(ce.Args)
	return env.Apply(ce.Proc, len(ce.Args), bind)
}

func (ce *applyErrExpr) Print(w io.Writer) (int, error) {
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

// ----- Threaded code infrastructure, mostly stack operations

// Reset the stacks.
func (env *Environment) Reset() {
	if env.stack != nil {
		env.stack = env.stack[:0]
	}
	if env.bstack != nil {
		env.bstack = env.bstack[:0]
	}
}

// Stack returns the stack.
func (env *Environment) Stack() iter.Seq[sx.Object] { return slices.Values(env.stack) }

// Size returns the stack size.
func (env *Environment) Size() int { return len(env.stack) }

// Push a value to the stack
func (env *Environment) Push(val sx.Object) { env.stack = append(env.stack, val) }

// PushArgs pushes a whole slice to the stack.
func (env *Environment) PushArgs(args []sx.Object) { env.stack = append(env.stack, args...) }

// Top returns the TOS
func (env *Environment) Top() sx.Object { return env.stack[len(env.stack)-1] }

// Set the value on top of the stack
func (env *Environment) Set(val sx.Object) {
	env.stack[len(env.stack)-1] = val
}

// Pop a value from the stack
func (env *Environment) Pop() sx.Object {
	stack := env.stack
	sp := len(stack) - 1
	val := stack[sp]
	env.stack = stack[0:sp]
	return val
}

// Kill some elements on the stack
func (env *Environment) Kill(num int) { env.stack = env.stack[:len(env.stack)-num] }

// Args returns a given number of values on top of the stack as a slice.
// Use the slice only if you are sure that the stack is not changed.
func (env *Environment) Args(numargs int) sx.Vector {
	sp := len(env.stack)
	return env.stack[sp-numargs : sp]
}

// CopyArgs copies the number of values on top of the stack into a separate slice.
func (env *Environment) CopyArgs(numargs int) []sx.Object { return slices.Clone(env.Args(numargs)) }

// SaveBinding stores a binding for later restore.
func (env *Environment) SaveBinding(bind *Binding) { env.bstack = append(env.bstack, bind) }

// RestoreBinding retrieves the last saved binding.
func (env *Environment) RestoreBinding() *Binding {
	bstack := env.bstack
	sp := len(bstack) - 1
	bind := bstack[sp]
	env.bstack = bstack[0:sp]
	return bind
}
