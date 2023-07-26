//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package builtins contains functions that help to build builtin functions.
package builtins

import (
	"fmt"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// CheckArgs validates the number of arguments given.
func CheckArgs(args []sxpf.Object, minArgs, maxArgs int) error {
	numArgs := len(args)
	if minArgs == maxArgs {
		if numArgs != minArgs {
			return fmt.Errorf("exactly %d arguments required, but %d given: %v", minArgs, numArgs, args)
		}
	} else if minArgs > maxArgs {
		if numArgs < minArgs {
			return fmt.Errorf("at least %d arguments required, but only %d given: %v", minArgs, numArgs, args)
		}
	} else {
		if numArgs < minArgs || maxArgs < numArgs {
			return fmt.Errorf("between %d and %d arguments required, but %d given: %v", minArgs, maxArgs, numArgs, args)
		}
	}
	return nil
}

// getObject returns the given argument as an object, but checks for errors.
func getObject(err error, args []sxpf.Object, pos int) (sxpf.Object, error) {
	if err != nil {
		return nil, err
	}
	if l := len(args); l <= pos {
		return nil, fmt.Errorf("need at least %d argument, but only %d given: %v", pos+1, l+1, args)
	}
	return args[pos], nil
}

// GetSymbol returns the given argument as a symbol, and checks for errors.
func GetSymbol(err error, args []sxpf.Object, pos int) (*sxpf.Symbol, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if sym, ok := sxpf.GetSymbol(obj); ok {
			return sym, nil
		}
		err = fmt.Errorf("argument %d is not a symbol, but %T/%v", pos+1, obj, obj)
	}
	return nil, err
}

// GetString returns the given argument as a string, and checks for errors.
func GetString(err error, args []sxpf.Object, pos int) (sxpf.String, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if s, isString := sxpf.GetString(obj); isString {
			return s, nil
		}
		err = fmt.Errorf("argument %d is not a string, but %T/%v", pos+1, obj, obj)
	}
	return "", err
}

// GetNumber returns the given argument as a number, and checks for errors.
func GetNumber(err error, args []sxpf.Object, pos int) (sxpf.Number, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if num, ok := sxpf.GetNumber(obj); ok {
			return num, nil
		}
		err = fmt.Errorf("argument %d is not a number, but %T/%v", pos+1, obj, obj)
	}
	return nil, err
}

// GetList returns the given argument as a list, and checks for errors.
func GetList(err error, args []sxpf.Object, pos int) (*sxpf.Pair, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if sxpf.IsNil(obj) {
			return nil, nil
		}
		if pair, isPair := sxpf.GetPair(obj); isPair {
			return pair, nil
		}
		err = fmt.Errorf("argument %d is not a list, but %T/%v", pos+1, obj, obj)
	}
	return nil, err
}

// GetPair returns the given argument as a non-nil list, and checks for errors.
func GetPair(err error, args []sxpf.Object, pos int) (*sxpf.Pair, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if !sxpf.IsNil(obj) {
			if pair, isPair := sxpf.GetPair(obj); isPair {
				return pair, nil
			}
		}
		err = fmt.Errorf("argument %d is not a pair, but %T/%v", pos+1, obj, obj)
	}
	return nil, err
}

// GetEnvironment returns the given argument as an environment, and checks for errors.
func GetEnvironment(err error, args []sxpf.Object, pos int) (sxpf.Environment, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if env, ok := sxpf.GetEnvironment(obj); ok {
			return env, nil
		}
		err = fmt.Errorf("argument %d is not an environment, but %T/%v", pos+1, obj, obj)
	}
	return nil, err
}

// GetCallable returns the given argument as a callable, and checks for errors.
func GetCallable(err error, args []sxpf.Object, pos int) (eval.Callable, error) {
	obj, err := getObject(err, args, pos)
	if err == nil {
		if fn, ok := eval.GetCallable(obj); ok {
			return fn, nil
		}
		err = fmt.Errorf("argument %d is not a function, but %T/%v", pos+1, obj, obj)
	}
	return nil, err
}

// ParseExprSeq parses a sequence of expressions.
func ParseExprSeq(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (front []eval.Expr, last eval.Expr, _ error) {
	if args == nil {
		return nil, nil, nil
	}
	for node := args; ; {
		ex, err := eng.Parse(env, node.Car())
		if err != nil {
			return nil, nil, err
		}
		cdr := node.Cdr()
		if sxpf.IsNil(cdr) {
			return front, ex, nil
		}
		front = append(front, ex)
		if next, isPair := sxpf.GetPair(cdr); isPair {
			node = next
			continue
		}
		ex, err = eng.Parse(env, cdr)
		if err != nil {
			return nil, nil, err
		}
		return front, ex, nil
	}
}
