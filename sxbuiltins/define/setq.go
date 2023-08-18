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

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// DefineSyntax parses a (define name value) form.
func SetXS(frame *sxeval.Frame, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("needs at least two arguments")
	}
	car := args.Car()
	sym, ok := sx.GetSymbol(car)
	if !ok {
		return nil, fmt.Errorf("argument 1 must be a symbol, but is: %T/%v", car, car)
	}
	val, err := parseValueDefinition(frame, args)
	if err != nil {
		return val, err
	}
	return &SetXExpr{Sym: sym, Val: val}, nil
}

// SetXExpr stores data for a set! statement.
type SetXExpr struct {
	Sym *sx.Symbol
	Val sxeval.Expr
}

func (se *SetXExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	if _, found := frame.Lookup(se.Sym); !found {
		return nil, frame.MakeNotBoundError(se.Sym)
	}
	val, err := frame.Execute(se.Val)
	if err == nil {
		err = frame.Bind(se.Sym, val)
	}
	return val, err
}
func (se *SetXExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{SET! ")
	if err != nil {
		return length, err
	}
	l, err := sx.Print(w, se.Sym)
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
func (se *SetXExpr) Rework(ro *sxeval.ReworkOptions, env sxeval.Environment) sxeval.Expr {
	se.Val = se.Val.Rework(ro, env)
	return se
}
