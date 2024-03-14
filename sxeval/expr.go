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

	"zettelstore.de/sx.fossil"
)

// Expr are values that are computed for evaluation in an environment.
type Expr interface {
	// Unparse the expression as an sx.Object
	Unparse() sx.Object

	// Improve the expressions to a possible simpler one.
	Improve(*ReworkEnvironment) Expr

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
func (nilExpr) Improve(*ReworkEnvironment) Expr         { return NilExpr }
func (nilExpr) Compute(*Environment) (sx.Object, error) { return sx.Nil(), nil }
func (nilExpr) Print(w io.Writer) (int, error)          { return io.WriteString(w, "{NIL}") }
func (nilExpr) ConstObject() sx.Object                  { return sx.Nil() }

// ObjExpr returns the stored object.
type ObjExpr struct {
	Obj sx.Object
}

func (oe ObjExpr) Unparse() sx.Object { return oe.Obj }

func (oe ObjExpr) Improve(re *ReworkEnvironment) Expr {
	if obj := oe.Obj; sx.IsNil(obj) {
		return re.Rework(NilExpr)
	}
	return oe
}

func (oe ObjExpr) Compute(*Environment) (sx.Object, error) { return oe.Obj, nil }

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

// --- SymbolExpr -------------------------------------------------------------

// SymbolExpr is the common interface of Expr that handles symbols.
type SymbolExpr interface {
	Expr
	GetSymbol() *sx.Symbol
}

// UnboundSymbolExpr resolves the given symbol in an environment and returns its value.
type UnboundSymbolExpr struct{ sym *sx.Symbol }

func (use UnboundSymbolExpr) GetSymbol() *sx.Symbol { return use.sym }
func (use UnboundSymbolExpr) Unparse() sx.Object    { return use.sym }
func (use UnboundSymbolExpr) Improve(re *ReworkEnvironment) Expr {
	obj, depth, isConst := re.Resolve(use.sym)
	if depth == math.MinInt {
		return use
	}
	if isConst {
		return re.Rework(ObjExpr{Obj: obj})
	}
	if depth >= 0 {
		return re.Rework(&LookupSymbolExpr{sym: use.sym, lvl: depth})
	}
	return re.Rework(&ResolveSymbolExpr{sym: use.sym, skip: re.Height()})
}

func (use UnboundSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	return env.ResolveUnboundWithError(use.sym)
}

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

func (rse ResolveSymbolExpr) GetSymbol() *sx.Symbol              { return rse.sym }
func (rse ResolveSymbolExpr) Unparse() sx.Object                 { return rse.sym }
func (rse ResolveSymbolExpr) Improve(re *ReworkEnvironment) Expr { return rse }
func (use ResolveSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	return env.ResolveNWithError(use.sym, use.skip)
}

func (use ResolveSymbolExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE/%d %v}", use.skip, use.sym)
}

// LookupSymbolExpr is a special UnboundSymbolExpr that gives an indication
// about the nesting level of `Binding`s, where the symbol will be bound.
type LookupSymbolExpr struct {
	sym *sx.Symbol
	lvl int
}

func (lse *LookupSymbolExpr) GetSymbol() *sx.Symbol { return lse.sym }
func (lse *LookupSymbolExpr) GetLevel() int         { return lse.lvl }

func (lse *LookupSymbolExpr) Unparse() sx.Object              { return lse.sym }
func (lse *LookupSymbolExpr) Improve(*ReworkEnvironment) Expr { return lse }

func (lse *LookupSymbolExpr) Compute(env *Environment) (sx.Object, error) {
	return env.LookupNWithError(lse.sym, lse.lvl)
}

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

func (ce *CallExpr) Unparse() sx.Object {
	lst := make(sx.Vector, len(ce.Args)+1)
	lst[0] = ce.Proc.Unparse()
	for i, arg := range ce.Args {
		lst[i+1] = arg.Unparse()
	}
	return sx.MakeList(lst...)
}

func (ce *CallExpr) Improve(re *ReworkEnvironment) Expr {
	// If the ce.Proc is a builtin, rework to a BuiltinCallExpr.

	proc := re.Rework(ce.Proc)
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		if b, isBuiltin := objExpr.Obj.(*Builtin); isBuiltin {
			bce := &BuiltinCallExpr{
				Proc: b,
				Args: ce.Args,
			}
			return re.Rework(bce)
		}
	}
	ce.Proc = proc
	for i, arg := range ce.Args {
		ce.Args[i] = re.Rework(arg)
	}
	return ce
}

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

func (bce *BuiltinCallExpr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: bce.Args}
	return ce.Unparse()
}

func (bce *BuiltinCallExpr) Improve(re *ReworkEnvironment) Expr {
	// Improve checks if the Builtin is pure and if all args are
	// constant sx.Object's. If this is true, it will call the builtin with
	// the args. If no error was signaled, the result object will be used
	// instead the BuiltinCallExpr. This assumes that there is no side effect
	// when the builtin is called. This is checked with `Builtin.IsPure`.
	mayInline := true
	args := make(sx.Vector, len(bce.Args))
	for i, arg := range bce.Args {
		expr := re.Rework(arg)
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
	return re.Rework(ObjExpr{Obj: result})
}

func (bce *BuiltinCallExpr) Compute(env *Environment) (sx.Object, error) {
	return computeCallable(env, bce.Proc, bce.Args)
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

// --- ExprObj ---------------------------------------------------------------

// ExprObj encapsulates an Expr in an sx.Object.
type ExprObj struct {
	expr Expr
}

// MakeExprObj creates an ExprObj from an Expr.
func MakeExprObj(expr Expr) *ExprObj {
	return &ExprObj{expr}
}

func (eo *ExprObj) IsNil() bool { return eo == nil }
func (*ExprObj) IsAtom() bool   { return false }

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

func (eo *ExprObj) String() string {
	var sb strings.Builder
	sb.WriteString("#<")
	if _, err := eo.expr.Print(&sb); err != nil {
		return err.Error()
	}
	sb.WriteByte('>')
	return sb.String()
}

func (eo *ExprObj) GoString() string {
	var sb strings.Builder
	if _, err := eo.expr.Print(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

func (eo *ExprObj) GetExpr() Expr {
	if eo == nil {
		return nil
	}
	return eo.expr
}

// GetExprObj returns the object as a expression object, if possible.
func GetExprObj(obj sx.Object) (*ExprObj, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	eo, ok := obj.(*ExprObj)
	return eo, ok
}
