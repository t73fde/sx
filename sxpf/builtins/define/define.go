//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package define contains all syntaxes and builtins to bind values to symbols.
package define

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins/callable"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// DefineS parses a (define name value) form.
func DefineS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("need at least two arguments")
	}
	switch car := args.Car().(type) {
	case *sxpf.Symbol:
		val, err := parseValueDefinition(eng, env, args)
		if err != nil {
			return val, err
		}
		return &DefineExpr{Sym: car, Val: val}, nil
	case *sxpf.Pair:
		sym, fun, err := parseProcedureDefinition(eng, env, car, args)
		if err != nil {
			return fun, err
		}
		return &DefineExpr{Sym: sym, Val: fun}, nil
	default:
		return nil, fmt.Errorf("argument 1 must be a symbol or a list, but is: %T/%v", car, car)
	}
}

func parseValueDefinition(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	cdr := args.Cdr()
	if sxpf.IsNil(cdr) {
		return nil, fmt.Errorf("argument 2 missing")
	}
	pair, isPair := sxpf.GetPair(cdr)
	if !isPair {
		return nil, fmt.Errorf("argument 2 must be a proper list")
	}
	return eng.Parse(env, pair.Car())
}

func parseProcedureDefinition(eng *eval.Engine, env sxpf.Environment, head, args *sxpf.Pair) (*sxpf.Symbol, eval.Expr, error) {
	if head == nil {
		return nil, nil, fmt.Errorf("empty function head")
	}
	sym, ok := sxpf.GetSymbol(head.Car())
	if !ok {
		return nil, nil, fmt.Errorf("first element in function head is not a symbol, but: %T/%v", head.Car(), head.Car())
	}
	expr, err := callable.ParseProcedure(eng, env, sym.Name(), head.Cdr(), args.Cdr())
	return sym, expr, err
}

// DefineExpr stores data for a define statement.
type DefineExpr struct {
	Sym *sxpf.Symbol
	Val eval.Expr
}

func (de *DefineExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	val, err := eng.Execute(env, de.Val)
	if err == nil {
		err = env.Bind(de.Sym, val)
	}
	return val, err
}
func (de *DefineExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{DEFINE ")
	if err != nil {
		return length, err
	}
	l, err := sxpf.Print(w, de.Sym)
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
func (de *DefineExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	de.Val = de.Val.Rework(ro, env)
	return de
}
