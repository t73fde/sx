//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
)

// Expr are values that are computed for evaluation in an environment.
type Expr interface {
	// Unparse the expression as an sx.Object
	Unparse() sx.Object

	// Rework the expressions to a possible simpler one.
	Rework(*ReworkEnvironment) Expr

	// Compute the expression in a frame and return the result.
	// It may have side-effects, on the given environment, or on the
	// general environment of the system.
	Compute(*Environment) (sx.Object, error)

	// Print the expression on the given writer.
	Print(io.Writer) (int, error)
}

// PrintExprs is a helper method to implement Expr.Print.
func PrintExprs(w io.Writer, exprs []Expr) (int, error) {
	length := 0
	for _, expr := range exprs {
		l, err := io.WriteString(w, " ")
		length += l
		if err != nil {
			return length, err
		}
		l, err = expr.Print(w)
		length += l
		if err != nil {
			return length, err
		}
	}
	return length, nil
}

// ConstObjectExpr is an Expr that results in a specific, constant sx.Object.
type ConstObjectExpr interface {
	ConstObject() sx.Object
}

// GetConstExpr returns the Expr as a ConstObjectExpr, if possible.
func GetConstExpr(expr Expr) (ConstObjectExpr, bool) {
	if coe, isCoe := expr.(ConstObjectExpr); isCoe {
		return coe, true
	}
	return nil, false
}

// NilExpr returns always Nil
var NilExpr = nilExpr{}

type nilExpr struct{}

func (nilExpr) Unparse() sx.Object                      { return sx.Nil() }
func (nilExpr) Rework(*ReworkEnvironment) Expr          { return NilExpr }
func (nilExpr) Compute(*Environment) (sx.Object, error) { return sx.Nil(), nil }
func (nilExpr) Print(w io.Writer) (int, error)          { return io.WriteString(w, "{NIL}") }
func (nilExpr) ConstObject() sx.Object                  { return sx.Nil() }

// ObjExpr returns the stored object.
type ObjExpr struct {
	Obj sx.Object
}

func (oe ObjExpr) Unparse() sx.Object { return oe.Obj }

func (oe ObjExpr) Rework(re *ReworkEnvironment) Expr {
	if obj := oe.Obj; sx.IsNil(obj) {
		return NilExpr.Rework(re)
	}
	return oe
}

func (oe ObjExpr) Compute(*Environment) (sx.Object, error) { return oe.Obj, nil }

func (oe ObjExpr) IsEqual(other Expr) bool {
	if oe == other {
		return true
	}
	if otherO, ok := other.(ObjExpr); ok {
		return oe.Obj.IsEqual(otherO.Obj)
	}
	return false
}

func (oe ObjExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{OBJ ")
	if err != nil {
		return length, err
	}
	l, err := sx.Print(w, oe.Obj)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
func (oe ObjExpr) ConstObject() sx.Object { return oe.Obj }

// ResolveSymbolExpr resolves the given symbol in an environment and returns the value.
type ResolveSymbolExpr struct {
	Symbol *sx.Symbol
}

func (rse ResolveSymbolExpr) Unparse() sx.Object { return rse.Symbol }

func (rse ResolveSymbolExpr) Rework(re *ReworkEnvironment) Expr {
	if obj, found := re.ResolveConst(rse.Symbol); found {
		return ObjExpr{Obj: obj}.Rework(re)
	}
	return rse
}

func (rse ResolveSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	if obj, found := env.Resolve(rse.Symbol); found {
		return obj, nil
	}
	return nil, env.MakeNotBoundError(rse.Symbol)
}

func (rse ResolveSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE %v}", rse.Symbol)
}

// CallExpr calls a procedure and returns the resulting objects.
type CallExpr struct {
	Proc Expr
	Args []Expr
}

func (ce *CallExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }

func (ce *CallExpr) Unparse() sx.Object {
	lst := make(sx.Vector, len(ce.Args)+1)
	lst[0] = ce.Proc.Unparse()
	for i, arg := range ce.Args {
		lst[i+1] = arg.Unparse()
	}
	return sx.MakeList(lst...)
}

func (ce *CallExpr) Rework(re *ReworkEnvironment) Expr {
	// If the ce.Proc is a builtin, rework to a BuiltinCallExpr.

	proc := ce.Proc.Rework(re)
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		if b, isBuiltin := objExpr.Obj.(*Builtin); isBuiltin {
			bce := &BuiltinCallExpr{
				Proc: b,
				Args: ce.Args,
			}
			return bce.Rework(re)
		}
	}
	ce.Proc = proc
	for i, arg := range ce.Args {
		ce.Args[i] = arg.Rework(re)
	}
	return ce
}

