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
	Fn0 func(*Environment) (sx.Object, error)

	// The actual builtin function, with one argument
	Fn1 func(*Environment, sx.Object) (sx.Object, error)

	// The actual builtin function, with any number of arguments
	Fn func(*Environment, sx.Vector) (sx.Object, error)

	// Do not add a CallError
	NoCallError bool
}

// AssertPure is a TestPure function that alsways returns true.
func AssertPure(sx.Vector) bool { return true }

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
	if testPure := b.TestPure; testPure != nil {
		return testPure(objs)
	}
	return false
}

// Call the builtin function with the given environment and arguments.
func (b *Builtin) Call(env *Environment, args sx.Vector) (sx.Object, error) {
	if err := b.CheckCallArity(len(args), func() []sx.Object { return args }); err != nil {
		return sx.Nil(), b.handleCallError(err)
	}
	switch len(args) {
	case 0:
		obj, err := b.Fn0(env)
		return obj, b.handleCallError(err)
	case 1:
		obj, err := b.Fn1(env, args[0])
		return obj, b.handleCallError(err)
	default:
		obj, err := b.Fn(env, args)
		return obj, b.handleCallError(err)
	}
}

// CheckCallArity check the builtin function to match allowed number of args.
func (b *Builtin) CheckCallArity(nargs int, argsFn func() []sx.Object) error {
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
