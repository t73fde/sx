//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package env provides some special/builtin functions to work with environments.
package env

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
)

// Env returns the current environment
func Env(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 0, 0)
	if err != nil {
		return nil, err
	}
	return frame.Environment(), nil
}

// ParentEnv returns the parent environment of the given environment.
func ParentEnv(args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 1, 1)
	env, err := sxbuiltins.GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return env.Parent(), nil
}

// Bindings returns the bindings as a a-list of the given environment.
func Bindings(args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 1, 1)
	env, err := sxbuiltins.GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return env.Bindings(), nil
}

// AllBindings returns all bindings as a a-list of the given environment.
func AllBindings(args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 1, 1)
	env, err := sxbuiltins.GetEnvironment(err, args, 0)
	if err != nil {
		return nil, err
	}
	return sxeval.AllBindings(env), nil
}

// BoundP returns true, if the given symbol is bound in the given environment.
func BoundP(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 1, 1)
	sym, err := sxbuiltins.GetSymbol(err, args, 0)
	if err != nil {
		return nil, err
	}
	_, found := frame.Resolve(sym)
	return sx.MakeBoolean(found), nil
}
