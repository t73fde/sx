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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// GetSymbol returns the given argument as a symbol, and checks for errors.
func GetSymbol(args sx.Vector, pos int) (sx.Symbol, error) {
	obj := args[pos]
	if sym, ok := sx.GetSymbol(obj); ok {
		return sym, nil
	}
	return sx.MakeSymbol(""), fmt.Errorf("argument %d is not a symbol, but %T/%v", pos+1, obj, obj)
}

// GetString returns the given argument as a string, and checks for errors.
func GetString(args sx.Vector, pos int) (sx.String, error) {
	obj := args[pos]
	if s, isString := sx.GetString(obj); isString {
		return s, nil
	}
	return "", fmt.Errorf("argument %d is not a string, but %T/%v", pos+1, obj, obj)
}

// GetNumber returns the given argument as a number, and checks for errors.
func GetNumber(args sx.Vector, pos int) (sx.Number, error) {
	obj := args[pos]
	if num, ok := sx.GetNumber(obj); ok {
		return num, nil
	}
	return nil, fmt.Errorf("argument %d is not a number, but %T/%v", pos+1, obj, obj)
}

// GetList returns the given argument as a list, and checks for errors.
func GetList(args sx.Vector, pos int) (*sx.Pair, error) {
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
func GetPair(args sx.Vector, pos int) (*sx.Pair, error) {
	obj := args[pos]
	if !sx.IsNil(obj) {
		if pair, isPair := sx.GetPair(obj); isPair {
			return pair, nil
		}
	}
	return nil, fmt.Errorf("argument %d is not a pair, but %T/%v", pos+1, obj, obj)
}

// GetVector returns the given argument as a vector, and checks for errors.
func GetVector(args sx.Vector, pos int) (sx.Vector, error) {
	obj := args[pos]
	if v, ok := sx.GetVector(obj); ok {
		return v, nil
	}
	return nil, fmt.Errorf("argument %d is not a vector, but %T/%v", pos+1, obj, obj)
}

// GetSequence returns the given argument as a sequence, and checks for errors.
func GetSequence(args sx.Vector, pos int) (sx.Sequence, error) {
	obj := args[pos]
	if seq, ok := sx.GetSequence(obj); ok {
		if sx.IsNil(seq) {
			return sx.Nil(), nil
		}
		return seq, nil
	}
	return nil, fmt.Errorf("argument %d is not a sequence, but %T/%v", pos+1, obj, obj)
}

// GetBinding returns the given argument as a binding, and checks for errors.
func GetBinding(args sx.Vector, pos int) (*sxeval.Binding, error) {
	obj := args[pos]
	if bind, ok := sxeval.GetBinding(obj); ok {
		return bind, nil
	}
	return nil, fmt.Errorf("argument %d is not a binding, but %T/%v", pos+1, obj, obj)
}

// GetCallable returns the given argument as a callable, and checks for errors.
func GetCallable(args sx.Vector, pos int) (sxeval.Callable, error) {
	obj := args[pos]
	if fn, ok := sxeval.GetCallable(obj); ok {
		return fn, nil
	}
	return nil, fmt.Errorf("argument %d is not a function, but %T/%v", pos+1, obj, obj)
}
