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
	"strings"

	"t73f.de/r/sx"
)

// Callable is something that can be called for evaluation.
type Callable interface {
	// IsPure checks if the callable is independent of a full environment and
	// does not produce any side effects.
	IsPure(sx.Vector) bool

	// ExecuteCall calls the value with the given args in the given environment
	// in context of the AST evaluator.
	// Args are transported via env.Args(numargs).
	ExecuteCall(*Environment, int, *Binding) (sx.Object, error)
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
