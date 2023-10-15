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
	"reflect"

	"zettelstore.de/sx.fossil"
)

// Builtin is the type for normal builtin functions.
type Builtin struct {
	// The canonical Name of the builtin
	Name string

	// Minimum and maximum arity. If MaxArity < MinArity, maximum arity is unlimited
	MinArity, MaxArity int

	// Will a call to the builtin produce some side effect?
	HasSideEffect bool

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
	numArgs, minArity, maxArity := len(args), b.MinArity, b.MaxArity
	if minArity == maxArity {
		if numArgs != minArity {
			err := fmt.Errorf("exactly %d arguments required, but %d given: %v", minArity, numArgs, args)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else if minArity > maxArity {
		if numArgs < minArity {
			err := fmt.Errorf("at least %d arguments required, but only %d given: %v", minArity, numArgs, args)
			return nil, CallError{Name: b.Name, Err: err}
		}
	} else {
		if numArgs < minArity || maxArity < numArgs {
			err := fmt.Errorf("between %d and %d arguments required, but %d given: %v", minArity, maxArity, numArgs, args)
			return nil, CallError{Name: b.Name, Err: err}
		}
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

// --- The following code is deprecated.

// BuiltinOld is a callable with a name
type BuiltinOld interface {
	sx.Object
	Callable

	Name(*Engine) string
}

// BuiltinAold is the signature of all normal builtin functions.
//
// These functions are not allowed to have a side effect. Otherwise you should
// us BuiltinFAold instead.
type BuiltinAold func([]sx.Object) (sx.Object, error)

func (b BuiltinAold) IsNil() bool  { return b == nil }
func (b BuiltinAold) IsAtom() bool { return b == nil }
func (b BuiltinAold) IsEqual(other sx.Object) bool {
	return reflect.ValueOf(b).Pointer() == reflect.ValueOf(other).Pointer()
}
func (b BuiltinAold) String() string                 { return b.Repr() }
func (b BuiltinAold) Repr() string                   { return sx.Repr(b) }
func (b BuiltinAold) Print(w io.Writer) (int, error) { return printBuiltin(w, b) }
func (b BuiltinAold) Name(eng *Engine) string        { return eng.BuiltinName(b) }

// Call the builtin function.
func (b BuiltinAold) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	res, err := b(args)
	var engine *Engine
	if frame != nil {
		engine = frame.engine
	}
	err = handleBuiltinError(engine, b, err)
	return res, err
}

func printBuiltin(w io.Writer, b BuiltinOld) (int, error) {
	return sx.WriteStrings(w, "#<builtin:", b.Name(nil), ">")
}

func handleBuiltinError(eng *Engine, b BuiltinOld, err error) error {
	if err != nil {
		if _, ok := (err).(executeAgain); ok {
			return err
		}
		if _, ok := err.(CallError); !ok {
			if name := b.Name(eng); name != "" {
				err = CallError{Name: b.Name(eng), Err: err}
			}
		}
	}
	return err
}

// BuiltinFAold is the signature of builtin functions that use all information,
// frame (i.e. engine, environment), and arguments.
type BuiltinFAold func(*Frame, []sx.Object) (sx.Object, error)

func (b BuiltinFAold) IsNil() bool  { return b == nil }
func (b BuiltinFAold) IsAtom() bool { return b == nil }
func (b BuiltinFAold) IsEqual(other sx.Object) bool {
	return reflect.ValueOf(b).Pointer() == reflect.ValueOf(other).Pointer()
}
func (b BuiltinFAold) String() string                 { return b.Repr() }
func (b BuiltinFAold) Repr() string                   { return sx.Repr(b) }
func (b BuiltinFAold) Print(w io.Writer) (int, error) { return printBuiltin(w, b) }
func (b BuiltinFAold) Name(eng *Engine) string        { return eng.BuiltinName(b) }

// Call the builtin function.
func (b BuiltinFAold) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	res, err := b(frame, args)
	var engine *Engine
	if frame != nil {
		engine = frame.engine
	}
	err = handleBuiltinError(engine, b, err)
	return res, err
}
