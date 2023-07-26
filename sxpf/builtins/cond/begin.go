//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package cond

import (
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// BeginS parses a begin-statement: (begin expr...).
func BeginS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	front, last, err := builtins.ParseExprSeq(eng, env, args)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return eval.NilExpr, nil
	}
	if len(front) == 0 {
		return last, nil
	}
	return &BeginExpr{front, last}, nil
}

// BeginExpr represents the begin form.
type BeginExpr struct {
	Front []eval.Expr // all expressions, but the last
	Last  eval.Expr
}

func (be *BeginExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	for _, e := range be.Front {
		_, err := eng.Execute(env, e)
		if err != nil {
			return nil, err
		}
	}
	return eng.ExecuteTCO(env, be.Last)
}
func (be *BeginExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{BEGIN")
	if err != nil {
		return length, err
	}
	l, err := eval.PrintFrontLast(w, be.Front, be.Last)
	length += l
	return length, err

}
func (be *BeginExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	for i, expr := range be.Front {
		be.Front[i] = expr.Rework(ro, env)
	}
	be.Last = be.Last.Rework(ro, env)
	return be
}
