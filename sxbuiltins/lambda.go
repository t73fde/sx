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
	"errors"
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
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
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

var errNoParameterSpecAndBody = errors.New("parameter spec and body missing")

func parseDefProc(pf *sxeval.ParseEnvironment, args *sx.Pair) (*sx.Symbol, *LambdaExpr, error) {
	if args == nil {
		return nil, nil, sxeval.ErrNoArgs
	}
	sym, isSymbol := sx.GetSymbol(args.Car())
	if !isSymbol {
		return nil, nil, fmt.Errorf("not a symbol: %T/%v", args.Car(), args.Car())
	}
	args = args.Tail()
	if args == nil {
		return nil, nil, errNoParameterSpecAndBody
	}
	le, err := ParseProcedure(pf, sym.String(), args.Car(), args.Cdr())
	return sym, le, err
}

const lambdaName = "lambda"

// LambdaS parses a procedure specification.
var LambdaS = sxeval.Special{
	Name: lambdaName,
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		if args == nil {
			return nil, errNoParameterSpecAndBody
		}
		car := args.Car()
		return ParseProcedure(pf, car.String(), car, args.Cdr())
	},
}

// ParseProcedure parses a procedure definition, where some parsing is already done.
func ParseProcedure(pf *sxeval.ParseEnvironment, name string, paramSpec, bodySpec sx.Object) (*LambdaExpr, error) {
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
	if rest != nil {
		bindSize++
	}
	fnFrame := pf.MakeChildFrame(name+"-def", bindSize)
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
	expr, err := ParseExprSeq(fnFrame, body)
	if err != nil {
		return nil, err
	}
	fn := &LambdaExpr{
		Name:   name,
		Params: params,
		Rest:   rest,
		Expr:   expr,
		Type:   lexLambdaType,
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

type LambdaExpr struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Expr   sxeval.Expr
	Type   int // <0: Macro, =0: LexLambda, >0: DynLambda
}

const (
	macroType     = -1
	lexLambdaType = 0
	dynLambdaType = 1
)

func (le *LambdaExpr) Unparse() sx.Object {
	expr := le.Expr.Unparse()
	var params sx.Object = le.Rest
	for i := len(le.Params) - 1; i >= 0; i-- {
		params = sx.Cons(le.Params[i], params)
	}
	return sx.MakeList(sx.MakeSymbol(lambdaName), params, expr)
}

func (le *LambdaExpr) Rework(re *sxeval.ReworkEnvironment) sxeval.Expr {
	bindSize := len(le.Params)
	if le.Rest != nil {
		bindSize++
	}
	fnFrame := re.MakeChildFrame(le.Name+"-rework", bindSize)
	for _, sym := range le.Params {
		_ = fnFrame.Bind(sym)
	}
	if rest := le.Rest; rest != nil {
		_ = fnFrame.Bind(rest)
	}

	le.Expr = le.Expr.Rework(fnFrame)
	return le
}

func (le *LambdaExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	leType := le.Type
	if leType == 0 {
		return &LexLambda{
			Binding: env.Binding(),
			Name:    le.Name,
			Params:  le.Params,
			Rest:    le.Rest,
			Expr:    le.Expr,
		}, nil
	}
	if leType > 0 {
		return &DynLambda{
			Name:   le.Name,
			Params: le.Params,
			Rest:   le.Rest,
			Expr:   le.Expr,
		}, nil
	}
	return &Macro{
		Env:     env,
		Binding: env.Binding(),
		Name:    le.Name,
		Params:  le.Params,
		Rest:    le.Rest,
		Expr:    le.Expr,
	}, nil
}

func (le *LambdaExpr) Print(w io.Writer) (int, error) {
	var typeString string
	if leType := le.Type; leType < 0 {
		typeString = "MACRO-"
	} else if leType > 0 {
		typeString = "DYN-"
	}
	length, err := fmt.Fprintf(w, "{%sLAMBDA %q ", typeString, le.Name)
	if err != nil {
		return length, err
	}
	var l int
	if params, rest := le.Params, le.Rest; len(params) == 0 && rest != nil {
		l, err = fmt.Fprintf(w, "%v ", le.Rest)
	} else {
		l, err = io.WriteString(w, "(")
		length += l
		if err != nil {
			return length, err
		}
		for i, p := range le.Params {
			if i == 0 {
				l, err = fmt.Fprintf(w, "%v", p)
			} else {
				l, err = fmt.Fprintf(w, " %v", p)
			}
			length += l
			if err != nil {
				return length, err
			}
		}
		if rest != nil {
			l, err = fmt.Fprintf(w, ". %v) ", rest)
		} else {
			l, err = io.WriteString(w, ") ")
		}
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

// LexLambda represents the lexical procedure definition form.
type LexLambda struct {
	Binding *sxeval.Binding
	Name    string
	Params  []*sx.Symbol
	Rest    *sx.Symbol
	Expr    sxeval.Expr
}

func (ll *LexLambda) IsNil() bool                  { return ll == nil }
func (ll *LexLambda) IsAtom() bool                 { return ll == nil }
func (ll *LexLambda) IsEqual(other sx.Object) bool { return ll == other }
func (ll *LexLambda) String() string               { return "#<lambda:" + ll.Name + ">" }
func (ll *LexLambda) GoString() string             { return ll.String() }

// --- Builtin methods to implement sxeval.Callable

// IsPure tests if the Procedure needs an environment value and does not
// produce any other side effects.
func (ll *LexLambda) IsPure(sx.Vector) bool { return false }

// Call the Procedure.
func (ll *LexLambda) Call(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
	numParams := len(ll.Params)
	if len(args) < numParams {
		return nil, fmt.Errorf("%s: missing arguments: %v", ll.Name, ll.Params[len(args):])
	}
	bindSize := numParams
	if ll.Rest != nil {
		bindSize++
	}
	lexicalEnv := env.NewLexicalEnvironment(ll.Binding, ll.Name, bindSize)
	for i, p := range ll.Params {
		err := lexicalEnv.Bind(p, args[i])
		if err != nil {
			return nil, err
		}
	}
	if ll.Rest != nil {
		err := lexicalEnv.Bind(ll.Rest, sx.MakeList(args[numParams:]...))
		if err != nil {
			return nil, err
		}
	} else if len(args) > numParams {
		return nil, fmt.Errorf("%s: excess arguments: %v", ll.Name, []sx.Object(args[numParams:]))
	}
	return lexicalEnv.ExecuteTCO(ll.Expr)
}

// DefDynS parses a procedure definition with dynamic binding.
var DefDynS = sxeval.Special{
	Name: "defdyn",
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		sym, le, err := parseDefProc(pf, args)
		if err != nil {
			return nil, err
		}
		le.Type = dynLambdaType
		return &DefineExpr{Sym: sym, Val: le}, nil
	},
}

// DynLambda represents the dynamic binding procedure definition form.
type DynLambda struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Expr   sxeval.Expr
}

func (dl *DynLambda) IsNil() bool                  { return dl == nil }
func (dl *DynLambda) IsAtom() bool                 { return dl == nil }
func (dl *DynLambda) IsEqual(other sx.Object) bool { return dl == other }
func (dl *DynLambda) String() string               { return "#<dyn-lambda:" + dl.Name + ">" }
func (dl *DynLambda) GoString() string             { return dl.String() }

// --- Builtin methods to implement sxeval.Callable

// IsPure tests if the Procedure needs an environment value and does not
// produce any other side effects.
func (dl *DynLambda) IsPure(sx.Vector) bool { return false }

// Call the Procedure.
func (dl *DynLambda) Call(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
	numParams := len(dl.Params)
	if len(args) < numParams {
		return nil, fmt.Errorf("%s: missing arguments: %v", dl.Name, dl.Params[len(args):])
	}
	bindSize := numParams
	if dl.Rest != nil {
		bindSize++
	}
	dynEnv := env.NewLexicalEnvironment(env.Binding(), dl.Name, bindSize)
	for i, p := range dl.Params {
		err := dynEnv.Bind(p, args[i])
		if err != nil {
			return nil, err
		}
	}
	if dl.Rest != nil {
		err := dynEnv.Bind(dl.Rest, sx.MakeList(args[numParams:]...))
		if err != nil {
			return nil, err
		}
	} else if len(args) > numParams {
		return nil, fmt.Errorf("%s: excess arguments: %v", dl.Name, []sx.Object(args[numParams:]))
	}
	return dynEnv.ExecuteTCO(dl.Expr)
}
