//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package define contains all syntaxes and builtins to bind values to symbols.
package define

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins/callable"
	"zettelstore.de/sx.fossil/sxeval"
)

// DefineS parses a (define name value) form.
func DefineS(frame *sxeval.Frame, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("need at least two arguments")
	}
	switch car := args.Car().(type) {
	case *sx.Symbol:
		val, err := parseValueDefinition(frame, args)
		if err != nil {
			return val, err
		}
		return &DefineExpr{Sym: car, Val: val}, nil
	case *sx.Pair:
		sym, fun, err := parseProcedureDefinition(frame, car, args)
		if err != nil {
			return fun, err
		}
		return &DefineExpr{Sym: sym, Val: fun}, nil
	default:
		return nil, fmt.Errorf("argument 1 must be a symbol or a list, but is: %T/%v", car, car)
	}
}

func parseValueDefinition(frame *sxeval.Frame, args *sx.Pair) (sxeval.Expr, error) {
	cdr := args.Cdr()
	if sx.IsNil(cdr) {
		return nil, fmt.Errorf("argument 2 missing")
	}
	pair, isPair := sx.GetPair(cdr)
	if !isPair {
		return nil, fmt.Errorf("argument 2 must be a proper list")
	}
	return frame.Parse(pair.Car())
}

func parseProcedureDefinition(frame *sxeval.Frame, head, args *sx.Pair) (*sx.Symbol, sxeval.Expr, error) {
	if head == nil {
		return nil, nil, fmt.Errorf("empty function head")
	}
	sym, ok := sx.GetSymbol(head.Car())
	if !ok {
		return nil, nil, fmt.Errorf("first element in function head is not a symbol, but: %T/%v", head.Car(), head.Car())
	}
	expr, err := callable.ParseProcedure(frame, sym.Name(), head.Cdr(), args.Cdr())
	return sym, expr, err
}

// DefineExpr stores data for a define statement.
type DefineExpr struct {
	Sym *sx.Symbol
	Val sxeval.Expr
}

func (de *DefineExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	val, err := frame.Execute(de.Val)
	if err == nil {
		err = frame.Bind(de.Sym, val)
	}
	return val, err
}
func (de *DefineExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{DEFINE ")
	if err != nil {
		return length, err
	}
	l, err := sx.Print(w, de.Sym)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = de.Val.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
func (de *DefineExpr) Rework(ro *sxeval.ReworkOptions, env sxeval.Environment) sxeval.Expr {
	de.Val = de.Val.Rework(ro, env)
	return de
}
