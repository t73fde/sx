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
	"strings"

	"t73f.de/r/sx"
)

// Environment is a runtime object of the current computing environment.
type Environment struct {
	stack []sx.Object

	globals *Binding

	newExpr Expr
	newBind *Binding

	obCompute ComputeObserver
	obParse   ParseObserver
	obImprove ImproveObserver
}

func (env *Environment) String() string {
	const stackElems = 5

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d: (", len(env.stack))
	sp := len(env.stack) - 1
	for i := 0; i < stackElems && sp-i >= 0; i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		_, _ = sx.Print(&sb, env.stack[sp-i])
	}
	if sp-stackElems >= 0 {
		sb.WriteString(" ...")
	}
	sb.WriteByte(')')
	return sb.String()
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
func MakeEnvironment(globals *Binding) *Environment {
	return &Environment{
		stack:   make([]sx.Object, 0, 1024),
		globals: globals,
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

// Eval parses the given object and runs it in the environment.
func (env *Environment) Eval(obj sx.Object, bind *Binding) (sx.Object, error) {
	expr, err := env.Parse(obj, bind)
	if err != nil {
		return sx.Nil(), err
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
		globals:  env.globals,
		observer: env.obImprove,
	}
	imp.base = imp
	return imp.Improve(expr)
}

// MakeParseEnvironment builds a parsing environment to parse a form.
func (env *Environment) MakeParseEnvironment() *ParseEnvironment {
	return &ParseEnvironment{env: env}
}

// Run the given expression.
func (env *Environment) Run(expr Expr, bind *Binding) (sx.Object, error) {
	return env.Execute(expr, bind)
}

// Execute the given expression.
func (env *Environment) Execute(expr Expr, bind *Binding) (res sx.Object, err error) {
	for {
		if exec := env.obCompute; exec != nil {
			if expr, err = exec.BeforeCompute(env, expr, bind); err != nil {
				break
			}
		}

		if res, err = expr.Compute(env, bind); err == nil {
			if exec := env.obCompute; exec != nil {
				exec.AfterCompute(env, expr, bind, res, nil)
			}
			return res, nil
		}

		if exec := env.obCompute; exec != nil {
			exec.AfterCompute(env, expr, bind, res, err)
		}
		if err != errExecuteAgain {
			break
		}
		expr = env.newExpr
		bind = env.newBind
	}
	return res, env.addExecuteError(expr, bind, err)
}

// ExecuteTCO is called when the expression should be executed at last
// position, aka as tail call order.
func (env *Environment) ExecuteTCO(expr Expr, bind *Binding) (sx.Object, error) {
	// Uncomment this line to test for non-TCO
	// return env.Execute(expr, bind)

	// Just return relevant data for real TCO
	env.newExpr = expr
	env.newBind = bind
	return nil, errExecuteAgain
}

// ApplyMacro executes the Callable in a macro environment.
func (env *Environment) ApplyMacro(name string, fn Callable, args sx.Vector, bind *Binding) (sx.Object, error) {
	macroBind := bind.MakeChildBinding(name, 0)
	return env.Apply(fn, args, macroBind)
}

// Apply the given Callable with the arguments.
func (env *Environment) Apply(fn Callable, args sx.Vector, bind *Binding) (res sx.Object, err error) {
	if res, err = fn.ExecuteCall(env, args, bind); err == nil {
		return res, nil
	}
	if err == errExecuteAgain {
		return env.Execute(env.newExpr, env.newBind)
	}
	return nil, env.addExecuteError(&applyErrExpr{Proc: fn, Args: args}, bind, err)
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

func (ce *applyErrExpr) Compute(env *Environment, bind *Binding) (sx.Object, error) {
	return env.Apply(ce.Proc, ce.Args, bind)
}

func (ce *applyErrExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{call %v %v}", ce.Proc, ce.Args)
}

// errExecuteAgain is a non-error error signalling that the given expression should be
// executed again in the given binding.
var errExecuteAgain = errors.New("TCO trampoline")

func (env *Environment) addExecuteError(expr Expr, bind *Binding, err error) error {
	var execError ExecuteError
	if errors.As(err, &execError) {
		execError.CallStack = append(
			execError.CallStack, ExecuteState{env, slices.Clone(env.stack), bind, expr})
		return execError
	}
	return ExecuteError{
		CallStack: []ExecuteState{{env, slices.Clone(env.stack), bind, expr}},
		err:       err,
	}
}

// ExecuteError is the error that may occur if an expression is computed.
// It contains the call stack.
type ExecuteError struct {
	CallStack []ExecuteState
	err       error
}

// ExecuteState stores the curent environment, the current stack content,
// the current binding, and the expression to be computed, for better error messages.
type ExecuteState struct {
	Env   *Environment
	Stack []sx.Object
	Bind  *Binding
	Expr  Expr
}

func (ee ExecuteError) Error() string { return ee.err.Error() }
func (ee ExecuteError) Unwrap() error { return ee.err }

const lengthDataStackOnError = 7

// PrintCallStack prints the calling stack to an io.Writer.
func (ee ExecuteError) PrintCallStack(w io.Writer, prefix string, logger *slog.Logger, logmsg string) {
	for i, elem := range ee.CallStack {
		val := elem.Expr.Unparse()
		stack := elem.Stack
		if logger != nil {
			logger.Debug(logmsg, "env", elem.Env, "stack", stack, "binding", elem.Bind, "expr", val)
		}
		_, _ = fmt.Fprintf(w, "%v%2d: expr = %T/%v\n", prefix, i, val, val)
		_, _ = fmt.Fprintf(w, "%v    bind = %v\n", prefix, elem.Bind)
		_, _ = fmt.Fprintf(w, "%v    stack= %d:(", prefix, len(stack))
		for j := 0; j < lengthDataStackOnError && len(stack)-j > 0; j++ {
			if j > 0 {
				_, _ = io.WriteString(w, " ")
			}
			_, _ = sx.Print(w, stack[len(stack)-1-j])
		}
		if len(stack)-lengthDataStackOnError > 0 {
			_, _ = io.WriteString(w, " ...")
		}
		_, _ = io.WriteString(w, ")\n")
		_, _ = io.WriteString(w, prefix)
		_, _ = io.WriteString(w, "    code = ")
		_, _ = elem.Expr.Print(w)
		_, _ = io.WriteString(w, "\n")
	}

}

// ----- Stack operations

// Stack returns the stack.
func (env *Environment) Stack() iter.Seq[sx.Object] { return slices.Values(env.stack) }

// Size returns the stack size.
func (env *Environment) Size() int { return len(env.stack) }

// Push a value to the stack
func (env *Environment) Push(val sx.Object) { env.stack = append(env.stack, val) }

// Kill some elements on the stack
func (env *Environment) Kill(num int) { env.stack = env.stack[:len(env.stack)-num] }

// Args returns a given number of values on top of the stack as a slice.
func (env *Environment) Args(numargs int) []sx.Object {
	sp := len(env.stack)
	return env.stack[sp-numargs : sp : sp]
}
