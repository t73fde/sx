//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package sxbuiltins contains functions that help to build builtin functions.
package sxbuiltins

import (
	"fmt"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// GetSymbol returns the given argument as a symbol, and checks for errors.
func GetSymbol(args []sx.Object, pos int) (*sx.Symbol, error) {
	obj := args[pos]
	if sym, ok := sx.GetSymbol(obj); ok {
		return sym, nil
	}
	return nil, fmt.Errorf("argument %d is not a symbol, but %T/%v", pos+1, obj, obj)
}

// GetString returns the given argument as a string, and checks for errors.
func GetString(args []sx.Object, pos int) (sx.String, error) {
	obj := args[pos]
	if s, isString := sx.GetString(obj); isString {
		return s, nil
	}
	return "", fmt.Errorf("argument %d is not a string, but %T/%v", pos+1, obj, obj)
}

// GetNumber returns the given argument as a number, and checks for errors.
func GetNumber(args []sx.Object, pos int) (sx.Number, error) {
	obj := args[pos]
	if num, ok := sx.GetNumber(obj); ok {
		return num, nil
	}
	return nil, fmt.Errorf("argument %d is not a number, but %T/%v", pos+1, obj, obj)
}

// GetList returns the given argument as a list, and checks for errors.
func GetList(args []sx.Object, pos int) (*sx.Pair, error) {
	obj := args[pos]
	if sx.IsNil(obj) {
		return nil, nil
	}
	if pair, isPair := sx.GetPair(obj); isPair {
		return pair, nil
	}
	return nil, fmt.Errorf("argument %d is not a list, but %T/%v", pos+1, obj, obj)
}

// GetPair returns the given argument as a non-nil list, and checks for errors.
func GetPair(args []sx.Object, pos int) (*sx.Pair, error) {
	obj := args[pos]
	if !sx.IsNil(obj) {
		if pair, isPair := sx.GetPair(obj); isPair {
			return pair, nil
		}
	}
	return nil, fmt.Errorf("argument %d is not a pair, but %T/%v", pos+1, obj, obj)
}

// GetEnvironment returns the given argument as an environment, and checks for errors.
func GetEnvironment(args []sx.Object, pos int) (sxeval.Environment, error) {
	obj := args[pos]
	if env, ok := sxeval.GetEnvironment(obj); ok {
		return env, nil
	}
	return nil, fmt.Errorf("argument %d is not an environment, but %T/%v", pos+1, obj, obj)
}

// GetCallable returns the given argument as a callable, and checks for errors.
func GetCallable(args []sx.Object, pos int) (sxeval.Callable, error) {
	obj := args[pos]
	if fn, ok := sxeval.GetCallable(obj); ok {
		return fn, nil
	}
	return nil, fmt.Errorf("argument %d is not a function, but %T/%v", pos+1, obj, obj)
}

// ParseExprSeq parses a sequence of expressions.
func ParseExprSeq(pf *sxeval.ParseFrame, args *sx.Pair) ([]sxeval.Expr, error) {
	if args == nil {
		return nil, nil
	}
	var front []sxeval.Expr
	for node := args; ; {
		ex, err := pf.Parse(node.Car())
		if err != nil {
			return nil, err
		}
		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			return append(front, ex), nil
		}
		front = append(front, ex)
		if next, isPair := sx.GetPair(cdr); isPair {
			node = next
			continue
		}
		ex, err = pf.Parse(cdr)
		if err != nil {
			return nil, err
		}
		return append(front, ex), nil
	}
}
