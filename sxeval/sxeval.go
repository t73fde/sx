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
	"strings"
	"time"

	"zettelstore.de/sx.fossil"
)

// Callable is a value that can be called for evaluation.
type Callable interface {
	// Call the value with the given args and frame.
	Call(*Frame, []sx.Object) (sx.Object, error)
}

// GetCallable returns the object as a Callable, if possible.
func GetCallable(obj sx.Object) (Callable, bool) {
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
	// Execute the expression in a frame and return the result.
	// It may have side-effects, on the given frame, or on the
	// general environment of the system.
	Execute(*Frame, Expr) (sx.Object, error)
}

// SimpleExecutor just computes an expression.
type SimpleExecutor struct{}

// Execute the given expression in the given frame.
func (*SimpleExecutor) Execute(frame *Frame, expr Expr) (sx.Object, error) {
	return expr.Compute(frame)
}

// limitExecutor computes just some steps and/or for a given time limit.
type limitExecutor struct {
	stepCount   uint
	maxSteps    uint
	deadline    time.Time
	maxDuration time.Duration
}

// MakeLimitExecutor creates a new Executor with the given limits.
func MakeLimitExecutor(maxDuration time.Duration, maxSteps uint) Executor {
	if maxDuration <= 0 && maxSteps == 0 {
		return &SimpleExecutor{}
	}
	return &limitExecutor{
		stepCount:   0,
		maxSteps:    maxSteps,
		deadline:    time.Now().Add(maxDuration),
		maxDuration: maxDuration,
	}
}

// Execute the given expression in the given frame within the limits of this executor.
func (lex *limitExecutor) Execute(frame *Frame, expr Expr) (sx.Object, error) {
	stepCount := lex.stepCount
	stepCount++
	if stepCount > lex.maxSteps {
		return nil, &LimitError{MaxSteps: lex.maxSteps, MaxDuration: 0}
	}
	if stepCount%1000 == 0 && time.Now().After(lex.deadline) {
		return nil, &LimitError{MaxSteps: 0, MaxDuration: lex.maxDuration}
	}
	lex.stepCount = stepCount
	return expr.Compute(frame)
}

// LimitError is signaled when execution limits are exceeded.
type LimitError struct {
	MaxSteps    uint
	MaxDuration time.Duration
}

// Error returns a string representation of this error.
func (limErr *LimitError) Error() string {
	if maxSteps := limErr.MaxSteps; maxSteps > 0 {
		return fmt.Sprintf("execution limit of %v steps exceeded", maxSteps)
	}
	return fmt.Sprintf("execution time limit of %v exceeded", limErr.MaxDuration)
}
