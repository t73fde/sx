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
	Parse(*ParseEnvironment, *sx.Pair) (Expr, error)
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

// Bind the special form to a given environment.
func (sp *Special) Bind(bi *Binding) error {
	return bi.Bind(sx.MakeSymbol(sp.Name), sp)
}

// BindSpecials will bind many builtins to an environment.
func BindSpecials(bind *Binding, sps ...*Special) error {
	for _, sp := range sps {
		if err := sp.Bind(bind); err != nil {
			return err
		}
	}
	return nil
}

// IsNil returns true if the object must be treated like a sx.Nil() object.
func (sp *Special) IsNil() bool { return sp == nil }

// IsAtom returns true if the object is atomic.
func (sp *Special) IsAtom() bool { return sp == nil }

// IsEqual returns true if the other object has the same content.
func (sp *Special) IsEqual(other sx.Object) bool { return sp == other }

// String returns a string representation.
func (sp *Special) String() string { return "#<special:" + sp.Name + ">" }

// GoString returns a string representation to be used in Go code.
func (sp *Special) GoString() string { return sp.String() }

// Parse the args by calling the syntax function.
func (sp *Special) Parse(pe *ParseEnvironment, args *sx.Pair) (Expr, error) {
	res, err := sp.Fn(pe, args)
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
