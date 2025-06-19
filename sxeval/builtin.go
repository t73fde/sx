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
	"io"

	"t73f.de/r/sx"
)

// Builtin is the type of normal predefined functions.
type Builtin struct {
	// The canonical Name of the builtin.
	Name string

	// Minimum and maximum arity. If MaxArity < 0, maximum arity is unlimited.
	MinArity, MaxArity int

	// Test builtin to be independent of the environment and does not produce some side effect.
	TestPure func(sx.Vector) bool

	// The actual builtin function, with no argument.
	Fn0 func(*Environment, *Frame) (sx.Object, error)

	// The actual builtin function, with one argument.
	Fn1 func(*Environment, sx.Object, *Frame) (sx.Object, error)

	// The actual builtin function, with any number of arguments.
	//
	// The function is alowed to read each single element of the vector, but it
	// is not allowed to store and process the vector itself. The vector is
	// essentially only a slice on top of the evaluation stack, which may
	// change its elements.
	Fn func(*Environment, sx.Vector, *Frame) (sx.Object, error)

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
	if testPure := b.TestPure; testPure != nil {
		numargs, minArity, maxArity := len(objs), b.MinArity, b.MaxArity
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

// ExecuteCall the builtin function with the given environment and arguments.
func (b *Builtin) ExecuteCall(env *Environment, args sx.Vector, frame *Frame) (obj sx.Object, err error) {
	if err = b.checkCallArity(len(args), func() []sx.Object { return args }); err != nil {
		return nil, b.handleCallError(err)
	}
	switch len(args) {
	case 0:
		obj, err = b.Fn0(env, frame)
	case 1:
		obj, err = b.Fn1(env, args[0], frame)
	default:
		obj, err = b.Fn(env, args, frame)
	}
	if err == nil {
		return obj, nil
	}
	return nil, b.handleCallError(err)
}

// checkCallArity check the builtin function to match allowed number of args.
func (b *Builtin) checkCallArity(nargs int, argsFn func() []sx.Object) error {
	if minArity, maxArity := b.MinArity, b.MaxArity; minArity == maxArity {
		if nargs != minArity {
			if nargs == 0 {
				return fmt.Errorf("exactly %d arguments required, but none given", minArity)
			}
			return fmt.Errorf("exactly %d arguments required, but %d given: %v", minArity, nargs, argsFn())
		}
	} else if maxArity < 0 {
		if nargs < minArity {
			if nargs == 0 {
				return fmt.Errorf("at least %d arguments required, but none given", minArity)
			}
			return fmt.Errorf("at least %d arguments required, but only %d given: %v", minArity, nargs, argsFn())
		}
	} else if nargs < minArity || maxArity < nargs {
		if nargs == 0 {
			return fmt.Errorf("between %d and %d arguments required, but none given", minArity, maxArity)
		}
		return fmt.Errorf("between %d and %d arguments required, but %d given: %v", minArity, maxArity, nargs, argsFn())
	}
	return nil
}

func (b *Builtin) handleCallError(err error) error {
	if !b.NoCallError {
		var callError CallError
		if !errors.As(err, &callError) {
			err = CallError{Name: b.Name, Err: err}
		}
	}
	return err
}

// ----- builtinCallExpr, builtinCall0Expr, BuiltinCall1Expr

// builtinCallExpr calls a builtin and returns the resulting object.
// It is an optimization of `CallExpr.`
type builtinCallExpr struct {
	Proc *Builtin
	Args []Expr
}

func (bce *builtinCallExpr) String() string { return fmt.Sprintf("%v %v", bce.Proc, bce.Args) }

// IsPure signals an expression that has no side effects.
func (bce *builtinCallExpr) IsPure() bool {
	args := make(sx.Vector, len(bce.Args))
	for i, expr := range bce.Args {
		if !expr.IsPure() {
			return false
		}
		args[i] = sx.MakeUndefined()
	}
	return bce.Proc.IsPure(args)
}

// Unparse the expression back into a form object.
func (bce *builtinCallExpr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: bce.Args}
	return ce.Unparse()
}

// Improve the expression into a possible simpler one.
func (bce *builtinCallExpr) Improve(imp *Improver) (Expr, error) {
	argsFn := func() []sx.Object {
		result := make([]sx.Object, len(bce.Args))
		for i, arg := range bce.Args {
			if val, err := imp.env.Execute(arg, imp.frame); err == nil {
				result[i] = val
			} else {
				result[i] = arg.Unparse()
			}
		}
		return result
	}
	if err := bce.Proc.checkCallArity(len(bce.Args), argsFn); err != nil {
		return nil, bce.Proc.handleCallError(err)
	}

	switch len(bce.Args) {
	case 0:
		return imp.Improve(&builtinCall0Expr{bce.Proc})
	case 1:
		return imp.Improve(&BuiltinCall1Expr{bce.Proc, bce.Args[0]})
	}
	return bce, nil
}

// Compute the value of this expression in the given environment.
func (bce *builtinCallExpr) Compute(env *Environment, frame *Frame) (sx.Object, error) {
	args := bce.Args
	if err := computeArgs(env, args, frame); err != nil {
		return nil, err
	}
	obj, err := bce.Proc.Fn(env, env.Args(len(args)), frame)
	env.Kill(len(args))
	if err != nil {
		return nil, bce.Proc.handleCallError(err)
	}
	return obj, nil
}

// Print the expression to a io.Writer.
func (bce *builtinCallExpr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, bce.Args}
	return ce.doPrint(w, "{BCALL ")
}

