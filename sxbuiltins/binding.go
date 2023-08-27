//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

// Package binding contains builtins and syntax to bind values.

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// LetS parses the `(let (binding...) expr...)` syntax.`
func LetS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("binding spec and body missing")
	}
	bindings, isPair := sx.GetPair(args.Car())
	if !isPair {
		return nil, fmt.Errorf("binding spec must be a list, but got: %t/%v", args.Car(), args.Car())
	}
	body, isPair := sx.GetPair(args.Cdr())
	if !isPair {
		return nil, sx.ErrImproper{Pair: args}
	}
	if bindings == nil {
		return BeginS(pf, body)
	}
	letExpr := LetExpr{
		Symbols: nil,
		Expr:    nil,
		Front:   nil,
		Last:    nil,
	}
	letFrame := pf.MakeChildFrame("let-def", 128)
	for node := bindings; node != nil; {
		sym, err := GetParameterSymbol(letExpr.Symbols, node.Car())
		if err != nil {
			return nil, err
		}
		next, isPair2 := sx.GetPair(node.Cdr())
		if !isPair2 {
			return nil, sx.ErrImproper{Pair: bindings}
		}
		if next == nil {
			return nil, fmt.Errorf("binding missing for %v", sym)
		}
		expr, err := pf.Parse(next.Car())
		if err != nil {
			return nil, err
		}
		err = letFrame.Bind(sym, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
		letExpr.Symbols = append(letExpr.Symbols, sym)
		letExpr.Expr = append(letExpr.Expr, expr)
		next, isPair2 = sx.GetPair(next.Cdr())
		if !isPair2 {
			return nil, sx.ErrImproper{Pair: bindings}
		}
		node = next
	}

	front, last, err := ParseExprSeq(letFrame, body)
	if err != nil {
		return nil, err
	}
	letExpr.Front = front
	letExpr.Last = last
	return &letExpr, nil
}

// LetExpr stores everything for a (let ...) expression.
type LetExpr struct {
	Symbols []*sx.Symbol
	Expr    []sxeval.Expr
	Front   []sxeval.Expr
	Last    sxeval.Expr
}

func (le *LetExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	for i, expr := range le.Expr {
		le.Expr[i] = expr.Rework(rf)
	}
	for i, expr := range le.Front {
		le.Front[i] = expr.Rework(rf)
	}
	le.Last = le.Last.Rework(rf)
	return le
}
func (le *LetExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	letFrame := frame.MakeLetFrame("let", len(le.Symbols))
	for i, sym := range le.Symbols {
		obj, err := frame.Execute(le.Expr[i])
		if err != nil {
			return nil, err
		}
		err = letFrame.Bind(sym, obj)
		if err != nil {
			return nil, err
		}
	}
	for _, expr := range le.Front {
		_, err := letFrame.Execute(expr)
		if err != nil {
			return nil, err
		}
	}
	return letFrame.ExecuteTCO(le.Last)
}
func (le *LetExpr) IsEqual(other sxeval.Expr) bool {
	if le == other {
		return true
	}
	if otherL, ok := other.(*LetExpr); ok && otherL != nil {
		return sxeval.EqualSymbolSlice(le.Symbols, otherL.Symbols) &&
			sxeval.EqualExprSlice(le.Expr, otherL.Expr) &&
			sxeval.EqualExprSlice(le.Front, otherL.Front) &&
			le.Last.IsEqual(otherL.Last)
	}
	return false
}
func (le *LetExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{LET")
	if err != nil {
		return length, err
	}
	for i, sym := range le.Symbols {
		l, err2 := io.WriteString(w, " ")
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = io.WriteString(w, sym.Repr())
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = io.WriteString(w, " ")
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = le.Expr[i].Print(w)
		length += l
		if err2 != nil {
			return length, err2
		}
	}
	l, err := sxeval.PrintFrontLast(w, le.Front, le.Last)
	length += l
	return length, err
}
