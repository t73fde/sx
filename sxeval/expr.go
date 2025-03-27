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
	// Unparse the expression as an sx.Object
	Unparse() sx.Object

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

// Unparse the expression back into a form object.
func (nilExpr) Unparse() sx.Object { return sx.Nil() }

// Compute the expression in a frame and return the result.
func (nilExpr) Compute(*Environment) (sx.Object, error) { return sx.Nil(), nil }

// Print the expression on the given writer.
func (nilExpr) Print(w io.Writer) (int, error) { return io.WriteString(w, "{NIL}") }
func (nilExpr) ConstObject() sx.Object         { return sx.Nil() }

// ObjExpr returns the stored object.
type ObjExpr struct {
	Obj sx.Object
}

// Unparse the expression back into a form object.
func (oe ObjExpr) Unparse() sx.Object { return oe.Obj }

// Improve the expression into a possible simpler one.
func (oe ObjExpr) Improve(re *Improver) Expr {
	if obj := oe.Obj; sx.IsNil(obj) {
		return re.Improve(NilExpr)
	}
	return oe
}

// Compute the expression in a frame and return the result.
func (oe ObjExpr) Compute(*Environment) (sx.Object, error) { return oe.Obj, nil }

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

// SymbolExpr is the common interface of Expr that handles symbols.
type SymbolExpr interface {
	Expr
	GetSymbol() *sx.Symbol
}

// UnboundSymbolExpr resolves the given symbol in an environment and returns its value.
type UnboundSymbolExpr struct{ sym *sx.Symbol }

// GetSymbol returns the symbol that is current not known to be bound to a value.
func (use UnboundSymbolExpr) GetSymbol() *sx.Symbol { return use.sym }

// Unparse the expression back into a form object.
func (use UnboundSymbolExpr) Unparse() sx.Object { return use.sym }

// Improve the expression into a possible simpler one.
func (use UnboundSymbolExpr) Improve(re *Improver) Expr {
	obj, depth, isConst := re.Resolve(use.sym)
	if depth == math.MinInt {
		return use
	}
	if isConst {
		return re.Improve(ObjExpr{Obj: obj})
	}
	if depth >= 0 {
		return re.Improve(&LookupSymbolExpr{sym: use.sym, lvl: depth})
	}
	return re.Improve(&ResolveSymbolExpr{sym: use.sym, skip: re.Height()})
}

// Compute the expression in a frame and return the result.
func (use UnboundSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	return env.ResolveUnboundWithError(use.sym)
}

// Print the expression on the given writer.
func (use UnboundSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{UNBOUND %v}", use.sym)
}

// ResolveSymbolExpr is a special `UnboundSymbolExpr` that must be resolved in
// the base environment. Traversal through all nested lexical bindings is not
// needed.
type ResolveSymbolExpr struct {
	sym  *sx.Symbol
	skip int
}

// GetSymbol returns the symbol that later must be resolved.
func (rse ResolveSymbolExpr) GetSymbol() *sx.Symbol { return rse.sym }

// Unparse the expression back into a form object.
func (rse ResolveSymbolExpr) Unparse() sx.Object { return rse.sym }

// Compute the expression in a frame and return the result.
func (rse ResolveSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	return env.ResolveNWithError(rse.sym, rse.skip)
}

// Print the expression on the given writer.
func (rse ResolveSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE/%d %v}", rse.skip, rse.sym)
}

// LookupSymbolExpr is a special UnboundSymbolExpr that gives an indication
// about the nesting level of `Binding`s, where the symbol will be bound.
type LookupSymbolExpr struct {
	sym *sx.Symbol
	lvl int
}

// GetSymbol returns the symbol that must later be looked up.
func (lse *LookupSymbolExpr) GetSymbol() *sx.Symbol { return lse.sym }

// GetLevel returns the nesting level to later start a look up.
func (lse *LookupSymbolExpr) GetLevel() int { return lse.lvl }

// Unparse the expression back into a form object.
func (lse *LookupSymbolExpr) Unparse() sx.Object { return lse.sym }

// Compute the expression in a frame and return the result.
func (lse *LookupSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	return env.LookupNWithError(lse.sym, lse.lvl)
}

// Print the expression on the given writer.
func (lse LookupSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{LOOKUP/%d %v}", lse.lvl, lse.sym)
}

// --- CallExpr ---------------------------------------------------------------

// CallExpr calls a procedure and returns the resulting objects.
type CallExpr struct {
	Proc Expr
	Args []Expr
}

func (ce *CallExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }

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
func (ce *CallExpr) Improve(re *Improver) Expr {
	// If the ce.Proc is a builtin, improve to a BuiltinCallExpr.

	proc := re.Improve(ce.Proc)
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		if b, isBuiltin := objExpr.Obj.(*Builtin); isBuiltin {
			bce := &BuiltinCallExpr{
				Proc: b,
				Args: ce.Args,
			}
			return re.Improve(bce)
		}
	}
	ce.Proc = proc
	for i, arg := range ce.Args {
		ce.Args[i] = re.Improve(arg)
	}
	return ce
}

// Compute the expression in a frame and return the result.
func (ce *CallExpr) Compute(env *Environment) (sx.Object, error) {
	val, err := env.Execute(ce.Proc)
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

// Print the expression on the given writer.
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
	switch numargs := len(args); numargs {
	case 0:
		return proc.Call0(env)
	case 1:
		arg, err := env.Execute(args[0])
		if err != nil {
			return nil, err
		}
		return proc.Call1(env, arg)
	case 2:
		arg0, err := env.Execute(args[0])
		if err != nil {
			return nil, err
		}
		arg1, err := env.Execute(args[1])
		if err != nil {
			return nil, err
		}
		return proc.Call2(env, arg0, arg1)
	default:
		objArgs := make(sx.Vector, numargs)
		for i, exprArg := range args {
			val, err := env.Execute(exprArg)
			if err != nil {
				return val, err
			}
			objArgs[i] = val
		}
		return proc.Call(env, objArgs)
	}
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

// Unparse the expression back into a form object.
func (bce *BuiltinCallExpr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: bce.Args}
	return ce.Unparse()
}

// Improve the expression into a possible simpler one.
func (bce *BuiltinCallExpr) Improve(re *Improver) Expr {
	// Improve checks if the Builtin is pure and if all args are
	// constant sx.Object's. If this is true, it will call the builtin with
	// the args. If no error was signaled, the result object will be used
	// instead the BuiltinCallExpr. This assumes that there is no side effect
	// when the builtin is called. This is checked with `Builtin.IsPure`.
	mayInline := true
	args := make(sx.Vector, len(bce.Args))
	for i, arg := range bce.Args {
		expr := re.Improve(arg)
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
	return re.Improve(ObjExpr{Obj: result})
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCallExpr) Compute(env *Environment) (sx.Object, error) {
	return computeCallable(env, bce.Proc, bce.Args)
}

// Print the expression to a io.Writer.
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
