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
func (oe ObjExpr) Improve(imp *Improver) (Expr, error) {
	if obj := oe.Obj; sx.IsNil(obj) {
		return imp.Improve(NilExpr)
	}
	return oe, nil
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
func (use UnboundSymbolExpr) Improve(imp *Improver) (Expr, error) {
	obj, depth, isConst := imp.Resolve(use.sym)
	if depth == math.MinInt {
		return use, nil
	}
	if isConst {
		return imp.Improve(ObjExpr{Obj: obj})
	}
	if depth >= 0 {
		return imp.Improve(&LookupSymbolExpr{sym: use.sym, lvl: depth})
	}
	return imp.Improve(&ResolveSymbolExpr{sym: use.sym, skip: imp.Height()})
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
func (ce *CallExpr) Improve(imp *Improver) (Expr, error) {
	// If the ce.Proc is a builtin, improve to a BuiltinCallExpr.

	proc, err := imp.Improve(ce.Proc)
	if err != nil {
		return ce, err
	}
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		if b, isBuiltin := objExpr.Obj.(*Builtin); isBuiltin {
			bce := &BuiltinCallExpr{
				Proc: b,
				Args: ce.Args,
			}
			return imp.Improve(bce)
		}
	}
	ce.Proc = proc
	for i, arg := range ce.Args {
		expr, err2 := imp.Improve(arg)
		if err2 != nil {
			return ce, err
		}
		ce.Args[i] = expr
	}
	switch len(ce.Args) {
	case 0:
		return imp.Improve(&Call0Expr{proc})
	case 1:
		return imp.Improve(&Call1Expr{proc, ce.Args[0]})
	case 2:
		return imp.Improve(&Call2Expr{proc, ce.Args[0], ce.Args[1]})
	}
	return ce, nil
}

// Compute the expression in a frame and return the result.
func (ce *CallExpr) Compute(env *Environment) (sx.Object, error) {
	proc, err := computeProc(env, ce.Proc)
	if err != nil {
		return nil, err
	}
	objArgs, err := computeArgs(env, ce.Args)
	if err != nil {
		return nil, err
	}
	return proc.Call(env, objArgs)
}

func computeProc(env *Environment, proc Expr) (Callable, error) {
	val, err := env.Execute(proc)
	if err != nil {
		return nil, err
	}
	if !sx.IsNil(val) {
		if proc, ok := val.(Callable); ok {
			return proc, nil
		}
	}
	return nil, NotCallableError{Obj: val}
}

func computeArgs(env *Environment, args []Expr) (sx.Vector, error) {
	objArgs := make(sx.Vector, len(args))
	for i, exprArg := range args {
		val, err := env.Execute(exprArg)
		if err != nil {
			return nil, err
		}
		objArgs[i] = val
	}
	return objArgs, nil
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

// Call0Expr calls a procedure with no arg and returns the resulting objects.
type Call0Expr struct{ Proc Expr }

func (c0e *Call0Expr) String() string { return fmt.Sprintf("%v", c0e.Proc) }

// Unparse the expression back into a form object.
func (c0e *Call0Expr) Unparse() sx.Object { return sx.MakeList(c0e.Proc.Unparse()) }

// Compute the expression in a frame and return the result.
func (c0e *Call0Expr) Compute(env *Environment) (sx.Object, error) {
	proc, err := computeProc(env, c0e.Proc)
	if err != nil {
		return nil, err
	}
	return proc.Call(env, nil)
}

// Print the expression on the given writer.
func (c0e *Call0Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{c0e.Proc, nil}
	return ce.doPrint(w, "{CALL-0 ")
}

// Call1Expr calls a procedure with one arg and returns the resulting objects.
type Call1Expr struct {
	Proc Expr
	Arg  Expr
}

func (c1e *Call1Expr) String() string { return fmt.Sprintf("%v %v", c1e.Proc, c1e.Arg) }

// Unparse the expression back into a form object.
func (c1e *Call1Expr) Unparse() sx.Object {
	return sx.MakeList(c1e.Proc.Unparse(), c1e.Arg.Unparse())
}

// Compute the expression in a frame and return the result.
func (c1e *Call1Expr) Compute(env *Environment) (sx.Object, error) {
	proc, err := computeProc(env, c1e.Proc)
	if err != nil {
		return nil, err
	}

	val, err := env.Execute(c1e.Arg)
	if err != nil {
		return nil, err
	}
	return proc.Call1(env, val)
}

// Print the expression on the given writer.
func (c1e *Call1Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{c1e.Proc, []Expr{c1e.Arg}}
	return ce.doPrint(w, "{CALL-1 ")
}

// Call2Expr calls a procedure with two args and returns the resulting objects.
type Call2Expr struct {
	Proc Expr
	Arg0 Expr
	Arg1 Expr
}

func (c2e *Call2Expr) String() string { return fmt.Sprintf("%v %v %v", c2e.Proc, c2e.Arg0, c2e.Arg1) }

// Unparse the expression back into a form object.
func (c2e *Call2Expr) Unparse() sx.Object {
	return sx.MakeList(c2e.Proc.Unparse(), c2e.Arg0.Unparse(), c2e.Arg1.Unparse())
}

// Compute the expression in a frame and return the result.
func (c2e *Call2Expr) Compute(env *Environment) (sx.Object, error) {
	proc, err := computeProc(env, c2e.Proc)
	if err != nil {
		return nil, err
	}
	val0, err := env.Execute(c2e.Arg0)
	if err != nil {
		return nil, err
	}
	val1, err := env.Execute(c2e.Arg1)
	if err != nil {
		return nil, err
	}
	return proc.Call2(env, val0, val1)
}

// Print the expression on the given writer.
func (c2e *Call2Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{c2e.Proc, []Expr{c2e.Arg0, c2e.Arg1}}
	return ce.doPrint(w, "{CALL-2 ")
}

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
func (bce *BuiltinCallExpr) Improve(imp *Improver) (Expr, error) {
	// Improve checks if the Builtin is pure and if all args are
	// constant sx.Object's. If this is true, it will call the builtin with
	// the args. If no error was signaled, the result object will be used
	// instead the BuiltinCallExpr. This assumes that there is no side effect
	// when the builtin is called. This is checked with `Builtin.IsPure`.
	mayInline := true
	args := make(sx.Vector, len(bce.Args))
	for i, arg := range bce.Args {
		expr, err := imp.Improve(arg)
		if err != nil {
			return bce, err
		}
		if objExpr, isConstObject := expr.(ConstObjectExpr); isConstObject {
			args[i] = objExpr.ConstObject()
		} else {
			mayInline = false
		}
		bce.Args[i] = expr
	}
	if mayInline && bce.Proc.IsPure(args) {
		if result, err := imp.Call(bce.Proc, args); err == nil {
			return imp.Improve(ObjExpr{Obj: result})
		}
	}

	switch len(bce.Args) {
	case 0:
		return imp.Improve(&BuiltinCall0Expr{bce.Proc})
	case 1:
		return imp.Improve(&BuiltinCall1Expr{bce.Proc, bce.Args[0]})
	case 2:
		return imp.Improve(&BuiltinCall2Expr{bce.Proc, bce.Args[0], bce.Args[1]})
	}

	argsFn := func() []sx.Object {
		result := make([]sx.Object, len(bce.Args))
		for i, arg := range bce.Args {
			result[i] = arg.Unparse()
		}
		return result
	}
	if err := bce.Proc.CheckCallArity(len(bce.Args), argsFn); err != nil {
		return nil, CallError{Name: bce.Proc.Name, Err: err}
	}
	return bce, nil
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCallExpr) Compute(env *Environment) (sx.Object, error) {
	objArgs, err := computeArgs(env, bce.Args)
	if err != nil {
		return nil, err
	}
	obj, err := bce.Proc.Fn(env, objArgs)
	return obj, bce.Proc.handleCallError(err)
}

// Print the expression to a io.Writer.
func (bce *BuiltinCallExpr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, bce.Args}
	return ce.doPrint(w, "{BCALL ")
}

// BuiltinCall0Expr calls a builtin with no arg and returns the resulting object.
// It is an optimization of `CallExpr.`
type BuiltinCall0Expr struct {
	Proc *Builtin
}

func (bce *BuiltinCall0Expr) String() string { return fmt.Sprintf("%v", bce.Proc) }

// Unparse the expression back into a form object.
func (bce *BuiltinCall0Expr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: nil}
	return ce.Unparse()
}

