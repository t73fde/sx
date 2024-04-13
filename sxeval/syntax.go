//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval

import (
	"errors"

	"t73f.de/r/sx"
)

// Syntax is a form that produces an expression by parsing.
//
// It is not the same as interface `Parser`, because the second parameter is a pair.
type Syntax interface {
	// Parse the args.
	Parse(pf *ParseEnvironment, args *sx.Pair) (Expr, error)
}

// GetSyntax returns the object as a syntax value, if possible.
func GetSyntax(obj sx.Object) (Syntax, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	sy, ok := obj.(Syntax)
	return sy, ok
}

// Special represents all predefined syntax constructing functions implemented in Go.
type Special struct {
	Name string
	Fn   func(*ParseEnvironment, *sx.Pair) (Expr, error)
}

func (sp *Special) IsNil() bool                  { return sp == nil }
func (sp *Special) IsAtom() bool                 { return sp == nil }
func (sp *Special) IsEqual(other sx.Object) bool { return sp == other }
func (sp *Special) String() string               { return "#<special:" + sp.Name + ">" }
func (sp *Special) GoString() string             { return sp.String() }

// Parse the args by calling the syntax function.
func (sp *Special) Parse(pf *ParseEnvironment, args *sx.Pair) (Expr, error) {
	res, err := sp.Fn(pf, args)
	if err != nil {
		var callError CallError
		if !errors.As(err, &callError) {
			err = CallError{Name: sp.Name, Err: err}
		}
	}
	return res, err
}

// ErrNoArgs signals that no arguments were given.
var ErrNoArgs = errors.New("no arguments given")
