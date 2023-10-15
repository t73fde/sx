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

// CurrentEnvOld returns the current environment
func CurrentEnvOld(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 0, 0)
	if err != nil {
		return nil, err
	}
	return frame.Environment(), nil
}

// ParentEnvOld returns the parent environment of the given environment.
func ParentEnvOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	env, err := GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	if parent := env.Parent(); parent != nil {
		return parent, nil
	}
	return sx.MakeUndefined(), nil
}

// EnvBindingsOld returns the bindings as a a-list of the given environment.
func EnvBindingsOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	env, err := GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return env.Bindings(), nil
}

// BoundPold returns true, if the given symbol is bound in the given environment.
func BoundPold(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	sym, err := GetSymbol(err, args, 0)
	if err != nil {
		return nil, err
	}
	_, found := frame.Resolve(sym)
	return sx.MakeBoolean(found), nil
}

// EnvLookupOld returns the symbol's binding value in the given
// environment, or Undefined if there is no such value.
func EnvLookupOld(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
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

// EnvResolveOld returns the symbol's binding value in the given environment
// and all its parent environment, or Undefined if there is no such value.
func EnvResolveOld(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
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