// Improve the expression into a possible simpler one.
func (bce *BuiltinCall0Expr) Improve(*Improver) (Expr, error) {
	if err := bce.checkCall0Arity(); err != nil {
		return nil, CallError{Name: bce.Proc.Name, Err: err}
	}
	return bce, nil
}

// checkCall0Arity checks the builtin to allow zero args.
func (bce *BuiltinCall0Expr) checkCall0Arity() error {
	b := bce.Proc
	minArity, maxArity := b.MinArity, b.MaxArity
	if minArity == maxArity {
		if minArity != 0 {
			return fmt.Errorf("exactly %d arguments required, but none given", minArity)
		}
	} else if maxArity < 0 {
		if 0 < minArity {
			return fmt.Errorf("at least %d arguments required, but none given", minArity)
		}
	} else if 0 < minArity {
		return fmt.Errorf("between %d and %d arguments required, but none given", minArity, maxArity)
	}
	return nil
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCall0Expr) Compute(env *Environment) (sx.Object, error) {
	obj, err := bce.Proc.Fn0(env)
	return obj, bce.Proc.handleCallError(err)
}

// Print the expression to a io.Writer.
func (bce *BuiltinCall0Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, nil}
	return ce.doPrint(w, "{BCALL-0 ")
}

// BuiltinCall1Expr calls a builtin with one arg and returns the resulting object.
// It is an optimization of `CallExpr.`
type BuiltinCall1Expr struct {
	Proc *Builtin
	Arg  Expr
}

