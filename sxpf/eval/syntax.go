//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package eval

import (
	"errors"
	"io"

	"zettelstore.de/sx.fossil/sxpf"
)

// SpecialFn is the signature of all syntax constructing functions.
type SyntaxFn func(*Engine, sxpf.Environment, *sxpf.Pair) (Expr, error)

// Syntax represents all syntax constructing functions implemented in Go.
type Syntax struct {
	name string
	fn   SyntaxFn
}

// MakeSyntax creates a new special function.
func MakeSyntax(name string, fn SyntaxFn) *Syntax {
	return &Syntax{
		name: name,
		fn:   fn,
	}
}

func (sy *Syntax) IsNil() bool  { return sy == nil }
func (sy *Syntax) IsAtom() bool { return sy == nil }
func (sy *Syntax) IsEql(other sxpf.Object) bool {
	if sy == other {
		return true
	}
	if sy.IsNil() {
		return sxpf.IsNil(other)
	}
	if otherSy, ok := other.(*Syntax); ok {
		if sy.fn == nil {
			return otherSy.fn == nil
		}
		return sy.name == otherSy.name
	}
	return false
}
func (sy *Syntax) IsEqual(other sxpf.Object) bool { return sy.IsEql(other) }
func (sy *Syntax) String() string                 { return sy.Repr() }
func (sy *Syntax) Repr() string                   { return sxpf.Repr(sy) }
func (sy *Syntax) Print(w io.Writer) (int, error) {
	return sxpf.WriteStrings(w, "#<syntax:", sy.name, ">")
}

// Parse the args by calling the syntax function.
func (sy *Syntax) Parse(eng *Engine, env sxpf.Environment, args *sxpf.Pair) (Expr, error) {
	res, err := sy.fn(eng, env, args)
	if err != nil {
		if _, ok := err.(CallError); !ok {
			err = CallError{Name: sy.name, Err: err}
		}
	}
	return res, err
}

// GetSyntax returns the object as a syntax value if possible.
func GetSyntax(obj sxpf.Object) (*Syntax, bool) {
	if sxpf.IsNil(obj) {
		return nil, false
	}
	syn, ok := obj.(*Syntax)
	return syn, ok
}

// Special is a special form that produces an expression by parsing.
type Special interface {
	// Parse the args.
	Parse(eng *Engine, env sxpf.Environment, args *sxpf.Pair) (Expr, error)
}

// GetSpecial returns the object as a special value if possible.
func GetSpecial(obj sxpf.Object) (Special, bool) {
	if sxpf.IsNil(obj) {
		return nil, false
	}
	sp, ok := obj.(Special)
	return sp, ok
}

// ErrNoArgs signals that no arguments were given
var ErrNoArgs = errors.New("no arguments given")
