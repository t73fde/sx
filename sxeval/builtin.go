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
	"io"

	"zettelstore.de/sx.fossil"
)

// Builtin is a callable with a name
type Builtin interface {
	sx.Object
	Callable

	Name(*Engine) string
}

// BuiltinA is the signature of all normal builtin functions.
//
// These functions are not allowed to have a side effect. Otherwise you should
// us BuiltinEEA instead.
type BuiltinA func([]sx.Object) (sx.Object, error)

func (b BuiltinA) IsNil() bool                    { return b == nil }
func (b BuiltinA) IsAtom() bool                   { return b == nil }
func (b BuiltinA) IsEql(other sx.Object) bool     { return sx.Object(b) == other }
func (b BuiltinA) IsEqual(other sx.Object) bool   { return b.IsEql(other) }
func (b BuiltinA) String() string                 { return b.Repr() }
func (b BuiltinA) Repr() string                   { return sx.Repr(b) }
func (b BuiltinA) Print(w io.Writer) (int, error) { return printBuiltin(w, b) }
func (b BuiltinA) Name(eng *Engine) string        { return eng.BuiltinName(b) }

// Call the builtin function.
func (b BuiltinA) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	res, err := b(args)
	var engine *Engine
	if frame != nil {
		engine = frame.engine
	}
	err = handleBuiltinError(engine, b, err)
	return res, err
}

func printBuiltin(w io.Writer, b Builtin) (int, error) {
	return sx.WriteStrings(w, "#<builtin:", b.Name(nil), ">")
}

func handleBuiltinError(eng *Engine, b Builtin, err error) error {
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

// BuiltinFA is the signature of builtin functions that use all information,
// frame (i.e. engine, environment), and arguments.
type BuiltinFA func(*Frame, []sx.Object) (sx.Object, error)

func (b BuiltinFA) IsNil() bool                    { return b == nil }
func (b BuiltinFA) IsAtom() bool                   { return b == nil }
func (b BuiltinFA) IsEql(other sx.Object) bool     { return sx.Object(b) == other }
func (b BuiltinFA) IsEqual(other sx.Object) bool   { return b.IsEql(other) }
func (b BuiltinFA) String() string                 { return b.Repr() }
func (b BuiltinFA) Repr() string                   { return sx.Repr(b) }
func (b BuiltinFA) Print(w io.Writer) (int, error) { return printBuiltin(w, b) }
func (b BuiltinFA) Name(eng *Engine) string        { return eng.BuiltinName(b) }

// Call the builtin function.
func (b BuiltinFA) Call(frame *Frame, args []sx.Object) (sx.Object, error) {
	res, err := b(frame, args)
	var engine *Engine
	if frame != nil {
		engine = frame.engine
	}
	err = handleBuiltinError(engine, b, err)
	return res, err
}
