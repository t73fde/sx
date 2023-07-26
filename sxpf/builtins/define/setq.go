//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package define

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// DefineSyntax parses a (define name value) form.
func SetXS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("needs at least two arguments")
	}
	car := args.Car()
	sym, ok := sxpf.GetSymbol(car)
	if !ok {
		return nil, fmt.Errorf("argument 1 must be a symbol, but is: %T/%v", car, car)
	}
	val, err := parseValueDefinition(eng, env, args)
	if err != nil {
		return val, err
	}
	return &SetXExpr{Sym: sym, Val: val}, nil
}

// SetXExpr stores data for a set! statement.
type SetXExpr struct {
	Sym *sxpf.Symbol
	Val eval.Expr
}

func (se *SetXExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	if _, found := env.Lookup(se.Sym); !found {
		return nil, eval.NotBoundError{Env: env, Sym: se.Sym}
	}
	val, err := eng.Execute(env, se.Val)
	if err == nil {
		err = env.Bind(se.Sym, val)
	}
	return val, err
}
func (se *SetXExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{SET! ")
	if err != nil {
		return length, err
	}
	l, err := sxpf.Print(w, se.Sym)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = se.Val.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
func (se *SetXExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	se.Val = se.Val.Rework(ro, env)
	return se
}
