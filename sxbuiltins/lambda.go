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

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// CallableP returns True, if the given argument is a callable.
func CallableP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sxeval.GetCallable(args[0])
	return sx.MakeBoolean(ok), nil
}

// LambdaS parses a procedure specification.
func LambdaS(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, fmt.Errorf("parameter spec and body missing")
	}
	car := args.Car()
	return ParseProcedure(pf, sx.Repr(car), car, args.Cdr())
}

// ParseProcedure parses a procedure definition, where some parsing is already done.
func ParseProcedure(pf *sxeval.ParseFrame, name string, paramSpec, bodySpec sx.Object) (*LambdaExpr, error) {
	var params []*sx.Symbol
	var rest *sx.Symbol
	if !sx.IsNil(paramSpec) {
		switch p := paramSpec.(type) {
		case *sx.Symbol:
			params, rest = nil, p
		case *sx.Pair:
			ps, r, err := parseProcHead(p)
			if err != nil {
				return nil, err
			}
			params, rest = ps, r
		default:
			return nil, fmt.Errorf("only symbol and list allowed in parameter spec: %v,%v", p, p)
		}
	}
	if sx.IsNil(bodySpec) {
		return nil, fmt.Errorf("missing body")
	}
	body, isPair := sx.GetPair(bodySpec)
	if !isPair {
		return nil, fmt.Errorf("body must not be a dotted pair")
	}
	envSize := len(params)
	if rest != nil {
		envSize++
	}
	fnFrame := pf.MakeChildFrame(name+"-def", envSize)
	for _, p := range params {
		err := fnFrame.Bind(p, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	if rest != nil {
		err := fnFrame.Bind(rest, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	front, last, err := ParseExprSeq(fnFrame, body)
	if err != nil {
		return nil, err
	}

	fn := &LambdaExpr{
		Name:   name,
		Params: params,
		Rest:   rest,
		Front:  front,
		Last:   last,
	}
	return fn, nil
}

func parseProcHead(plist *sx.Pair) (params []*sx.Symbol, _ *sx.Symbol, _ error) {
	for node := plist; ; {
		sym, err := GetParameterSymbol(params, node.Car())
		if err != nil {
			return nil, nil, err
		}
		params = append(params, sym)

		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			return params, nil, nil
		}
		if next, isPair := sx.GetPair(cdr); isPair {
			node = next
			continue
		}

		sym, err = GetParameterSymbol(params, cdr)
		if err != nil {
			return nil, nil, err
		}
		return params, sym, nil
	}
}

func GetParameterSymbol(params []*sx.Symbol, obj sx.Object) (*sx.Symbol, error) {
	sym, ok := sx.GetSymbol(obj)
	if !ok {
		return nil, fmt.Errorf("symbol in list expected, but got %T/%v", obj, obj)
	}
	for _, p := range params {
		if sym.IsEql(p) {
			return nil, fmt.Errorf("symbol %v already defined", sym)
		}
	}
	return sym, nil
}

type LambdaExpr struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Front  []sxeval.Expr // all expressions, but the last
	Last   sxeval.Expr
}

func (le *LambdaExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	for i, expr := range le.Front {
		le.Front[i] = expr.Rework(rf)
	}
	le.Last = le.Last.Rework(rf)
	return le
}
func (le *LambdaExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	return &Procedure{
		PFrame: frame.MakeParseFrame(),
		Name:   le.Name,
		Params: le.Params,
		Rest:   le.Rest,
		Front:  le.Front,
		Last:   le.Last,
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
	if le.Rest == nil {
		l, err = io.WriteString(w, " none ")
	} else {
		l, err = fmt.Fprintf(w, " %v ", le.Rest)
	}
	length += l
	if err != nil {
		return length, err
	}
	l, err = sxeval.PrintFrontLast(w, le.Front, le.Last)
	length += l
	return length, err
}

// Procedure represents the procedure definition form (aka lambda).
type Procedure struct {
	PFrame *sxeval.ParseFrame
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Front  []sxeval.Expr // all expressions, but the last
	Last   sxeval.Expr
}

func (p *Procedure) IsNil() bool  { return p == nil }
func (p *Procedure) IsAtom() bool { return p == nil }
func (p *Procedure) IsEql(other sx.Object) bool {
	if p == other {
		return true
	}
	if p.IsNil() {
		return sx.IsNil(other)
	}
	if otherF, ok := other.(*Procedure); ok {
		if p.Name != otherF.Name || p.Rest != otherF.Rest || p.Last != otherF.Last {
			return false
		}
		if len(p.Params) != len(otherF.Params) || len(p.Front) != len(otherF.Front) {
			return false
		}
		for i, p := range p.Params {
			if p != otherF.Params[i] {
				return false
			}
		}
		for i, e := range p.Front {
			if e != otherF.Front[i] {
				return false
			}
		}
		return p.PFrame.IsEql(otherF.PFrame)
	}
	return false
}
func (p *Procedure) IsEqual(other sx.Object) bool { return p.IsEql(other) }
func (p *Procedure) String() string               { return p.Repr() }
func (p *Procedure) Repr() string                 { return sx.Repr(p) }
func (p *Procedure) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<lambda:", p.Name, ">")
}
func (p *Procedure) Call(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	numParams := len(p.Params)
	if len(args) < numParams {
		return nil, fmt.Errorf("%s: missing arguments: %v", p.Name, p.Params[len(args):])
	}
	envSize := numParams
	if p.Rest != nil {
		envSize++
	}
	childFrame := frame.UpdateChildFrame(p.PFrame, p.Name, envSize)
	for i, p := range p.Params {
		err := childFrame.Bind(p, args[i])
		if err != nil {
			return nil, err
		}
	}
	if p.Rest != nil {
		err := childFrame.Bind(p.Rest, sx.MakeList(args[numParams:]...))
		if err != nil {
			return nil, err
		}
	} else if len(args) > numParams {
		return nil, fmt.Errorf("%s: excess arguments: %v", p.Name, args[numParams:])
	}
	for _, e := range p.Front {
		_, err := childFrame.Execute(e)
		if err != nil {
			return nil, err
		}
	}
	return childFrame.ExecuteTCO(p.Last)
}
