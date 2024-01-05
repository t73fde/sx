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
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// CallableP returns true, if the given argument is a callable.
var CallableP = sxeval.Builtin{
	Name:     "callable?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn: func(_ *sxeval.Environment, args []sx.Object) (sx.Object, error) {
		_, ok := sxeval.GetCallable(args[0])
		return sx.MakeBoolean(ok), nil
	},
}

// DefunS parses a procedure/function specfication.
var DefunS = sxeval.Special{
	Name: "defun",
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		sym, le, err := parseDefProc(pf, args)
		if err != nil {
			return nil, err
		}
		return &DefineExpr{Sym: sym, Val: le}, nil
	},
}

func parseDefProc(pf *sxeval.ParseEnvironment, args *sx.Pair) (sx.Symbol, *LambdaExpr, error) {
	if args == nil {
		return "", nil, sxeval.ErrNoArgs
	}
	sym, isSymbol := sx.GetSymbol(args.Car())
	if !isSymbol {
		return "", nil, fmt.Errorf("not a symbol: %T/%v", args.Car(), args.Car())
	}
	args = args.Tail()
	if args == nil {
		return "", nil, fmt.Errorf("parameter spec and body missing")
	}
	le, err := ParseProcedure(pf, sx.Repr(sym), args.Car(), args.Cdr())
	return sym, le, err
}

// LambdaS parses a procedure specification.
var LambdaS = sxeval.Special{
	Name: "lambda",
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		if args == nil {
			return nil, fmt.Errorf("parameter spec and body missing")
		}
		car := args.Car()
		return ParseProcedure(pf, sx.Repr(car), car, args.Cdr())
	},
}

