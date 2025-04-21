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
	"math"
	"strings"

	"t73f.de/r/sx"
)

// Expr are values that are computed for evaluation in an environment.
type Expr interface {
	// IsPure signals an expression with no side effects.
	IsPure() bool

	// Unparse the expression as an sx.Object
	Unparse() sx.Object

	// Compute the expression in a frame and return the result.
	// It may have side-effects, on the given environment, or on the
	// general environment of the system.
	Compute(*Environment, *Binding) (sx.Object, error)

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

// IsPure signals an expression that has no side effects.
func (nilExpr) IsPure() bool { return true }

// Unparse the expression back into a form object.
func (nilExpr) Unparse() sx.Object { return sx.Nil() }

// Compute the expression in a frame and return the result.
func (nilExpr) Compute(*Environment, *Binding) (sx.Object, error) { return sx.Nil(), nil }

// Print the expression on the given writer.
func (nilExpr) Print(w io.Writer) (int, error) { return io.WriteString(w, "{NIL}") }
func (nilExpr) ConstObject() sx.Object         { return sx.Nil() }

// ObjExpr returns the stored object.
type ObjExpr struct {
	Obj sx.Object
}

// IsPure signals an expression that has no side effects.
func (ObjExpr) IsPure() bool { return true }

// Unparse the expression back into a form object.
func (oe ObjExpr) Unparse() sx.Object { return oe.Obj }

// Improve the expression into a possible simpler one.
func (oe ObjExpr) Improve(imp *Improver) (Expr, error) {
	if obj := oe.Obj; sx.IsNil(obj) {
		return imp.Improve(NilExpr)
	}
	return oe, nil
}

// Compute the expression in a frame and return the result.
func (oe ObjExpr) Compute(*Environment, *Binding) (sx.Object, error) { return oe.Obj, nil }

// Print the expression on the given writer.
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

// ConstObject returns the object stored by this expression. It must be treated
// as a constant value.
func (oe ObjExpr) ConstObject() sx.Object { return oe.Obj }

// --- SymbolExpr -------------------------------------------------------------

// UnboundSymbolExpr resolves the given symbol in an environment and returns its value.
type UnboundSymbolExpr struct{ sym *sx.Symbol }

// IsPure signals an expression that has no side effects.
func (UnboundSymbolExpr) IsPure() bool { return true }

// Unparse the expression back into a form object.
func (use UnboundSymbolExpr) Unparse() sx.Object { return use.sym }

// Improve the expression into a possible simpler one.
func (use UnboundSymbolExpr) Improve(imp *Improver) (Expr, error) {
	obj, depth, isConst := imp.Resolve(use.sym)
	if depth == math.MinInt {
		return use, nil
	}
	if isConst {
		return imp.Improve(ObjExpr{Obj: obj})
	}
	if depth >= 0 {
		return imp.Improve(&lookupSymbolExpr{sym: use.sym, lvl: depth})
	}
	return imp.Improve(&resolveSymbolExpr{sym: use.sym, skip: imp.Height()})
}

// Compute the expression in a frame and return the result.
func (use UnboundSymbolExpr) Compute(_ *Environment, bind *Binding) (sx.Object, error) {
	if obj, found := bind.Resolve(use.sym); found {
		return obj, nil
	}
	return nil, bind.MakeNotBoundError(use.sym)

}

// Print the expression on the given writer.
func (use UnboundSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{UNBOUND %v}", use.sym)
}

// resolveSymbolExpr is a special `UnboundSymbolExpr` that must be resolved in
// the base environment. Traversal through all nested lexical bindings is not
// needed.
type resolveSymbolExpr struct {
	sym  *sx.Symbol
	skip int
}

// IsPure signals an expression that has no side effects.
func (resolveSymbolExpr) IsPure() bool { return true }

// Unparse the expression back into a form object.
func (rse resolveSymbolExpr) Unparse() sx.Object { return rse.sym }

// Compute the expression in a frame and return the result.
func (rse resolveSymbolExpr) Compute(_ *Environment, bind *Binding) (sx.Object, error) {
	if obj, found := bind.ResolveN(rse.sym, rse.skip); found {
		return obj, nil
	}
	return nil, bind.MakeNotBoundError(rse.sym)
}

// Print the expression on the given writer.
func (rse resolveSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE/%d %v}", rse.skip, rse.sym)
}

