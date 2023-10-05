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
	bindings, body, err := ScanLet(args)
	if err != nil {
		return nil, err
	}
	if bindings == nil {
		return BeginS(pf, body)
	}
	symbols, objs, err := ScanLetBindings(bindings)
	if err != nil {
		return nil, err
	}
	letExpr := LetExpr{Symbols: symbols}
	letFrame := pf.MakeLetFrame("let-def", len(symbols))
	for i, sym := range symbols {
		expr, err2 := pf.Parse(objs[i])
		if err2 != nil {
			return nil, err2
		}
		err2 = letFrame.Bind(sym, sx.MakeUndefined())
		if err2 != nil {
			return nil, err2
		}
		letExpr.Exprs = append(letExpr.Exprs, expr)
	}
	es, err := ParseExprSeq(letFrame, body)
	letExpr.ExprSeq = es
	return &letExpr, err
}

// ScanLet scans and checks the existence of bindings and body.
func ScanLet(args *sx.Pair) (bindings, body *sx.Pair, err error) {
	if args == nil {
		return nil, nil, fmt.Errorf("binding spec and body missing")
	}
	bindings, isPair := sx.GetPair(args.Car())
	if !isPair {
		return nil, nil, fmt.Errorf("binding spec must be a list, but got: %t/%v", args.Car(), args.Car())
	}
	body, isPair = sx.GetPair(args.Cdr())
	if !isPair {
		return nil, nil, sx.ErrImproper{Pair: args}
	}
	return bindings, body, nil
}

// ScanLetBinding scans the bindings and returns the slice of symbols and
// objects (which have yet to be parsed).
func ScanLetBindings(bindings *sx.Pair) (symbols []*sx.Symbol, objs []sx.Object, _ error) {
	for node := bindings; node != nil; {
		sym, err := GetParameterSymbol(symbols, node.Car())
		if err != nil {
			return nil, nil, err
		}
		next, isPair := sx.GetPair(node.Cdr())
		if !isPair {
			return nil, nil, sx.ErrImproper{Pair: bindings}
		}
		if next == nil {
			return nil, nil, fmt.Errorf("binding missing for %v", sym)
		}
		symbols = append(symbols, sym)
		objs = append(objs, next.Car())
		next, isPair = sx.GetPair(next.Cdr())
		if !isPair {
			return nil, nil, sx.ErrImproper{Pair: bindings}
		}
		node = next
	}
	return symbols, objs, nil
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
		subFrame := frame.MakeCalleeFrame()
		obj, err := subFrame.Execute(le.Exprs[i])
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
	bindings, body, err := ScanLet(args)
	if err != nil {
		return nil, err
	}
	if bindings == nil {
		return BeginS(pf, body)
	}
	symbols, objs, err := ScanLetBindings(bindings)
	if err != nil {
		return nil, err
	}
	le := LetExpr{Symbols: symbols}
	letStarFrame := pf
	for i, sym := range symbols {
		expr, err2 := letStarFrame.Parse(objs[i])
		if err2 != nil {
			return nil, err2
		}
		letStarFrame = letStarFrame.MakeLetFrame("let*-def", 1)
		err2 = letStarFrame.Bind(sym, sx.MakeUndefined())
		if err2 != nil {
			return nil, err2
		}
		le.Exprs = append(le.Exprs, expr)
	}
	es, err := ParseExprSeq(letStarFrame, body)
	le.ExprSeq = es
	return &LetStarExpr{le}, err
}

// LetStarExpr stores everything for a (let* ...) expression.
type LetStarExpr struct{ LetExpr }

func (lse *LetStarExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	lse.ReworkInPlace(rf)
	if len(lse.Symbols) <= 1 {
		return &lse.LetExpr
	}
	return lse
}
func (lse *LetStarExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	letStarFrame := frame
	for i, sym := range lse.Symbols {
		subFrame := letStarFrame.MakeCalleeFrame()
		obj, err := subFrame.Execute(lse.Exprs[i])
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
	bindings, body, err := ScanLet(args)
	if err != nil {
		return nil, err
	}
	if bindings == nil {
		return BeginS(pf, body)
	}
	symbols, objs, err := ScanLetBindings(bindings)
	if err != nil {
		return nil, err
	}
	le := LetExpr{Symbols: symbols}
	letRecFrame := pf.MakeLetFrame("letrec-def", len(symbols))
	for _, sym := range symbols {
		err = letRecFrame.Bind(sym, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	for _, obj := range objs {
		expr, err2 := letRecFrame.Parse(obj)
		if err2 != nil {
			return nil, err2
		}
		le.Exprs = append(le.Exprs, expr)
	}
	es, err := ParseExprSeq(letRecFrame, body)
	le.ExprSeq = es
	return &LetRecExpr{le}, err
}

// LetRecExpr stores everything for a (letrec ...) expression.
type LetRecExpr struct{ LetExpr }

func (lre *LetRecExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	lre.ReworkInPlace(rf)
	if len(lre.Symbols) <= 1 {
		return &lre.LetExpr
	}
	return lre
}
func (lre *LetRecExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	letRecFrame := frame.MakeLetFrame("let", len(lre.Symbols))
	for _, sym := range lre.Symbols {
		letRecFrame.Bind(sym, sx.MakeUndefined())
	}
	for i, sym := range lre.Symbols {
		subFrame := letRecFrame.MakeCalleeFrame()
		obj, err := subFrame.Execute(lre.Exprs[i])
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
