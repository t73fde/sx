//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval

import (
	"errors"
	"io"

	"zettelstore.de/sx.fossil"
)

// Syntax is a form that produces an expression by parsing.
//
// It is not the same as interface `Parser`, because the second parameter is a pair.
type Syntax interface {
	// Parse the args.
	Parse(pf *ParseFrame, args *sx.Pair) (Expr, error)
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
	Fn   func(*ParseFrame, *sx.Pair) (Expr, error)
}

func (sp *Special) IsNil() bool  { return sp == nil }
func (sp *Special) IsAtom() bool { return sp == nil }
func (sp *Special) IsEqual(other sx.Object) bool {
	if sp == other {
		return true
	}
	if sp.IsNil() {
		return sx.IsNil(other)
	}
	if otherSp, ok := other.(*Special); ok {
		if sp.Fn == nil {
			return otherSp.Fn == nil
		}
		return sp.Name == otherSp.Name
	}
	return false
}
func (sp *Special) String() string { return sp.Repr() }
func (sp *Special) Repr() string   { return sx.Repr(sp) }
func (sp *Special) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<special:", sp.Name, ">")
}

// Parse the args by calling the syntax function.
func (sp *Special) Parse(pf *ParseFrame, args *sx.Pair) (Expr, error) {
	res, err := sp.Fn(pf, args)
	if err != nil {
		if _, ok := err.(CallError); !ok {
			err = CallError{Name: sp.Name, Err: err}
		}
	}
	return res, err
}

// ErrNoArgs signals that no arguments were given.
var ErrNoArgs = errors.New("no arguments given")
