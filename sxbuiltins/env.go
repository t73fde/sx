//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

// Provides some special/builtin functions to work with environments.

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// CurrentEnv returns the current environment
func CurrentEnv(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 0, 0)
	if err != nil {
		return nil, err
	}
	return frame.Environment(), nil
}

// ParentEnv returns the parent environment of the given environment.
func ParentEnv(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	env, err := GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return env.Parent(), nil
}

// Bindings returns the bindings as a a-list of the given environment.
func Bindings(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	env, err := GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return env.Bindings(), nil
}

// AllBindings returns all bindings as a a-list of the given environment.
func AllBindings(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	env, err := GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return sxeval.AllBindings(env), nil
}

// BoundP returns true, if the given symbol is bound in the given environment.
func BoundP(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	sym, err := GetSymbol(err, args, 0)
	if err != nil {
		return nil, err
	}
	_, found := frame.Resolve(sym)
	return sx.MakeBoolean(found), nil
}

// Lookup returns the symbol's binding value in the given
// environment, or Undefined if there is no such value.
func Lookup(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 2)
	sym, err := GetSymbol(err, args, 0)
	var env sxeval.Environment
	if len(args) == 1 {
		env = frame.Environment()
	} else {
		env, err = GetEnvironment(err, args, 1)
	}
	if err != nil {
		return nil, err
	}
	if obj, found := env.Lookup(sym); found {
		return obj, nil
	}
	return sx.MakeUndefined(), nil
}

// Resolve returns the symbol's binding value in the given environment
// and all its parent environment, or Undefined if there is no such value.
func Resolve(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 2)
	sym, err := GetSymbol(err, args, 0)
	var env sxeval.Environment
	if len(args) == 1 {
		env = frame.Environment()
	} else {
		env, err = GetEnvironment(err, args, 1)
	}
	if err != nil {
		return nil, err
	}
	if obj, found := sxeval.Resolve(env, sym); found {
		return obj, nil
	}
	return sx.MakeUndefined(), nil
}
