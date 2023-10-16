//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval

import (
	"fmt"
	"io"
	"math"

	"zettelstore.de/sx.fossil"
)

// Builtin is the type for normal builtin functions.
type Builtin struct {
	// The canonical Name of the builtin
	Name string

	// Minimum and maximum arity. If MaxArity < 0, maximum arity is unlimited
	MinArity, MaxArity int16

	// Is a call to the builtin independent of the evironment and does not produce some side effect?
	IsPure bool

	// The actual builtin function
	Fn func(*Frame, []sx.Object) (sx.Object, error)
}

// --- Builtin methods to implement sx.Object

// IsNil checks if the concrete object is nil.
func (b *Builtin) IsNil() bool { return b == nil }

// IsAtom returns true iff the object is an object that is not further decomposable.
func (b *Builtin) IsAtom() bool { return b == nil }

// IsEqual compare two objects for deep equality.
func (b *Builtin) IsEqual(other sx.Object) bool {
	if b == other {
		return true
	}
	if b == nil {
		return sx.IsNil(other)
	}
	if otherB, ok := other.(*Builtin); ok {
		return b.Name == otherB.Name
	}
	return false

}

// Repr returns the object representation.
func (b *Builtin) Repr() string { return sx.Repr(b) }

// String returns go representation.
func (b *Builtin) String() string { return b.Repr() }

func (b *Builtin) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<builtin:", b.Name, ">")
}

// --- Builtin methods to implement sxeval.Callable

// Call the builtin function with the given frame and arguments.
func (b *Builtin) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	// Check arity
	nargs := len(args)
	if nargs > math.MaxInt16 {
		err := fmt.Errorf("more than %d arguments are not supported, but %d given", math.MaxInt16, nargs)
		return nil, CallError{Name: b.Name, Err: err}
	}
	numArgs, minArity, maxArity := int16(nargs), b.MinArity, b.MaxArity
	if minArity == maxArity {
		if numArgs != minArity {
			err := fmt.Errorf("exactly %d arguments required, but %d given: %v", minArity, numArgs, args)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if maxArity < 0 {
		if numArgs < minArity {
			err := fmt.Errorf("at least %d arguments required, but only %d given: %v", minArity, numArgs, args)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if numArgs < minArity || maxArity < numArgs {
		err := fmt.Errorf("between %d and %d arguments required, but %d given: %v", minArity, maxArity, numArgs, args)
		return nil, CallError{Name: b.Name, Err: err}
	}

	obj, err := b.Fn(frame, args)
	if err == nil {
		return obj, nil
	}
	if _, ok := (err).(executeAgain); ok {
		return obj, err
	}
	if _, ok := err.(CallError); !ok {
		err = CallError{Name: b.Name, Err: err}
	}
	return obj, err
}
