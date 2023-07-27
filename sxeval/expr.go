//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
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

	"zettelstore.de/sx.fossil"
)

// Expr are values that are computed for evaluation in an environment.
type Expr interface {
	// Compute the expression in an environment and return the result.
	// It may have side-effects, on the given environment, or on the
	// general environment of the system.
	Compute(*Engine, sx.Environment) (sx.Object, error)

	// Print the expression on the given writer.
	Print(io.Writer) (int, error)

	// Rework the expressions to a possible simpler one.
	Rework(*ReworkOptions, sx.Environment) Expr
}

// ReworkOptions controls the behaviour of Expr.Rework.
type ReworkOptions struct {
	// The environment where resolve should try to resolve a symbol.
	ResolveEnv sx.Environment
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

// PrintFrontLast is a helper method to implement Expr.Print.
func PrintFrontLast(w io.Writer, front []Expr, last Expr) (int, error) {
	length, err := PrintExprs(w, front)
	if err != nil {
		return length, err
	}
	l, err := last.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

// ObjectExpr is an Expr that results in a specific sx.Object.
type ObjectExpr interface {
	Object() sx.Object
}

// NilExpr returns always Nil
var NilExpr = nilExpr{}

type nilExpr struct{}

func (nilExpr) Compute(*Engine, sx.Environment) (sx.Object, error) { return sx.Nil(), nil }
func (nilExpr) Print(w io.Writer) (int, error)                     { return io.WriteString(w, "{NIL}") }
func (nilExpr) Rework(*ReworkOptions, sx.Environment) Expr         { return NilExpr }
func (nilExpr) Object() sx.Object                                  { return sx.Nil() }

// FalseExpr returns always False
var FalseExpr = falseExpr{}

type falseExpr struct{}

func (falseExpr) Compute(*Engine, sx.Environment) (sx.Object, error) { return sx.False, nil }
func (falseExpr) Print(w io.Writer) (int, error)                     { return io.WriteString(w, "{FALSE}") }
func (falseExpr) Rework(*ReworkOptions, sx.Environment) Expr         { return FalseExpr }
func (falseExpr) Object() sx.Object                                  { return sx.False }

// TrueExpr returns always True
var TrueExpr = trueExpr{}

type trueExpr struct{}

func (trueExpr) Compute(*Engine, sx.Environment) (sx.Object, error) { return sx.True, nil }
func (trueExpr) Print(w io.Writer) (int, error)                     { return io.WriteString(w, "{TRUE}") }
func (trueExpr) Rework(*ReworkOptions, sx.Environment) Expr         { return TrueExpr }
func (trueExpr) Object() sx.Object                                  { return sx.True }

// ObjExpr returns the stored object.
type ObjExpr struct {
	Obj sx.Object
}

func (oe ObjExpr) Compute(*Engine, sx.Environment) (sx.Object, error) { return oe.Obj, nil }
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
func (oe ObjExpr) Rework(ro *ReworkOptions, env sx.Environment) Expr {
	if obj := oe.Obj; sx.IsNil(obj) {
		return NilExpr.Rework(ro, env)
	} else if obj == sx.False {
		return FalseExpr.Rework(ro, env)
	} else if obj == sx.True {
		return TrueExpr.Rework(ro, env)
	}
	return oe
}
func (oe ObjExpr) Object() sx.Object { return oe.Obj }

// ResolveExpr resolves the given symbol in an environment and returns the value.
type ResolveExpr struct {
	Symbol *sx.Symbol
}

func (re ResolveExpr) Compute(_ *Engine, env sx.Environment) (sx.Object, error) {
	if obj, found := sx.Resolve(env, re.Symbol); found {
		return obj, nil
	}
	return nil, NotBoundError{Env: env, Sym: re.Symbol}
}
func (re ResolveExpr) Print(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "{RESOLVE %v}", re.Symbol)
}
func (re ResolveExpr) Rework(ro *ReworkOptions, env sx.Environment) Expr {
	if reEnv := ro.ResolveEnv; reEnv != nil {
		if obj, found := sx.Resolve(reEnv, re.Symbol); found {
			return ObjExpr{Obj: obj}.Rework(ro, env)
		}
	}
	return re
}

// NotBoundError signals that a symbol was not found in an environment.
type NotBoundError struct {
	Env sx.Environment
	Sym *sx.Symbol
}

func (e NotBoundError) Error() string {
	return fmt.Sprintf("symbol %q not bound in environment %q", e.Sym.Name(), e.Env.String())
}

// CallExpr calls a procedure and returns the resulting objects.
type CallExpr struct {
	Proc Expr
	Args []Expr
}

func (ce *CallExpr) String() string { return fmt.Sprintf("%v %v", ce.Proc, ce.Args) }
func (ce *CallExpr) Compute(eng *Engine, env sx.Environment) (sx.Object, error) {
	val, err := eng.Execute(env, ce.Proc)
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

	return computeCallable(eng, env, proc, ce.Args)
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
func (ce *CallExpr) Rework(ro *ReworkOptions, env sx.Environment) Expr {
	// If the ce.Proc is a builtin, rework to a BuiltinCallExpr.

	proc := ce.Proc.Rework(ro, env)
	if objExpr, isObjExpr := proc.(ObjExpr); isObjExpr {
		if bi, isBuiltin := objExpr.Obj.(Builtin); isBuiltin {
			bce := &BuiltinCallExpr{
				Proc: bi,
				Args: ce.Args,
			}
			return bce.Rework(ro, env)
		}
	}
	ce.Proc = proc
	for i, arg := range ce.Args {
		ce.Args[i] = arg.Rework(ro, env)
	}
	return ce
}

func computeCallable(eng *Engine, env sx.Environment, proc Callable, args []Expr) (sx.Object, error) {
	if len(args) == 0 {
		return proc.Call(eng, env, nil)
	}
	objArgs := make([]sx.Object, len(args))
	for i, exprArg := range args {
		val, err := eng.Execute(env, exprArg)
		if err != nil {
			return val, err
		}
		objArgs[i] = val
	}
	return proc.Call(eng, env, objArgs)
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
	Proc Builtin
	Args []Expr
}

func (bce *BuiltinCallExpr) String() string { return fmt.Sprintf("%v %v", bce.Proc, bce.Args) }
func (bce *BuiltinCallExpr) Compute(eng *Engine, env sx.Environment) (sx.Object, error) {
	return computeCallable(eng, env, bce.Proc, bce.Args)
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
func (bce *BuiltinCallExpr) Rework(ro *ReworkOptions, env sx.Environment) Expr {
	// Rework checks if the Builtin is a simple BuilinA and if all args are
	// constant sx.Object's. If this is true, it will call the builtin with
	// the args. If no error was signaled, the result object will be used
	// instead the BuiltinCallExpr. This assumes that there is no side effect
	// when the builtin is called.
	mayInline := true
	if _, isBuiltinA := bce.Proc.(BuiltinA); !isBuiltinA {
		mayInline = false
	}
	for i, arg := range bce.Args {
		bce.Args[i] = arg.Rework(ro, env)
		if _, isObjectExpr := bce.Args[i].(ObjectExpr); !isObjectExpr {
			mayInline = false
		}
	}
	if !mayInline {
		return bce
	}
	args := make([]sx.Object, len(bce.Args))
	for i, arg := range bce.Args {
		args[i] = arg.(ObjectExpr).Object()
	}
	result, err := bce.Proc.(BuiltinA)(args)
	if err != nil {
		return bce
	}
	return ObjExpr{Obj: result}.Rework(ro, env)
}