func (bce *BuiltinCall1Expr) String() string { return fmt.Sprintf("%v %v", bce.Proc, bce.Arg) }

// Unparse the expression back into a form object.
func (bce *BuiltinCall1Expr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: []Expr{bce.Arg}}
	return ce.Unparse()
}

// Improve the expression into a possible simpler one.
func (bce *BuiltinCall1Expr) Improve(*Improver) (Expr, error) {
	if err := bce.Proc.CheckCall1Arity(func() sx.Object { return bce.Arg.Unparse() }); err != nil {
		return nil, CallError{Name: bce.Proc.Name, Err: err}
	}
	return bce, nil
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCall1Expr) Compute(env *Environment) (sx.Object, error) {
	val, err := env.Execute(bce.Arg)
	if err != nil {
		return nil, err
	}

	obj, err := bce.Proc.Fn1(env, val)
	return obj, bce.Proc.handleCallError(err)
}

// Print the expression to a io.Writer.
func (bce *BuiltinCall1Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, []Expr{bce.Arg}}
	return ce.doPrint(w, "{BCALL-1 ")
}

// BuiltinCall2Expr calls a builtin with two args and returns the resulting object.
// It is an optimization of `CallExpr.`
type BuiltinCall2Expr struct {
	Proc *Builtin
	Arg0 Expr
	Arg1 Expr
}

func (bce *BuiltinCall2Expr) String() string {
	return fmt.Sprintf("%v %v %v", bce.Proc, bce.Arg0, bce.Arg1)
}

// Unparse the expression back into a form object.
func (bce *BuiltinCall2Expr) Unparse() sx.Object {
	ce := CallExpr{Proc: ObjExpr{bce.Proc}, Args: []Expr{bce.Arg0, bce.Arg1}}
	return ce.Unparse()
}

// Improve the expression into a possible simpler one.
func (bce *BuiltinCall2Expr) Improve(*Improver) (Expr, error) {
	if err := bce.Proc.CheckCall2Arity(func() (sx.Object, sx.Object) { return bce.Arg0.Unparse(), bce.Arg1.Unparse() }); err != nil {
		return nil, CallError{Name: bce.Proc.Name, Err: err}
	}
	return bce, nil
}

// Compute the value of this expression in the given environment.
func (bce *BuiltinCall2Expr) Compute(env *Environment) (sx.Object, error) {
	val0, err := env.Execute(bce.Arg0)
	if err != nil {
		return nil, err
	}
	val1, err := env.Execute(bce.Arg1)
	if err != nil {
		return nil, err
	}

	obj, err := bce.Proc.Fn2(env, val0, val1)
	return obj, bce.Proc.handleCallError(err)
}

// Print the expression to a io.Writer.
func (bce *BuiltinCall2Expr) Print(w io.Writer) (int, error) {
	ce := CallExpr{ObjExpr{bce.Proc}, []Expr{bce.Arg0, bce.Arg1}}
	return ce.doPrint(w, "{BCALL-2 ")
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