// lookupSymbolExpr is a special UnboundSymbolExpr that gives an indication
// about the nesting level of `Binding`s, where the symbol will be bound.
type lookupSymbolExpr struct {
	sym *sx.Symbol
	lvl int
}

// IsPure signals an expression that has no side effects.
func (*lookupSymbolExpr) IsPure() bool { return true }

// Unparse the expression back into a form object.
func (lse *lookupSymbolExpr) Unparse() sx.Object { return lse.sym }

// Compute the expression in a frame and return the result.
func (lse *lookupSymbolExpr) Compute(_ *Environment, bind *Binding) (sx.Object, error) {
	if obj, found := bind.LookupN(lse.sym, lse.lvl); found {
		return obj, nil
	}
	return nil, bind.MakeNotBoundError(lse.sym)
}

// Print the expression on the given writer.
func (lse lookupSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{LOOKUP/%d %v}", lse.lvl, lse.sym)
}

// --- CallExpr ---------------------------------------------------------------

// CallExpr calls a procedure and returns the resulting objects.
type CallExpr struct {
	Proc Expr
	Args []Expr
}

func (ce *CallExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }

// IsPure signals an expression that has no side effects.
func (*CallExpr) IsPure() bool { return false }

// Unparse the expression back into a form object.
func (ce *CallExpr) Unparse() sx.Object {
	lst := make(sx.Vector, len(ce.Args)+1)
	lst[0] = ce.Proc.Unparse()
	for i, arg := range ce.Args {
		lst[i+1] = arg.Unparse()
	}
	return sx.MakeList(lst...)
}

// Improve the expression into a possible simpler one.
func (ce *CallExpr) Improve(imp *Improver) (Expr, error) {
	proc, err := imp.Improve(ce.Proc)
	if err != nil {
		return ce, err
	}
	if err = imp.ImproveSlice(ce.Args); err != nil {
		return ce, err
	}
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		// If call can be folded into a constant value, use that value.
		if c, isCallable := objExpr.Obj.(Callable); isCallable {
			if foldExpr, foldErr := imp.ImproveFoldCall(c, ce.Args); foldErr == nil && foldExpr != nil {
				return foldExpr, nil
			}
		}

		// If the ce.Proc is a builtin, improve to a BuiltinCallExpr.
		if b, isBuiltin := objExpr.Obj.(*Builtin); isBuiltin {
			bce := &builtinCallExpr{
				Proc: b,
				Args: ce.Args,
			}
			return imp.Improve(bce)
		}
	}

	ce.Proc = proc
	return ce, nil
}

// Compute the expression in a frame and return the result.
func (ce *CallExpr) Compute(env *Environment, bind *Binding) (sx.Object, error) {
	val, err := env.Execute(ce.Proc, bind)
	if err != nil {
		return nil, err
	}
	if !sx.IsNil(val) {
		if proc, isCallable := val.(Callable); isCallable {
			args := ce.Args
			if err = computeArgs(env, args, bind); err != nil {
				return nil, err
			}
			obj, err2 := proc.Call(env, env.Args(len(args)), bind)
			env.Kill(len(args))
			return obj, err2
		}
	}
	return nil, NotCallableError{Obj: val}
}

func computeArgs(env *Environment, args []Expr, bind *Binding) error {
	for _, exprArg := range args {
		val, err := env.Execute(exprArg, bind)
		if err != nil {
			return err
		}
		env.Push(val)
	}
	return nil
}

// Print the expression on the given writer.
func (ce *CallExpr) Print(w io.Writer) (int, error) {
	return ce.doPrint(w, "{CALL ")
}

