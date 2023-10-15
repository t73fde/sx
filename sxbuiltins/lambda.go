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

// CallablePold returns True, if the given argument is a callable.
func CallablePold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sxeval.GetCallable(args[0])
	return sx.MakeBoolean(ok), nil
}

// DefunS parses a procedure/function specfication and assigns it to a value.
func DefunS(frame *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	sym, le, err := parseDefProc(frame, args)
	if err != nil {
		return nil, err
	}
	return &DefineExpr{Sym: sym, Val: le}, nil
}

func parseDefProc(frame *sxeval.ParseFrame, args *sx.Pair) (*sx.Symbol, *LambdaExpr, error) {
	if args == nil {
		return nil, nil, sxeval.ErrNoArgs
	}
	sym, isSymbol := sx.GetSymbol(args.Car())
	if !isSymbol {
		return nil, nil, fmt.Errorf("not a symbol: %T/%v", args.Car(), args.Car())
	}
	args = args.Tail()
	if args == nil {
		return nil, nil, fmt.Errorf("parameter spec and body missing")
	}
	le, err := ParseProcedure(frame, sx.Repr(sym), args.Car(), args.Cdr())
	return sym, le, err
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
	es, err := ParseExprSeq(fnFrame, body)
	if err != nil {
		return nil, err
	}
	front, last := splitToFrontLast(es)
	fn := &LambdaExpr{
		Name:    name,
		Params:  params,
		Rest:    rest,
		Front:   front,
		Last:    last,
		IsMacro: false,
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
	sym, isSymbol := sx.GetSymbol(obj)
	if !isSymbol {
		return nil, fmt.Errorf("symbol in list expected, but got %T/%v", obj, obj)
	}
	for _, p := range params {
		if sym.IsEqual(p) {
			return nil, fmt.Errorf("symbol %v already defined", sym)
		}
	}
	return sym, nil
}

func splitToFrontLast(seq []sxeval.Expr) (front []sxeval.Expr, last sxeval.Expr) {
	switch l := len(seq); l {
	case 0:
		return nil, nil
	case 1:
		return nil, seq[0]
	default:
		return seq[0 : l-1], seq[l-1]
	}
}

type LambdaExpr struct {
	Name    string
	Params  []*sx.Symbol
	Rest    *sx.Symbol
	Front   []sxeval.Expr
	Last    sxeval.Expr
	IsMacro bool
}

func (le *LambdaExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	envSize := len(le.Params)
	if le.Rest != nil {
		envSize++
	}
	fnFrame := rf.MakeChildFrame(le.Name+"-rework", envSize)
	for _, sym := range le.Params {
		fnFrame.Bind(sym)
	}
	if rest := le.Rest; rest != nil {
		fnFrame.Bind(rest)
	}

	for i, e := range le.Front {
		le.Front[i] = e.Rework(fnFrame)
	}
	le.Last = le.Last.Rework(fnFrame)
	return le
}
func (le *LambdaExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	if le.IsMacro {
		return &Macro{
			Frame:  frame,
			PFrame: frame.MakeParseFrame(),
			Name:   le.Name,
			Params: le.Params,
			Rest:   le.Rest,
			Front:  le.Front,
			Last:   le.Last,
		}, nil
	}
	return &Procedure{
		PFrame: frame.MakeParseFrame(),
		Name:   le.Name,
		Params: le.Params,
		Rest:   le.Rest,
		Front:  le.Front,
		Last:   le.Last,
	}, nil
}
func (le *LambdaExpr) IsEqual(other sxeval.Expr) bool {
	if le == other {
		return true
	}
	if otherL, ok := other.(*LambdaExpr); ok && otherL != nil {
		return le.Name == otherL.Name &&
			sxeval.EqualSymbolSlice(le.Params, otherL.Params) &&
			le.Rest.IsEqual(otherL.Rest) &&
			sxeval.EqualExprSlice(le.Front, otherL.Front) &&
			le.Last.IsEqual(otherL.Last)
	}
	return false
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

	l, err = sxeval.PrintExprs(w, le.Front)
	length += l
	if err != nil {
		return length, err
	}
	l, err = le.Last.Print(w)
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
	PFrame *sxeval.ParseFrame
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Front  []sxeval.Expr
	Last   sxeval.Expr
}

func (p *Procedure) IsNil() bool  { return p == nil }
func (p *Procedure) IsAtom() bool { return p == nil }
func (p *Procedure) IsEqual(other sx.Object) bool {
	if p == other {
		return true
	}
	if p.IsNil() {
		return sx.IsNil(other)
	}
	if otherP, ok := other.(*Procedure); ok {
		// Don't compare Name, because they are always different, but that does not matter.
		return p.PFrame.IsEqual(otherP.PFrame) &&
			sxeval.EqualSymbolSlice(p.Params, otherP.Params) &&
			p.Rest.IsEqual(otherP.Rest) &&
			sxeval.EqualExprSlice(p.Front, otherP.Front) &&
			p.Last.IsEqual(otherP.Last)
	}
	return false
}
func (p *Procedure) String() string { return p.Repr() }
func (p *Procedure) Repr() string   { return sx.Repr(p) }
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
	lambdaFrame := frame.MakeLambdaFrame(p.PFrame, p.Name, envSize)
	for i, p := range p.Params {
		err := lambdaFrame.Bind(p, args[i])
		if err != nil {
			return nil, err
		}
	}
	if p.Rest != nil {
		err := lambdaFrame.Bind(p.Rest, sx.MakeList(args[numParams:]...))
		if err != nil {
			return nil, err
		}
	} else if len(args) > numParams {
		return nil, fmt.Errorf("%s: excess arguments: %v", p.Name, args[numParams:])
	}

	for _, e := range p.Front {
		subFrame := lambdaFrame.MakeCalleeFrame()
		_, err := subFrame.Execute(e)
		if err != nil {
			return nil, err
		}
	}
	return lambdaFrame.ExecuteTCO(p.Last)
}
