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
)

const letName = "let"

// LetS parses the `(let (binding...) expr...)` syntax.`
var LetS = sxeval.Special{
	Name: letName,
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		return parseBindingsBody(pf, args)
	},
}

var errNoBindingSpecAndBody = errors.New("binding spec and body missing")

func parseBindingsBody(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, errNoBindingSpecAndBody
	}
	argsCar := args.Car()
	bindings, isBindings := sx.GetPair(argsCar)
	if !isBindings {
		return nil, fmt.Errorf("bindings must be a list, but is %T/%v", argsCar, argsCar)
	}
	var symbols []*sx.Symbol
	var objs []sx.Object
	for node := bindings; node != nil; {
		car := node.Car()
		binding, isPair := sx.GetPair(car)
		if !isPair || binding == nil {
			return nil, fmt.Errorf("single binding must be a list, but is %T/%v", car, car)
		}
		sym, err := GetParameterSymbol(symbols, binding.Car())
		if err != nil {
			return nil, err
		}
		pair, isPair := sx.GetPair(binding.Cdr())
		if !isPair {
			return nil, sx.ErrImproper{Pair: binding}
		}
		if pair == nil {
			return nil, fmt.Errorf("binding missing for symbol %v", sym)
		}
		if cdr := pair.Cdr(); !sx.IsNil(cdr) {
			return nil, fmt.Errorf("too many bindings for symbol %v: %T/%v", sym, cdr, cdr)
		}
		symbols = append(symbols, sym)
		objs = append(objs, pair.Car())

		next, isPair := sx.GetPair(node.Cdr())
		if !isPair {
			return nil, sx.ErrImproper{Pair: bindings}
		}
		node = next
	}

	vals := make([]sxeval.Expr, len(objs))
	letEnv := pf.MakeChildEnvironment("let-def", len(symbols))
	for i, sym := range symbols {
		expr, err := pf.Parse(objs[i])
		if err != nil {
			return nil, err
		}
		if err = letEnv.Bind(sym, sx.MakeUndefined()); err != nil {
			return nil, err
		}
		vals[i] = expr
	}

	bodyArgs, isPair := sx.GetPair(args.Cdr())
	if !isPair {
		return nil, sx.ErrImproper{Pair: args}
	}
	body, err := ParseExprSeq(letEnv, bodyArgs)
	if err != nil {
		return nil, err
	}
	return &LetExpr{
		Symbols: symbols,
		Vals:    vals,
		Body:    body,
	}, nil
}

// LetExpr stores everything for a (let ...) expression.
type LetExpr struct {
	Symbols []*sx.Symbol
	Vals    []sxeval.Expr
	Body    sxeval.Expr
}

// Unparse the expression as an sx.Object
func (le *LetExpr) Unparse() sx.Object {
	var bindings sx.ListBuilder
	for i, sym := range le.Symbols {
		bindings.Add(sx.MakeList(sym, le.Vals[i].Unparse()))
	}
	body := le.Body.Unparse()
	return sx.MakeList(sx.MakeSymbol(letName), bindings.List(), body)
}

// Improve the expression into a possible simpler one.
func (le *LetExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	switch len(le.Vals) {
	case 0:
		return imp.Improve(le.Body)
	case 1:
		le1 := &Let1Expr{
			Symbol: le.Symbols[0],
			Value:  le.Vals[0],
			Body:   le.Body,
		}
		return imp.Improve(le1)
	}
	letEnv := imp.MakeChildImprover("let-improve", len(le.Vals))
	for i, val := range le.Vals {
		expr, err := imp.Improve(val)
		if err != nil {
			return le, err
		}
		le.Vals[i] = expr
		_ = letEnv.Bind(le.Symbols[i])
	}
	expr, err := letEnv.Improve(le.Body)
	if err == nil {
		le.Body = expr
	}
	return le, err
}