func (ce *CallExpr) doPrint(w io.Writer, init string) (int, error) {
	length, err := io.WriteString(w, init)
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

// NotCallableError signals that a value cannot be called when it must be called.
type NotCallableError struct {
	Obj sx.Object
}

func (e NotCallableError) Error() string {
	return fmt.Sprintf("not callable: %T/%v", e.Obj, e.Obj)
}
func (e NotCallableError) String() string { return e.Error() }

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
		args[i] = sx.Nil()
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
			result[i] = arg.Unparse()
		}
		return result
	}
	if err := bce.Proc.checkCallArity(len(bce.Args), argsFn); err != nil {
		return nil, CallError{Name: bce.Proc.Name, Err: err}
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
func (bce *builtinCallExpr) Compute(env *Environment, bind *Binding) (sx.Object, error) {
	args := bce.Args
	if err := computeArgs(env, args, bind); err != nil {
		return nil, err
	}
	proc := bce.Proc
	obj, err := proc.Fn(env, env.Args(len(args)), bind)
	env.Kill(len(args))
	return obj, proc.handleCallError(err)
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
func (bce *builtinCall0Expr) Compute(env *Environment, bind *Binding) (sx.Object, error) {
	proc := bce.Proc
	obj, err := proc.Fn0(env, bind)
	return obj, proc.handleCallError(err)
}

// Print the expression to a io.Writer.
func (bce *builtinCall0Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, nil}
	return ce.doPrint(w, "{BCALL-0 ")
}

// BuiltinCall1Expr calls a builtin with one arg and returns the resulting object.
// It is an optimization of `CallExpr`. Do not create it outside this package.
type BuiltinCall1Expr struct {
	Proc *Builtin
	Arg  Expr
}

func (bce *BuiltinCall1Expr) String() string { return fmt.Sprintf("%v %v", bce.Proc, bce.Arg) }

// IsPure signals an expression that has no side effects.
func (bce *BuiltinCall1Expr) IsPure() bool {
	return bce.Arg.IsPure() && bce.Proc.IsPure(sx.Vector{sx.Nil()})
}

// Unparse the expression back into a form object.
func (bce *BuiltinCall1Expr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: []Expr{bce.Arg}}
	return ce.Unparse()
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCall1Expr) Compute(env *Environment, bind *Binding) (sx.Object, error) {
	val, err := env.Execute(bce.Arg, bind)
	if err != nil {
		return nil, err
	}
	proc := bce.Proc
	obj, err := proc.Fn1(env, val, bind)
	return obj, proc.handleCallError(err)
}

// Print the expression to a io.Writer.
func (bce *BuiltinCall1Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, []Expr{bce.Arg}}
	return ce.doPrint(w, "{BCALL-1 ")
}

// --- ExprObj ---------------------------------------------------------------

// ExprObj encapsulates an Expr in an sx.Object.
type ExprObj struct {
	expr Expr
}

// MakeExprObj creates an ExprObj from an Expr.
func MakeExprObj(expr Expr) *ExprObj {
	return &ExprObj{expr}
}

// IsNil returns true if the object must be treated like a sx.Nil() object.
func (eo *ExprObj) IsNil() bool { return eo == nil }

// IsAtom returns true if the object is atomic.
func (*ExprObj) IsAtom() bool { return false }

// IsEqual returns true if the other object has the same content.
func (eo *ExprObj) IsEqual(other sx.Object) bool {
	if eo == nil {
		return sx.IsNil(other)
	}
	if sx.IsNil(other) {
		return false
	}
	otherEo, isEO := other.(*ExprObj)
	return isEO && (eo == otherEo || eo.expr == otherEo.expr)
}

// String returns a string representation.
func (eo *ExprObj) String() string {
	var sb strings.Builder
	sb.WriteString("#<")
	if _, err := eo.expr.Print(&sb); err != nil {
		return err.Error()
	}
	sb.WriteByte('>')
	return sb.String()
}

// GoString returns a string representation to be used in Go code.
func (eo *ExprObj) GoString() string {
	var sb strings.Builder
	if _, err := eo.expr.Print(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

// GetExpr returns the expression of this object.
func (eo *ExprObj) GetExpr() Expr {
	if eo == nil {
		return nil
	}
	return eo.expr
}

// GetExprObj returns the object as an expression object, if possible.
func GetExprObj(obj sx.Object) (*ExprObj, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	eo, ok := obj.(*ExprObj)
	return eo, ok
}
