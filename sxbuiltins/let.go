//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"errors"
	"fmt"
	"io"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
	"t73f.de/r/zero/set"
)

// LetData stores basic information about let bindings.
type LetData struct {
	Symbols []*sx.Symbol
	Vals    []sxeval.Expr
	Body    sxeval.Expr
}

var errNoBindingSpecAndBody = errors.New("binding spec and body missing")

func parseBindingsBody(pf *sxeval.ParseEnvironment, args *sx.Pair, dupSyms bool, data *LetData) error {
	if args == nil {
		return errNoBindingSpecAndBody
	}
	argsCar := args.Car()
	bindings, isBindings := sx.GetPair(argsCar)
	if !isBindings {
		return fmt.Errorf("bindings must be a list, but is %T/%v", argsCar, argsCar)
	}
	var symbols []*sx.Symbol
	var objs []sx.Object
	for node := bindings; node != nil; {
		car := node.Car()
		binding, isPair := sx.GetPair(car)
		if !isPair || binding == nil {
			return fmt.Errorf("single binding must be a list, but is %T/%v", car, car)
		}
		var sym *sx.Symbol
		var err error
		if dupSyms {
			sym, err = GetParameterSymbol(nil, binding.Car())
		} else {
			sym, err = GetParameterSymbol(symbols, binding.Car())
		}
		if err != nil {
			return err
		}
		pair, isPair := sx.GetPair(binding.Cdr())
		if !isPair {
			return sx.ErrImproper{Pair: binding}
		}
		if pair == nil {
			return fmt.Errorf("binding missing for symbol %v", sym)
		}
		if cdr := pair.Cdr(); !sx.IsNil(cdr) {
			return fmt.Errorf("too many bindings for symbol %v: %T/%v", sym, cdr, cdr)
		}
		symbols = append(symbols, sym)
		objs = append(objs, pair.Car())

		next, isPair := sx.GetPair(node.Cdr())
		if !isPair {
			return sx.ErrImproper{Pair: bindings}
		}
		node = next
	}

	vals := make([]sxeval.Expr, len(objs))
	letEnv := pf.MakeChildEnvironment("let-def", len(symbols))
	for i, sym := range symbols {
		expr, err := pf.Parse(objs[i])
		if err != nil {
			return err
		}
		if err = letEnv.Bind(sym, sx.MakeUndefined()); err != nil {
			return err
		}
		vals[i] = expr
	}

	bodyArgs, isPair := sx.GetPair(args.Cdr())
	if !isPair {
		return sx.ErrImproper{Pair: args}
	}
	body, err := ParseExprSeq(letEnv, bodyArgs)
	if err != nil {
		return err
	}
	*data = LetData{symbols, vals, body}
	return nil
}

// Unparse the expression as an sx.Object
func (ld *LetData) Unparse(letSym *sx.Symbol) sx.Object {
	var bindings sx.ListBuilder
	for i, sym := range ld.Symbols {
		bindings.Add(sx.MakeList(sym, ld.Vals[i].Unparse()))
	}
	body := ld.Body.Unparse()
	return sx.MakeList(letSym, bindings.List(), body)
}

