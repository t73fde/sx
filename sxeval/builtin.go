//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval

import (
	"errors"
	"fmt"
	"math"

	"t73f.de/r/sx"
)

// Builtin is the type for normal predefined functions.
type Builtin struct {
	// The canonical Name of the builtin
	Name string

	// Minimum and maximum arity. If MaxArity < 0, maximum arity is unlimited
	MinArity, MaxArity int16

	// Test builtin to be independent of the environment and does not produce some side effect
	TestPure func(sx.Vector) bool

	// The actual builtin function, with no argument
	Fn0 func(*Environment, *Binding) (sx.Object, error)

	// The actual builtin function, with one argument
	Fn1 func(*Environment, sx.Object, *Binding) (sx.Object, error)

	// The actual builtin function, with any number of arguments
	Fn func(*Environment, sx.Vector, *Binding) (sx.Object, error)

	// Do not add a CallError
	NoCallError bool
}

// AssertPure is a TestPure function that alsways returns true.
func AssertPure(sx.Vector) bool { return true }

// Bind the builtin to a given environment.
func (b *Builtin) Bind(bind *Binding) error {
	return bind.Bind(sx.MakeSymbol(b.Name), b)
}

// BindBuiltins will bind many builtins to an environment.
func BindBuiltins(bind *Binding, bs ...*Builtin) error {
	for _, b := range bs {
		if err := b.Bind(bind); err != nil {
			return err
		}
	}
	return nil
}

// GetBuiltin returns the object as a builtin, if possible.
func GetBuiltin(obj sx.Object) (*Builtin, bool) { b, ok := obj.(*Builtin); return b, ok }

// --- Builtin methods to implement sx.Object

// IsNil checks if the concrete object is nil.
func (b *Builtin) IsNil() bool { return b == nil }

// IsAtom returns true iff the object is an object that is not further decomposable.
func (b *Builtin) IsAtom() bool { return b == nil }

// IsEqual compare two objects for deep equality.
func (b *Builtin) IsEqual(other sx.Object) bool { return b == other }

// String returns the string representation.
func (b *Builtin) String() string { return "#<builtin:" + b.Name + ">" }

// GoString returns the go string representation.
func (b *Builtin) GoString() string { return b.String() }

// --- Builtin methods to implement sxeval.Callable

// IsPure returns true if builtin is a pure function.
func (b *Builtin) IsPure(objs sx.Vector) bool {
	if testPure := b.TestPure; testPure != nil && len(objs) <= math.MaxInt16 {
		numargs, minArity, maxArity := int16(len(objs)), b.MinArity, b.MaxArity
		if minArity == maxArity {
			if numargs != minArity {
				return false
			}
		} else if maxArity < 0 {
			if numargs < minArity {
				return false
			}
		} else if numargs < minArity || maxArity < numargs {
			return false
		}
		return testPure(objs)
	}
	return false
}

// ExecuteCall the builtin function with the given environment and number of arguments.
func (b *Builtin) ExecuteCall(env *Environment, numargs int, bind *Binding) (obj sx.Object, err error) {
	if err = b.checkCallArity(numargs, func() []sx.Object { return env.Args(numargs) }); err != nil {
		return sx.Nil(), b.handleCallError(err)
	}
	switch numargs {
	case 0:
		if obj, err = b.Fn0(env, bind); err == nil {
			return obj, nil
		}
	case 1:
		if obj, err = b.Fn1(env, env.Top(), bind); err == nil {
			return obj, nil
		}
	default:
		if obj, err = b.Fn(env, env.Args(numargs), bind); err == nil {
			return obj, nil
		}
	}
	return obj, b.handleCallError(err)
}

// checkCallArity check the builtin function to match allowed number of args.
func (b *Builtin) checkCallArity(nargs int, argsFn func() []sx.Object) error {
	if nargs > math.MaxInt16 {
		return fmt.Errorf("more than %d arguments are not supported, but %d given", math.MaxInt16, nargs)
	}
	numArgs, minArity, maxArity := int16(nargs), b.MinArity, b.MaxArity
	if minArity == maxArity {
		if numArgs != minArity {
			if nargs == 0 {
				return fmt.Errorf("exactly %d arguments required, but none given", minArity)
			}
			return fmt.Errorf("exactly %d arguments required, but %d given: %v", minArity, numArgs, argsFn())
		}
	} else if maxArity < 0 {
		if numArgs < minArity {
			if nargs == 0 {
				return fmt.Errorf("at least %d arguments required, but none given", minArity)
			}
			return fmt.Errorf("at least %d arguments required, but only %d given: %v", minArity, numArgs, argsFn())
		}
	} else if numArgs < minArity || maxArity < numArgs {
		if nargs == 0 {
			return fmt.Errorf("between %d and %d arguments required, but none given", minArity, maxArity)
		}
		return fmt.Errorf("between %d and %d arguments required, but %d given: %v", minArity, maxArity, numArgs, argsFn())
	}
	return nil
}

func (b *Builtin) handleCallError(err error) error {
	if err != nil && !b.NoCallError {
		var callError CallError
		if !errors.As(err, &callError) {
			err = CallError{Name: b.Name, Err: err}
		}
	}
	return err
}
