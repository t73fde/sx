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
	"slices"

	"zettelstore.de/sx.fossil"
)

// Expr are values that are computed for evaluation in an environment.
type Expr interface {
	// Rework the expressions to a possible simpler one.
	Rework(*ReworkFrame) Expr

	// Compute the expression in a frame and return the result.
	// It may have side-effects, on the given environment, or on the
	// general environment of the system.
	Compute(*Environment) (sx.Object, error)

	// IsEqual compare two expressions for deep equality.
	IsEqual(Expr) bool

	// Print the expression on the given writer.
	Print(io.Writer) (int, error)
}

// EqualExprSlice compares two `Expr` slices if they are `IsEqual`.
func EqualExprSlice(s1, s2 []Expr) bool {
	return slices.EqualFunc(s1, s2, func(e1, e2 Expr) bool { return e1.IsEqual(e2) })
}

// EqualSymbolSlice compares two `sx.Symbol` slices if they are `IsEqual`.
func EqualSymbolSlice(s1, s2 []*sx.Symbol) bool {
	return slices.EqualFunc(s1, s2, func(e1, e2 *sx.Symbol) bool { return e1.IsEqual(e2) })
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

// ObjectExpr is an Expr that results in a specific, constant sx.Object.
type ObjectExpr interface {
	Object() sx.Object
}

// NilExpr returns always Nil
var NilExpr = nilExpr{}

type nilExpr struct{}

func (nilExpr) Rework(*ReworkFrame) Expr                { return NilExpr }
func (nilExpr) Compute(*Environment) (sx.Object, error) { return sx.Nil(), nil }
func (nilExpr) IsEqual(other Expr) bool                 { return other == NilExpr }
func (nilExpr) Print(w io.Writer) (int, error)          { return io.WriteString(w, "{NIL}") }
func (nilExpr) Object() sx.Object                       { return sx.Nil() }

// ObjExpr returns the stored object.
type ObjExpr struct {
	Obj sx.Object
}

func (oe ObjExpr) Rework(rf *ReworkFrame) Expr {
	if obj := oe.Obj; sx.IsNil(obj) {
		return NilExpr.Rework(rf)
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
func (oe ObjExpr) Object() sx.Object { return oe.Obj }

// ResolveSymbolExpr resolves the given symbol in an environment and returns the value.
type ResolveSymbolExpr struct {
	Symbol *sx.Symbol
}

func (re ResolveSymbolExpr) Rework(rf *ReworkFrame) Expr {
	if obj, found := rf.ResolveConst(re.Symbol); found {
		return ObjExpr{Obj: obj}.Rework(rf)
	}
	return re
}
func (re ResolveSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	if obj, found := env.Resolve(re.Symbol); found {
		return obj, nil
	}
	return env.CallResolveSymbol(re.Symbol)
}
func (re ResolveSymbolExpr) IsEqual(other Expr) bool {
	if re == other {
		return true
	}
	if otherR, ok := other.(ResolveSymbolExpr); ok {
		return re.Symbol.IsEqual(otherR.Symbol)
	}
	return false
}
func (re ResolveSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE %v}", re.Symbol)
}

// ResolveProcSymbolExpr resolves the given symbol in an environment and returns the value.
// The symbol must resolve to a Callable, but this is not enforced by this expression.
type ResolveProcSymbolExpr struct {
	Symbol *sx.Symbol
}

func (re ResolveProcSymbolExpr) Rework(rf *ReworkFrame) Expr {
	if obj, found := rf.ResolveConst(re.Symbol); found {
		return ObjExpr{Obj: obj}.Rework(rf)
	}
	return re
}
func (re ResolveProcSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	if obj, found := env.Resolve(re.Symbol); found {
		return obj, nil
	}
	return env.CallResolveCallable(re.Symbol)
}
func (re ResolveProcSymbolExpr) IsEqual(other Expr) bool {
	if re == other {
		return true
	}
	if otherR, ok := other.(ResolveProcSymbolExpr); ok {
		return re.Symbol.IsEqual(otherR.Symbol)
	}
	return false
}
func (re ResolveProcSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE-PROC %v}", re.Symbol)
}

// CallExpr calls a procedure and returns the resulting objects.
type CallExpr struct {
	Proc Expr
	Args []Expr
}

func (ce *CallExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }
func (ce *CallExpr) Rework(rf *ReworkFrame) Expr {
	// If the ce.Proc is a builtin, rework to a BuiltinCallExpr.

	proc := ce.Proc.Rework(rf)
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		if b, isBuiltin := objExpr.Obj.(*Builtin); isBuiltin {
			bce := &BuiltinCallExpr{
				Proc: b,
				Args: ce.Args,
			}
			return bce.Rework(rf)
		}
	}
	ce.Proc = proc
	for i, arg := range ce.Args {
		ce.Args[i] = arg.Rework(rf)
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
func (ce *CallExpr) IsEqual(other Expr) bool {
	if ce == other {
		return true
	}
	if otherC, ok := other.(*CallExpr); ok && otherC != nil {
		if !ce.Proc.IsEqual(otherC.Proc) {
			return false
		}
		return EqualExprSlice(ce.Args, otherC.Args)
	}
	return false
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
	objArgs := make([]sx.Object, len(args))
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
func (bce *BuiltinCallExpr) Rework(rf *ReworkFrame) Expr {
	// Rework checks if the Builtin is pure and if all args are
	// constant sx.Object's. If this is true, it will call the builtin with
	// the args. If no error was signaled, the result object will be used
	// instead the BuiltinCallExpr. This assumes that there is no side effect
	// when the builtin is called.
	mayInline := true
	args := make([]sx.Object, len(bce.Args))
	for i, arg := range bce.Args {
		expr := arg.Rework(rf)
		if objExpr, isObjectExpr := expr.(ObjectExpr); isObjectExpr {
			args[i] = objExpr.Object()
		} else {
			mayInline = false
		}
		bce.Args[i] = expr
	}
	if !mayInline || !bce.Proc.IsPure(args) {
		return bce
	}
	result, err := rf.Call(bce.Proc, args)
	if err != nil {
		return bce
	}
	return ObjExpr{Obj: result}.Rework(rf)
}
func (bce *BuiltinCallExpr) Compute(env *Environment) (sx.Object, error) {
	subEnv := env.NewDynamicEnvironment()
	return computeCallable(subEnv, bce.Proc, bce.Args)
}
func (bce *BuiltinCallExpr) IsEqual(other Expr) bool {
	if bce == other {
		return true
	}
	if otherB, ok := other.(*BuiltinCallExpr); ok && otherB != nil {
		if !bce.Proc.IsEqual(otherB.Proc) {
			return false
		}
		return EqualExprSlice(bce.Args, otherB.Args)
	}
	return false
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