// Print the expression on the given writer.
func (ld *LetData) Print(w io.Writer, prefix string) (int, error) {
	length, err := io.WriteString(w, prefix)
	if err != nil {
		return length, err
	}
	var l int
	for i, sym := range ld.Symbols {
		if i == 0 {
			l, err = io.WriteString(w, "(")
		} else {
			l, err = io.WriteString(w, " (")
		}
		length += l
		if err != nil {
			return length, err
		}
		l, err = sym.Print(w)
		length += l
		if err != nil {
			return length, err
		}
		l, err = io.WriteString(w, " ")
		length += l
		if err != nil {
			return length, err
		}
		l, err = ld.Vals[i].Print(w)
		length += l
		if err != nil {
			return length, err
		}
		l, err = io.WriteString(w, ")")
		length += l
		if err != nil {
			return length, err
		}
	}
	l, err = io.WriteString(w, ") ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = ld.Body.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

// ----- (let ...)

const letName = "let"

// LetS parses the `(let (binding...) expr...)` syntax.`
var LetS = sxeval.Special{
	Name: letName,
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		var result LetExpr
		if err := parseBindingsBody(pf, args, false, &result.LetData); err != nil {
			return nil, err
		}
		return &result, nil
	},
}

// LetExpr stores everything for a (let ...) expression.
type LetExpr struct {
	LetData
}

// IsPure signals an expression that has no side effects.
func (*LetExpr) IsPure() bool { return false } // TODO: check pure-ness of binding-creation and pure-ness of body

// Unparse the expression as an sx.Object
func (le *LetExpr) Unparse() sx.Object { return le.LetData.Unparse(sx.MakeSymbol(letName)) }

// Improve the expression into a possible simpler one.
func (le *LetExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	if len(le.Vals) == 0 {
		return imp.Improve(le.Body)
	}

	if err := imp.ImproveSlice(le.Vals); err != nil {
		return le, err
	}

	letImp := imp.MakeChildImprover("let-improve", len(le.Vals))
	for _, sym := range le.Symbols {
		_ = letImp.Bind(sym)
	}
	expr, err := letImp.Improve(le.Body)
	if err == nil {
		le.Body = expr
	}
	return le, err
}

// Compute the expression in a frame and return the result.
func (le *LetExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	syms, vals := le.Symbols, le.Vals
	letBind := bind.MakeChildBinding("let", len(syms))
	for i, sym := range syms {
		obj, err := env.Execute(vals[i], bind)
		if err != nil {
			return nil, err
		}
		if err = letBind.Bind(sym, obj); err != nil {
			return nil, err
		}
	}
	return env.ExecuteTCO(le.Body, letBind)
}

// Print the expression on the given writer.
func (le *LetExpr) Print(w io.Writer) (int, error) { return le.LetData.Print(w, "{LET (") }

// ----- (let* ...)

const letStarName = "let*"

// LetStarS parses the `(let* (binding...) expr...)` syntax.`
var LetStarS = sxeval.Special{
	Name: letStarName,
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		var result LetStarExpr
		if err := parseBindingsBody(pf, args, true, &result.LetData); err != nil {
			return nil, err
		}
		result.numSymbols = set.New(result.Symbols...).Length()
		return &result, nil
	},
}

// LetStarExpr stores everything for a (let* ...) expression.
type LetStarExpr struct {
	LetData

	numSymbols int // number of unique symbols, since symbols may be given multiple times in (let* ...)
}

// IsPure signals an expression that has no side effects.
func (*LetStarExpr) IsPure() bool { return false } // TODO: check pure-ness: bindings, body

// Unparse the expression as an sx.Object
func (lse *LetStarExpr) Unparse() sx.Object { return lse.LetData.Unparse(sx.MakeSymbol(letStarName)) }

// Improve the expression into a possible simpler one.
func (lse *LetStarExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	if len(lse.Vals) < 2 {
		return (&LetExpr{LetData: lse.LetData}).Improve(imp)
	}

	letStarImp := imp
	for i, expr := range lse.Vals {
		iexpr, err := letStarImp.Improve(expr)
		if err != nil {
			return lse, err
		}
		lse.Vals[i] = iexpr
		if i == 0 {
			letStarImp = imp.MakeChildImprover("let*-improve", lse.numSymbols)
		}
		_ = letStarImp.Bind(lse.Symbols[i])
	}

	expr, err := letStarImp.Improve(lse.Body)
	if err == nil {
		lse.Body = expr
	}
	return lse, err
}

// Compute the expression in a frame and return the result.
func (lse *LetStarExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	syms, vals := lse.Symbols, lse.Vals
	letStarBind := bind
	for i, sym := range syms {
		obj, err := env.Execute(vals[i], letStarBind)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			letStarBind = bind.MakeChildBinding("let*", lse.numSymbols)
		}
		if err = letStarBind.Bind(sym, obj); err != nil {
			return nil, err
		}
	}
	return env.ExecuteTCO(lse.Body, letStarBind)
}

// Print the expression on the given writer.
func (lse *LetStarExpr) Print(w io.Writer) (int, error) { return lse.LetData.Print(w, "{LET* (") }
