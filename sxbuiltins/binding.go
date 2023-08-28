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
	letExpr := LetExpr{}
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
		letExpr.Exprs = append(letExpr.Exprs, expr)
		next, isPair2 = sx.GetPair(next.Cdr())
		if !isPair2 {
			return nil, sx.ErrImproper{Pair: bindings}
		}
		node = next
	}

	es, err := ParseExprSeq(letFrame, body)
	if err != nil {
		return nil, err
	}
	letExpr.ExprSeq = es
	return &letExpr, nil
}

// LetExpr stores everything for a (let ...) expression.
type LetExpr struct {
	Symbols []*sx.Symbol
	Exprs   []sxeval.Expr
	ExprSeq
}

func (le *LetExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	le.ReworkInPlace(rf)
	return le
}
func (le *LetExpr) ReworkInPlace(rf *sxeval.ReworkFrame) {
	for i, expr := range le.Exprs {
		le.Exprs[i] = expr.Rework(rf)
	}
	le.ExprSeq.Rework(rf)
}
func (le *LetExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	letFrame := frame.MakeLetFrame("let", len(le.Symbols))
	for i, sym := range le.Symbols {
		obj, err := frame.Execute(le.Exprs[i])
		if err != nil {
			return nil, err
		}
		err = letFrame.Bind(sym, obj)
		if err != nil {
			return nil, err
		}
	}
	return le.ExprSeq.Compute(letFrame)
}
func (le *LetExpr) IsEqual(other sxeval.Expr) bool {
	if le == other {
		return true
	}
	if otherL, ok := other.(*LetExpr); ok && otherL != nil {
		return sxeval.EqualSymbolSlice(le.Symbols, otherL.Symbols) &&
			sxeval.EqualExprSlice(le.Exprs, otherL.Exprs) &&
			le.ExprSeq.IsEqual(&otherL.ExprSeq)
	}
	return false
}
func (le *LetExpr) Print(w io.Writer) (int, error) {
	return le.PrintLet(w, "{let")
}
func (le *LetExpr) PrintLet(w io.Writer, init string) (int, error) {
	length, err := io.WriteString(w, init)
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
		l, err2 = le.Exprs[i].Print(w)
		length += l
		if err2 != nil {
			return length, err2
		}
	}
	l, err := le.ExprSeq.Print(w)
	length += l
	return length, err
}

// LetStarS parses the `(let* (binding...) expr...)` syntax.`
func LetStarS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	expr, err := LetS(pf, args) // TODO: let-def must be build incrementally
	if err != nil {
		return nil, err
	}
	if letExpr, isLet := expr.(*LetExpr); isLet {
		return &LetStarExpr{*letExpr}, nil
	}
	return expr, nil
}

// LetStarExpr stores everything for a (let* ...) expression.
type LetStarExpr struct{ LetExpr }

func (lse *LetStarExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	lse.ReworkInPlace(rf)
	return lse
}
func (lse *LetStarExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	letStarFrame := frame
	for i, sym := range lse.Symbols {
		obj, err := letStarFrame.Execute(lse.Exprs[i])
		if err != nil {
			return nil, err
		}
		letStarFrame = letStarFrame.MakeLetFrame("let*", 1)
		err = letStarFrame.Bind(sym, obj)
		if err != nil {
			return nil, err
		}
	}
	return lse.ExprSeq.Compute(letStarFrame)
}

func (lse *LetStarExpr) Print(w io.Writer) (int, error) {
	return lse.PrintLet(w, "{let*")
}

// LetRecS parses the `(letrec (binding...) expr...)` syntax.`
func LetRecS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	expr, err := LetS(pf, args) // TODO: let-def must be build upfront.
	if err != nil {
		return nil, err
	}
	if letExpr, isLet := expr.(*LetExpr); isLet {
		return &LetRecExpr{*letExpr}, nil
	}
	return expr, nil
}

// LetRecExpr stores everything for a (letrec ...) expression.
type LetRecExpr struct{ LetExpr }

func (lre *LetRecExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	lre.ReworkInPlace(rf)
	return lre
}
func (lre *LetRecExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	letRecFrame := frame.MakeLetFrame("let", len(lre.Symbols)+1) // +1, because env should not freeze
	for _, sym := range lre.Symbols {
		letRecFrame.Bind(sym, sx.MakeUndefined())
	}
	for i, sym := range lre.Symbols {
		obj, err := letRecFrame.Execute(lre.Exprs[i])
		if err != nil {
			return nil, err
		}
		err = letRecFrame.Bind(sym, obj)
		if err != nil {
			return nil, err
		}
	}
	return lre.ExprSeq.Compute(letRecFrame)
}

func (lre *LetRecExpr) Print(w io.Writer) (int, error) {
	return lre.PrintLet(w, "{letrec")
}
