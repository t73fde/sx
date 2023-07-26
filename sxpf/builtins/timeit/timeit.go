//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package timeit provides functions to measure evaluation.
package timeit

import (
	"fmt"
	"io"
	"time"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// TimeitS is a syntax to measure code execution time.
func TimeitS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("one argument expected")
	}
	expr, err := eng.Parse(env, args.Car())
	if err != nil {
		return nil, err
	}
	return &TimeitExpr{expr}, nil
}

// TimeitExpr stores information to measure execution time.
type TimeitExpr struct {
	expr eval.Expr
}

func (te *TimeitExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	start := time.Now()
	obj, err := eng.Execute(env, te.expr)
	duration := sxpf.MakeString(time.Since(start).String())
	if err == nil {
		return sxpf.MakeList(duration, obj), nil
	}
	return sxpf.MakeList(duration, obj, sxpf.MakeString(err.Error())), nil
}
func (te *TimeitExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{TIMEIT ")
	if err != nil {
		return length, err
	}
	l, err := te.expr.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
func (te *TimeitExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	te.expr = te.expr.Rework(ro, env)
	return te
}
