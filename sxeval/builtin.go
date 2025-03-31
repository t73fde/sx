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

	// The actual builtin function, with two arguments
	Fn2 func(*Environment, sx.Object, sx.Object) (sx.Object, error)

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

// Call0 the builtin function with the given environment and no arguments.
func (b *Builtin) Call0(env *Environment) (sx.Object, error) {
	// Check arity
	minArity, maxArity := b.MinArity, b.MaxArity
	if minArity == maxArity {
		if minArity != 0 {
			err := fmt.Errorf("exactly %d arguments required, but none given", minArity)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if maxArity < 0 {
		if 0 < minArity {
			err := fmt.Errorf("at least %d arguments required, but none given", minArity)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if 0 < minArity {
		err := fmt.Errorf("between %d and %d arguments required, but none given", minArity, maxArity)
		return nil, CallError{Name: b.Name, Err: err}
	}
	obj, err := b.Fn0(env)
	return b.handleCallError(obj, err)
}

// Call1 the builtin function with the given environment and one argument.
func (b *Builtin) Call1(env *Environment, arg sx.Object) (sx.Object, error) {
	// Check arity
	minArity, maxArity := b.MinArity, b.MaxArity
	if minArity == maxArity {
		if minArity != 1 {
			err := fmt.Errorf("exactly %d arguments required, but 1 given: [%v]", minArity, arg)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if maxArity < 0 {
		if 1 < minArity {
			err := fmt.Errorf("at least %d arguments required, but only 1 given: [%v]", minArity, arg)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if 1 < minArity || maxArity < 1 {
		err := fmt.Errorf("between %d and %d arguments required, but 1 given: [%v]", minArity, maxArity, arg)
		return nil, CallError{Name: b.Name, Err: err}
	}
	obj, err := b.Fn1(env, arg)
	return b.handleCallError(obj, err)
}

// Call2 the builtin function with the given environment and two arguments.
func (b *Builtin) Call2(env *Environment, arg0, arg1 sx.Object) (sx.Object, error) {
	// Check arity
	minArity, maxArity := b.MinArity, b.MaxArity
	if minArity == maxArity {
		if minArity != 2 {
			err := fmt.Errorf("exactly %d arguments required, but 2 given: [%v %v]", minArity, arg0, arg1)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if maxArity < 0 {
		if 2 < minArity {
			err := fmt.Errorf("at least %d arguments required, but only 2 given: [%v %v]", minArity, arg0, arg1)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if 2 < minArity || maxArity < 2 {
		err := fmt.Errorf("between %d and %d arguments required, but 2 given: [%v %v]", minArity, maxArity, arg0, arg1)
		return nil, CallError{Name: b.Name, Err: err}
	}
	obj, err := b.Fn2(env, arg0, arg1)
	return b.handleCallError(obj, err)
}

// Call the builtin function with the given environment and arguments.
func (b *Builtin) Call(env *Environment, args sx.Vector) (sx.Object, error) {
	// Check arity
	nargs := len(args)
	if nargs > math.MaxInt16 {
		err := fmt.Errorf("more than %d arguments are not supported, but %d given", math.MaxInt16, nargs)
		return nil, CallError{Name: b.Name, Err: err}
	}
	numArgs, minArity, maxArity := int16(nargs), b.MinArity, b.MaxArity
	if minArity == maxArity {
		if numArgs != minArity {
			var err error
			if numArgs == 0 {
				err = fmt.Errorf("exactly %d arguments required, but none given", minArity)
			} else {
				err = fmt.Errorf("exactly %d arguments required, but %d given: %v", minArity, numArgs, []sx.Object(args))
			}
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if maxArity < 0 {
		if numArgs < minArity {
			var err error
			if numArgs == 0 {
				err = fmt.Errorf("at least %d arguments required, but none given", minArity)
			} else {
				err = fmt.Errorf("at least %d arguments required, but only %d given: %v", minArity, numArgs, []sx.Object(args))
			}
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if numArgs < minArity || maxArity < numArgs {
		var err error
		if numArgs == 0 {
			err = fmt.Errorf("between %d and %d arguments required, but none given", minArity, maxArity)
		} else {
			err = fmt.Errorf("between %d and %d arguments required, but %d given: %v", minArity, maxArity, numArgs, []sx.Object(args))
		}
		return nil, CallError{Name: b.Name, Err: err}
	}
	obj, err := b.Fn(env, args)
	return b.handleCallError(obj, err)
}

func (b *Builtin) handleCallError(obj sx.Object, err error) (sx.Object, error) {
	if err != nil && !b.NoCallError {
		var callError CallError
		if !errors.As(err, &callError) {
			err = CallError{Name: b.Name, Err: err}
		}
	}
	return obj, err
}