// ParseProcedure parses a procedure definition, where some parsing is already done.
func ParseProcedure(pf *sxeval.ParseEnvironment, name string, paramSpec, bodySpec sx.Object) (*LambdaExpr, error) {
	var params []sx.Symbol
	var rest sx.Symbol
	if !sx.IsNil(paramSpec) {
		switch p := paramSpec.(type) {
		case sx.Symbol:
			params, rest = nil, p
		case *sx.Pair:
			ps, r, err := parseProcHead(p)
			if err != nil {
				return nil, err
			}
			params, rest = ps, r
		default:
			return nil, fmt.Errorf("only symbol and list allowed in parameter spec, but got: %T/%v", p, p)
		}
	}
	if sx.IsNil(bodySpec) {
		return nil, fmt.Errorf("missing body")
	}
	body, isPair := sx.GetPair(bodySpec)
	if !isPair {
		return nil, fmt.Errorf("body must not be a dotted pair")
	}
	bindSize := len(params)
	if rest != "" {
		bindSize++
	}
	fnFrame := pf.MakeChildFrame(name+"-def", bindSize)
	for _, p := range params {
		err := fnFrame.Bind(p, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	if rest != "" {
		err := fnFrame.Bind(rest, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	expr, err := ParseExprSeq(fnFrame, body)
	if err != nil {
		return nil, err
	}
	fn := &LambdaExpr{
		Name:    name,
		Params:  params,
		Rest:    rest,
		Expr:    expr,
		IsMacro: false,
	}
	return fn, nil
}

func parseProcHead(plist *sx.Pair) (params []sx.Symbol, _ sx.Symbol, _ error) {
	for node := plist; ; {
		sym, err := GetParameterSymbol(params, node.Car())
		if err != nil {
			return nil, "", err
		}
		params = append(params, sym)

		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			return params, "", nil
		}
		if next, isPair := sx.GetPair(cdr); isPair {
			node = next
			continue
		}

		sym, err = GetParameterSymbol(params, cdr)
		if err != nil {
			return nil, "", err
		}
		return params, sym, nil
	}
}

func GetParameterSymbol(params []sx.Symbol, obj sx.Object) (sx.Symbol, error) {
	sym, isSymbol := sx.GetSymbol(obj)
	if !isSymbol {
		return "", fmt.Errorf("symbol in list expected, but got %T/%v", obj, obj)
	}
	for _, p := range params {
		if sym.IsEqual(p) {
			return "", fmt.Errorf("symbol %v already defined", sym)
		}
	}
	return sym, nil
}

type LambdaExpr struct {
	Name    string
	Params  []sx.Symbol
	Rest    sx.Symbol
	Expr    sxeval.Expr
	IsMacro bool
}

func (le *LambdaExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	bindSize := len(le.Params)
	if le.Rest != "" {
		bindSize++
	}
	fnFrame := rf.MakeChildFrame(le.Name+"-rework", bindSize)
	for _, sym := range le.Params {
		fnFrame.Bind(sym)
	}
	if rest := le.Rest; rest != "" {
		fnFrame.Bind(rest)
	}

	le.Expr = le.Expr.Rework(fnFrame)
	return le
}

func (le *LambdaExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	if le.IsMacro {
		return &Macro{
			Env:     env,
			Binding: env.Binding(),
			Name:    le.Name,
			Params:  le.Params,
			Rest:    le.Rest,
			Expr:    le.Expr,
		}, nil
	}
	return &Procedure{
		Binding: env.Binding(),
		Name:    le.Name,
		Params:  le.Params,
		Rest:    le.Rest,
		Expr:    le.Expr,
	}, nil
}

func (le *LambdaExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{LAMBDA ")
	if err != nil {
		return length, err
	}
	l, err := io.WriteString(w, le.Name)
	length += l
	if err != nil {
		return length, err
	}
	for _, p := range le.Params {
		l2, err2 := fmt.Fprintf(w, " %v", p)
		length += l2
		if err2 != nil {
			return length, err2
		}
	}
	if le.Rest == "" {
		l, err = io.WriteString(w, " none ")
	} else {
		l, err = fmt.Fprintf(w, " %v ", le.Rest)
	}
	length += l
	if err != nil {
		return length, err
	}

	l, err = le.Expr.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

// Procedure represents the procedure definition form (aka lambda).
type Procedure struct {
	Binding *sxeval.Binding
	Name    string
	Params  []sx.Symbol
	Rest    sx.Symbol
	Expr    sxeval.Expr
}

func (p *Procedure) IsNil() bool                  { return p == nil }
func (p *Procedure) IsAtom() bool                 { return p == nil }
func (p *Procedure) IsEqual(other sx.Object) bool { return p == other }
func (p *Procedure) String() string               { return p.Repr() }
func (p *Procedure) Repr() string                 { return sx.Repr(p) }
func (p *Procedure) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<lambda:", p.Name, ">")
}

// --- Builtin methods to implement sxeval.Callable

// IsPure tests if the Procedure needs a Frame value and does not produce any other side effects.
func (p *Procedure) IsPure([]sx.Object) bool { return false }

// Call the Procedure.
func (p *Procedure) Call(env *sxeval.Environment, args []sx.Object) (sx.Object, error) {
	numParams := len(p.Params)
	if len(args) < numParams {
		return nil, fmt.Errorf("%s: missing arguments: %v", p.Name, p.Params[len(args):])
	}
	bindSize := numParams
	if p.Rest != "" {
		bindSize++
	}
	lexicalEnv := env.NewLexicalEnvironment(p.Binding, p.Name, bindSize)
	for i, p := range p.Params {
		err := lexicalEnv.Bind(p, args[i])
		if err != nil {
			return nil, err
		}
	}
	if p.Rest != "" {
		err := lexicalEnv.Bind(p.Rest, sx.MakeList(args[numParams:]...))
		if err != nil {
			return nil, err
		}
	} else if len(args) > numParams {
		return nil, fmt.Errorf("%s: excess arguments: %v", p.Name, args[numParams:])
	}
	return lexicalEnv.ExecuteTCO(p.Expr)
}
