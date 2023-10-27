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
	sp, ok := obj.(Syntax)
	return sp, ok
}

// Special represents all predefined syntax constructing functions implemented in Go.
type Special struct {
	Name string
	Fn   func(*ParseFrame, *sx.Pair) (Expr, error)
}

func (sy *Special) IsNil() bool  { return sy == nil }
func (sy *Special) IsAtom() bool { return sy == nil }
func (sy *Special) IsEqual(other sx.Object) bool {
	if sy == other {
		return true
	}
	if sy.IsNil() {
		return sx.IsNil(other)
	}
	if otherSy, ok := other.(*Special); ok {
		if sy.Fn == nil {
			return otherSy.Fn == nil
		}
		return sy.Name == otherSy.Name
	}
	return false
}
func (sy *Special) String() string { return sy.Repr() }
func (sy *Special) Repr() string   { return sx.Repr(sy) }
func (sy *Special) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<syntax:", sy.Name, ">")
}

// Parse the args by calling the syntax function.
func (sy *Special) Parse(pf *ParseFrame, args *sx.Pair) (Expr, error) {
	res, err := sy.Fn(pf, args)
	if err != nil {
		if _, ok := err.(CallError); !ok {
			err = CallError{Name: sy.Name, Err: err}
		}
	}
	return res, err
}

// ErrNoArgs signals that no arguments were given.
var ErrNoArgs = errors.New("no arguments given")