// Compile the expression.
func (le *LetExpr) Compile(sxc *sxeval.Compiler, tailPos bool) error {
	for _, val := range le.Vals {
		if err := sxc.Compile(val, false); err != nil {
			return nil
		}
	}
	syms := le.Symbols
	sxc.Emit(func(env *sxeval.Environment) error {
		letEnv := env.NewLexicalEnvironment(env.Binding(), "let", len(syms))
		for i, arg := range env.Args(len(syms)) {
			if err := letEnv.Bind(syms[i], arg); err != nil {
				return err
			}
		}
		return sxeval.SwitchEnvironment(letEnv)
	}, "LET", fmt.Sprintf("%v", syms))
	sxc.EmitKill(len(syms))
	return sxc.Compile(le.Body, tailPos)
}

// Compute the expression in a frame and return the result.
func (le *LetExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	letEnv := env.NewLexicalEnvironment(env.Binding(), "let", len(le.Symbols))
	for i, sym := range le.Symbols {
		obj, err := env.Execute(le.Vals[i])
		if err != nil {
			return nil, err
		}
		if err = letEnv.Bind(sym, obj); err != nil {
			return nil, err
		}
	}
	return letEnv.ExecuteTCO(le.Body)
}

// Print the expression on the given writer.
func (le *LetExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{LET (")
	if err != nil {
		return length, err
	}
	var l int
	for i, sym := range le.Symbols {
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
		l, err = le.Vals[i].Print(w)
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
	l, err = le.Body.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

// Let1Expr is a special case of `LetExpr`, where there is just one binding.
type Let1Expr struct {
	Symbol *sx.Symbol
	Value  sxeval.Expr
	Body   sxeval.Expr
}

// Unparse the expression as an sx.Object
func (le1 *Let1Expr) Unparse() sx.Object {
	return sx.MakeList(
		sx.MakeSymbol(letName),
		sx.Cons(sx.MakeList(le1.Symbol, le1.Value.Unparse()), sx.Nil()),
		le1.Body.Unparse(),
	)
}

// Improve the expression into a possible simpler one.
func (le1 *Let1Expr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	letEnv := imp.MakeChildImprover("let1-improve", 1)
	expr, err := imp.Improve(le1.Value)
	if err != nil {
		return le1, err
	}
	le1.Value = expr
	_ = letEnv.Bind(le1.Symbol)
	expr, err = letEnv.Improve(le1.Body)
	if err == nil {
		le1.Body = expr
	}
	return le1, err
}

// Compile the expression.
func (le1 *Let1Expr) Compile(sxc *sxeval.Compiler, tailPos bool) error {
	if err := sxc.Compile(le1.Value, false); err != nil {
		return nil
	}
	sxc.AdjustStack(-1)
	sym := le1.Symbol
	sxc.Emit(func(env *sxeval.Environment) error {
		letEnv := env.NewLexicalEnvironment(env.Binding(), "let1", 1)
		if err := letEnv.Bind(sym, env.Pop()); err != nil {
			return err
		}
		return sxeval.SwitchEnvironment(letEnv)
	}, "LET1", sym.String())
	return sxc.Compile(le1.Body, tailPos)
}

// Compute the expression in a frame and return the result.
func (le1 *Let1Expr) Compute(env *sxeval.Environment) (sx.Object, error) {
	letEnv := env.NewLexicalEnvironment(env.Binding(), "let1", 1)
	obj, err := env.Execute(le1.Value)
	if err != nil {
		return nil, err
	}
	if err = letEnv.Bind(le1.Symbol, obj); err != nil {
		return nil, err
	}
	return letEnv.ExecuteTCO(le1.Body)
}

// Print the expression on the given writer.
func (le1 *Let1Expr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{LET1 ((")
	if err != nil {
		return length, err
	}
	var l int
	l, err = le1.Symbol.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, " ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = le1.Value.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, ")) ")
	length += l
	if err != nil {
		return length, err
	}
	l, err = le1.Body.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