func (ce *CallExpr) Compute(env *Environment) (sx.Object, error) {
	subEnv := env.NewDynamicEnvironment()
	val, err := subEnv.Execute(ce.Proc)
	if err != nil {
		return nil, err
	}
	if sx.IsNil(val) {
		return nil, NotCallableError{Obj: val}
	}
	proc, ok := val.(Callable)
	if !ok {
		return nil, NotCallableError{Obj: val}
	}

	return computeCallable(env, proc, ce.Args)
}

func (ce *CallExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{CALL ")
	if err != nil {
		return length, err
	}
	l, err := ce.Proc.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = PrintExprs(w, ce.Args)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

func computeCallable(env *Environment, proc Callable, args []Expr) (sx.Object, error) {
	if len(args) == 0 {
		return proc.Call(env, nil)
	}
	objArgs := make(sx.Vector, len(args))
	for i, exprArg := range args {
		subEnv := env.NewDynamicEnvironment()
		val, err := subEnv.Execute(exprArg)
		if err != nil {
			return val, err
		}
		objArgs[i] = val
	}
	return proc.Call(env, objArgs)
}

// NotCallableError signals that a value cannot be called when it must be called.
type NotCallableError struct {
	Obj sx.Object
}

func (e NotCallableError) Error() string {
	return fmt.Sprintf("not callable: %T/%v", e.Obj, e.Obj)
}
func (e NotCallableError) String() string { return e.Error() }

// BuiltinCallExpr calls a builtin and returns the resulting object.
// It is an optimization of `CallExpr.`
type BuiltinCallExpr struct {
	Proc *Builtin
	Args []Expr
}

func (bce *BuiltinCallExpr) String() string { return fmt.Sprintf("%v %v", bce.Proc, bce.Args) }

func (bce *BuiltinCallExpr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: bce.Args}
	return ce.Unparse()
}

func (bce *BuiltinCallExpr) Rework(re *ReworkEnvironment) Expr {
	// Rework checks if the Builtin is pure and if all args are
	// constant sx.Object's. If this is true, it will call the builtin with
	// the args. If no error was signaled, the result object will be used
	// instead the BuiltinCallExpr. This assumes that there is no side effect
	// when the builtin is called. This is checked with `Builtin.IsPure`.
	mayInline := true
	args := make(sx.Vector, len(bce.Args))
	for i, arg := range bce.Args {
		expr := arg.Rework(re)
		if objExpr, isConstObject := expr.(ConstObjectExpr); isConstObject {
			args[i] = objExpr.ConstObject()
		} else {
			mayInline = false
		}
		bce.Args[i] = expr
	}
	if !mayInline || !bce.Proc.IsPure(args) {
		return bce
	}
	result, err := re.Call(bce.Proc, args)
	if err != nil {
		return bce
	}
	return ObjExpr{Obj: result}.Rework(re)
}

func (bce *BuiltinCallExpr) Compute(env *Environment) (sx.Object, error) {
	subEnv := env.NewDynamicEnvironment()
	return computeCallable(subEnv, bce.Proc, bce.Args)
}

func (bce *BuiltinCallExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{BCALL ")
	if err != nil {
		return length, err
	}
	l, err := sx.Print(w, bce.Proc)
	length += l
	if err != nil {
		return length, err
	}
	l, err = PrintExprs(w, bce.Args)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