// builtinCall0Expr calls a builtin with no arg and returns the resulting object.
// It is an optimization of `CallExpr.`
type builtinCall0Expr struct {
	Proc *Builtin
}

func (bce *builtinCall0Expr) String() string { return fmt.Sprintf("%v", bce.Proc) }

// IsPure signals an expression that has no side effects.
func (bce *builtinCall0Expr) IsPure() bool { return bce.Proc.IsPure(nil) }

// Unparse the expression back into a form object.
func (bce *builtinCall0Expr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: nil}
	return ce.Unparse()
}

// Compute the value of this expression in the given environment.
func (bce *builtinCall0Expr) Compute(env *Environment, frame *Frame) (sx.Object, error) {
	obj, err := bce.Proc.Fn0(env, frame)
	if err != nil {
		return nil, bce.Proc.handleCallError(err)
	}
	return obj, nil
}

// Print the expression to a io.Writer.
func (bce *builtinCall0Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, nil}
	return ce.doPrint(w, "{BCALL-0 ")
}

// BuiltinCall1Expr calls a builtin with one arg and returns the resulting object.
// It is an optimization of `CallExpr`. Do not create it outside this package.
// It is public, because it is used for conditional expressions, especially to
// detect a (not ...) expression.
type BuiltinCall1Expr struct {
	Proc *Builtin
	Arg  Expr
}

func (bce *BuiltinCall1Expr) String() string { return fmt.Sprintf("%v %v", bce.Proc, bce.Arg) }

// IsPure signals an expression that has no side effects.
func (bce *BuiltinCall1Expr) IsPure() bool {
	return bce.Arg.IsPure() && bce.Proc.IsPure(sx.Vector{sx.MakeUndefined()})
}

// Unparse the expression back into a form object.
func (bce *BuiltinCall1Expr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: []Expr{bce.Arg}}
	return ce.Unparse()
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCall1Expr) Compute(env *Environment, frame *Frame) (sx.Object, error) {
	val, err := env.Execute(bce.Arg, frame)
	if err != nil {
		return nil, err
	}
	obj, err := bce.Proc.Fn1(env, val, frame)
	if err != nil {
		return nil, bce.Proc.handleCallError(err)
	}
	return obj, nil
}

// Print the expression to a io.Writer.
func (bce *BuiltinCall1Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, []Expr{bce.Arg}}
	return ce.doPrint(w, "{BCALL-1 ")
}
