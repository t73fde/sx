//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// BeginS parses a sequence of expressions.
var BeginS = sxeval.Special{
	Name: "begin",
	Fn:   ParseExprSeq,
}

// ParseExprSeq parses a sequence of expressions.
func ParseExprSeq(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return sxeval.NilExpr, nil
	}
	var front []sxeval.Expr
	var last sxeval.Expr
	for node := args; ; {
		ex, err := pf.Parse(node.Car())
		if err != nil {
			return nil, err
		}
		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			last = ex
			break
		}
		front = append(front, ex)
		if next, isPair := sx.GetPair(cdr); isPair {
			node = next
			continue
		}
		ex, err = pf.Parse(cdr)
		if err != nil {
			return nil, err
		}
		last = ex
		break
	}
	if len(front) == 0 {
		return last, nil
	}
	return &BeginExpr{Front: front, Last: last}, nil
}

// BeginExpr represents the begin form.
type BeginExpr struct {
	Front []sxeval.Expr
	Last  sxeval.Expr
}

func (be *BeginExpr) Rework(re *sxeval.ReworkEnvironment) sxeval.Expr {
	last := be.Last.Rework(re)
	frontLen := len(be.Front)
	if frontLen == 0 {
		return last
	}
	seq := make([]sxeval.Expr, 0, frontLen)
	for _, expr := range be.Front {
		re := expr.Rework(re)
		if _, isConstObject := re.(sxeval.ConstObjectExpr); isConstObject {
			// An object has no side effect, there it can be ignored in the sequence
			continue
		}
		seq = append(seq, re)
	}
	if seqLen := len(seq); seqLen == 0 {
		return last
	} else if seqLen == cap(be.Front) {
		copy(be.Front, seq)
	} else {
		newFront := make([]sxeval.Expr, seqLen)
		copy(newFront, seq)
		be.Front = newFront
	}
	be.Last = last
	return be
}

func (be *BeginExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	for _, e := range be.Front {
		subEnv := env.NewDynamicEnvironment()
		_, err := subEnv.Execute(e)
		if err != nil {
			return nil, err
		}
	}
	return env.ExecuteTCO(be.Last)
}

func (be *BeginExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{BEGIN")
	if err != nil {
		return length, err
	}
	var l int
	for _, expr := range be.Front {
		l, err = io.WriteString(w, " ")
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
	l, err = be.Last.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
