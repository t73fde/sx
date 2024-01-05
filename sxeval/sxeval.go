//-----------------------------------------------------------------------------
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
	// IsPure checks if the callable is independent of a full environment and
	// does not produce any side effects.
	IsPure([]sx.Object) bool

	// Call the value with the given args in the given environment.
	Call(*Environment, []sx.Object) (sx.Object, error)
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
	// Reset prepares for a new execution cylcle. It is typically called by the
	// environment.
	Reset()

	// Execute the expression in an environment and return the result.
	// It may have side-effects, on the given environment, or on the
	// general environment of the system.
	Execute(*Environment, Expr) (sx.Object, error)
}

// SimpleExecutor just computes an expression.
type SimpleExecutor struct{}

// Reset the simple executor.
func (*SimpleExecutor) Reset() {}

// Execute the given expression in the given frame.
func (*SimpleExecutor) Execute(env *Environment, expr Expr) (sx.Object, error) {
	return expr.Compute(env)
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

// Reset all current execution statistics.
func (lex *limitExecutor) Reset() {
	lex.stepCount = 0
	lex.deadline = time.Now().Add(lex.maxDuration)
}

// Execute the given expression in the given frame within the limits of this executor.
func (lex *limitExecutor) Execute(env *Environment, expr Expr) (sx.Object, error) {
	stepCount := lex.stepCount
	stepCount++
	if stepCount > lex.maxSteps {
		return nil, &LimitError{MaxSteps: lex.maxSteps, MaxDuration: 0}
	}
	if stepCount%1000 == 0 && time.Now().After(lex.deadline) {
		return nil, &LimitError{MaxSteps: 0, MaxDuration: lex.maxDuration}
	}
	lex.stepCount = stepCount
	return expr.Compute(env)
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
