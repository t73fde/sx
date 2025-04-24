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

// Package sxbuiltins contains functions that help to build builtin functions.
package sxbuiltins

import (
	"fmt"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// GetSymbol returns the given argument as a symbol, and checks for errors.
func GetSymbol(arg sx.Object, pos int) (*sx.Symbol, error) {
	if sym, ok := sx.GetSymbol(arg); ok {
		return sym, nil
	}
	return nil, fmt.Errorf("argument %d is not a symbol, but %T/%v", pos+1, arg, arg)
}

// GetString returns the given argument as a string, and checks for errors.
func GetString(arg sx.Object, pos int) (sx.String, error) {
	if s, isString := sx.GetString(arg); isString {
		return s, nil
	}
	return sx.String{}, fmt.Errorf("argument %d is not a string, but %T/%v", pos+1, arg, arg)
}

// GetNumber returns the given argument as a number, and checks for errors.
func GetNumber(arg sx.Object, pos int) (sx.Number, error) {
	if num, ok := sx.GetNumber(arg); ok {
		return num, nil
	}
	return nil, fmt.Errorf("argument %d is not a number, but %T/%v", pos+1, arg, arg)
}

// GetList returns the given argument as a list, and checks for errors.
func GetList(arg sx.Object, pos int) (*sx.Pair, error) {
	if sx.IsNil(arg) {
		return nil, nil
	}
	if pair, isPair := sx.GetPair(arg); isPair {
		return pair, nil
	}
	return nil, fmt.Errorf("argument %d is not a list, but %T/%v", pos+1, arg, arg)
}

// GetPair returns the given argument as a non-nil list, and checks for errors.
func GetPair(arg sx.Object, pos int) (*sx.Pair, error) {
	if !sx.IsNil(arg) {
		if pair, isPair := sx.GetPair(arg); isPair {
			return pair, nil
		}
	}
	return nil, fmt.Errorf("argument %d is not a pair, but %T/%v", pos+1, arg, arg)
}

// GetVector returns the given argument as a vector, and checks for errors.
func GetVector(arg sx.Object, pos int) (sx.Vector, error) {
	if v, ok := sx.GetVector(arg); ok {
		return v, nil
	}
	return nil, fmt.Errorf("argument %d is not a vector, but %T/%v", pos+1, arg, arg)
}

// GetSequence returns the given argument as a sequence, and checks for errors.
func GetSequence(arg sx.Object, pos int) (sx.Sequence, error) {
	if seq, ok := sx.GetSequence(arg); ok {
		if sx.IsNil(seq) {
			return sx.Nil(), nil
		}
		return seq, nil
	}
	return nil, fmt.Errorf("argument %d is not a sequence, but %T/%v", pos+1, arg, arg)
}

// GetBinding returns the given argument as a binding, and checks for errors.
func GetBinding(arg sx.Object, pos int) (*sxeval.Binding, error) {
	if bind, ok := sxeval.GetBinding(arg); ok {
		return bind, nil
	}
	return nil, fmt.Errorf("argument %d is not a binding, but %T/%v", pos+1, arg, arg)
}

// GetCallable returns the given argument as a callable, and checks for errors.
func GetCallable(arg sx.Object, pos int) (sxeval.Callable, error) {
	if fn, ok := sxeval.GetCallable(arg); ok {
		return fn, nil
	}
	return nil, fmt.Errorf("argument %d is not a function, but %T/%v", pos+1, arg, arg)
}

// GetExprObj returns the given argument as a expression object, and checks for errors.
func GetExprObj(arg sx.Object, pos int) (*sxeval.ExprObj, error) {
	if fn, ok := sxeval.GetExprObj(arg); ok {
		return fn, nil
	}
	return nil, fmt.Errorf("argument %d is not an expression, but %T/%v", pos+1, arg, arg)
}

// ----- BindAll

// BindAll binds all builtins / spacial forms to the given binding.
func BindAll(bind *sxeval.Binding) error {
	err := sxeval.BindSpecials(bind,
		&QuoteS, &QuasiquoteS, // quote, quasiquote
		&UnquoteS, &UnquoteSplicingS, // unquote, unquote-splicing
		&DefVarS,          // defvar
		&DefunS, &LambdaS, // defun, lambda
		&DefDynS, &DynLambdaS, // defdyn, dyn-lambda
		&DefMacroS,       //  defmacro
		&LetS, &LetStarS, // let, let*
		&SetXS,      // set!
		&IfS,        // if
		&BeginS,     // begin
		&AndS, &OrS, // and, or
	)
	if err != nil {
		return err
	}
	err = sxeval.BindBuiltins(bind,
		&Equal,         // =
		&Identical,     // ==
		&SymbolP,       // symbol?
		&NullP,         // null?
		&Cons,          // cons
		&PairP, &ListP, // pair?, list?
		&Car, &Cdr, // car, cdr
		&Caar, &Cadr, &Cdar, &Cddr,
		&Caaar, &Caadr, &Cadar, &Caddr,
		&Cdaar, &Cdadr, &Cddar, &Cdddr,
		&Caaaar, &Caaadr, &Caadar, &Caaddr,
		&Cadaar, &Cadadr, &Caddar, &Cadddr,
		&Cdaaar, &Cdaadr, &Cdadar, &Cdaddr,
		&Cddaar, &Cddadr, &Cdddar, &Cddddr,
		&Last,            // last
		&List, &ListStar, // list, list*
		&Append,    // append
		&Reverse,   // reverse
		&Assoc,     // assoc
		&All, &Any, // all, any
		&Map,                // map
		&Apply,              // apply
		&Fold, &FoldReverse, // fold, fold-reverse
		&Not,             // not
		&NumberP,         // number?
		&Add, &Sub, &Mul, // +, -, *
		&Div, &Mod, // div, mod
		&NumLess, &NumLessEqual, // <, <=
		&NumGreater, &NumGreaterEqual, // >, >=
		&ToString, &Concat, // ->string, concat
		&Vector, &VectorP, // vector, vector?
		&VectorSetBang,        // vset!
		&List2Vector,          // list->vector
		&Length, &LengthEqual, // length, length=
		&LengthLess, &LengthGreater, // length<, length>
		&Nth,                   // nth
		&Sequence2List,         // seq->list
		&CallableP,             // callable?
		&Macroexpand0,          // macroexpand-0
		&DefinedP,              // defined?
		&CurrentBinding,        // current-binding
		&ParentBinding,         // parent-binding
		&Bindings,              // bindings
		&BoundP,                // bound?
		&BindingLookup,         // binding-lookup
		&BindingResolve,        // binding-resolve
		&Pretty,                // pp
		&Error,                 // error
		&NotBoundError,         // not-bound-error
		&ParseExpression,       // parse-expression
		&UnparseExpression,     // unparse-expression
		&RunExpression,         // run-expression
		&Eval,                  // eval
		&Compile, &Disassemble, // compile, disassemble
	)
	return err
}
