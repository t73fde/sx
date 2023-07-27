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
func (b BuiltinA) Call(eng *Engine, _ sx.Environment, args []sx.Object) (sx.Object, error) {
	res, err := b(args)
	err = handleBuiltinError(eng, b, err)
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

// BuiltinEEA is the signature of builtin functions that use all information,
// engine, environment, and arguments.
type BuiltinEEA func(*Engine, sx.Environment, []sx.Object) (sx.Object, error)

func (b BuiltinEEA) IsNil() bool                    { return b == nil }
func (b BuiltinEEA) IsAtom() bool                   { return b == nil }
func (b BuiltinEEA) IsEql(other sx.Object) bool     { return sx.Object(b) == other }
func (b BuiltinEEA) IsEqual(other sx.Object) bool   { return b.IsEql(other) }
func (b BuiltinEEA) String() string                 { return b.Repr() }
func (b BuiltinEEA) Repr() string                   { return sx.Repr(b) }
func (b BuiltinEEA) Print(w io.Writer) (int, error) { return printBuiltin(w, b) }
func (b BuiltinEEA) Name(eng *Engine) string        { return eng.BuiltinName(b) }

// Call the builtin function.
func (b BuiltinEEA) Call(eng *Engine, env sx.Environment, args []sx.Object) (sx.Object, error) {
	res, err := b(eng, env, args)
	err = handleBuiltinError(eng, b, err)
	return res, err
}
