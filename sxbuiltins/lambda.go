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

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// CallableP returns true, if the given argument is a callable.
var CallableP = sxeval.Builtin{
	Name:     "callable?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		_, ok := sxeval.GetCallable(env.Top())
		env.Set(sx.MakeBoolean(ok))
		return nil
	},
}

// DefunS parses a procedure/function specfication.
var DefunS = sxeval.Special{
	Name: "defun",
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (sxeval.Expr, error) {
		sym, le, err := parseDefProc(pe, args, bind)
		if err != nil {
			return nil, err
		}
		return &DefineExpr{Sym: sym, Val: le}, nil
	},
}

var errNoParameterSpecAndBody = errors.New("parameter spec and body missing")

func parseDefProc(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (*sx.Symbol, *LambdaExpr, error) {
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
	le, err := ParseProcedure(pe, sym.String(), args.Car(), args.Cdr(), bind)
	return sym, le, err
}

// LambdaS parses a procedure specification.
var LambdaS = sxeval.Special{
	Name: lambdaName,
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (sxeval.Expr, error) {
		if args == nil {
			return nil, errNoParameterSpecAndBody
		}
		car := args.Car()
		var name string
		if sName, isString := sx.GetString(car); isString {
			name = sName.GetValue()
			args = args.Tail()
			if args == nil {
				return nil, errNoParameterSpecAndBody
			}
			car = args.Car()
		} else {
			name = car.String()
		}
		return ParseProcedure(pe, name, car, args.Cdr(), bind)
	},
}

// ParseProcedure parses a procedure definition, where some parsing is already done.
func ParseProcedure(pe *sxeval.ParseEnvironment, name string, paramSpec, bodySpec sx.Object, bind *sxeval.Binding) (*LambdaExpr, error) {
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
	lambdaBind := bind.MakeChildBinding(name+"-def", bindSize)
	for _, p := range params {
		err := lambdaBind.Bind(p, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	if rest != nil {
		err := lambdaBind.Bind(rest, sx.MakeUndefined())
		if err != nil {
			return nil, err
		}
	}
	expr, err := ParseExprSeq(pe, body, lambdaBind)
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

// GetParameterSymbol parses a symbol to be used in the parameter list.
// It is an error if the symbol occurs more than one.
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

// LambdaExpr stores all data for a lambda expression. It may be a macro,
// a lexical or a dynamic lambda.
type LambdaExpr struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Expr   sxeval.Expr
	Type   int // <0: Macro, =0: LexLambda, >0: DynLambda
}

const (
	macroType = -1
	macroName = "macro"

	lexLambdaType = 0
	lambdaName    = "lambda"

	dynLambdaType = 1
	dynLambdaName = "dyn-lambda"
)

// IsPure signals an expression that has no side effects.
func (*LambdaExpr) IsPure() bool { return false }

// Unparse the expression as an sx.Object
func (le *LambdaExpr) Unparse() sx.Object {
	expr := le.Expr.Unparse()
	var params sx.Object = le.Rest
	for i := len(le.Params) - 1; i >= 0; i-- {
		params = sx.Cons(le.Params[i], params)
	}
	name := lambdaName
	if le.Type < 0 {
		name = macroName
	} else if le.Type > 0 {
		name = dynLambdaName
	}
	return sx.MakeList(sx.MakeSymbol(name), params, expr)
}

// Improve the expression into a possible simpler one.
func (le *LambdaExpr) Improve(imp *sxeval.Improver) (sxeval.Expr, error) {
	bindSize := len(le.Params)
	if le.Rest != nil {
		bindSize++
	}
	lambdaImp := imp.MakeChildImprover(le.Name+"-improve", bindSize)
	for _, sym := range le.Params {
		_ = lambdaImp.Bind(sym)
	}
	if rest := le.Rest; rest != nil {
		_ = lambdaImp.Bind(rest)
	}

	expr, err := lambdaImp.Improve(le.Expr)
	if err == nil {
		le.Expr = expr
	}
	return le, err
}

// Compute the expression in a frame and return the result.
func (le *LambdaExpr) Compute(env *sxeval.Environment, bind *sxeval.Binding) (sx.Object, error) {
	leType := le.Type
	if leType == 0 {
		return &LexLambda{
			Binding: bind,
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
		Binding: bind,
		Name:    le.Name,
		Params:  le.Params,
		Rest:    le.Rest,
		Expr:    le.Expr,
	}, nil
}

// Print the expression on the given writer.
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

// IsNil returns true if the object must be treated like a sx.Nil() object.
func (ll *LexLambda) IsNil() bool { return ll == nil }

// IsAtom returns true if the object is atomic.
func (ll *LexLambda) IsAtom() bool { return ll == nil }

// IsEqual returns true if the other object has the same content.
func (ll *LexLambda) IsEqual(other sx.Object) bool { return ll == other }

// String returns a string representation.
func (ll *LexLambda) String() string { return "#<lambda:" + ll.Name + ">" }

// GoString returns a string representation to be used in Go code.
func (ll *LexLambda) GoString() string { return ll.String() }

// --- Builtin methods to implement sxeval.Callable

// IsPure tests if the Procedure needs an environment value and does not
// produce any other side effects.
func (ll *LexLambda) IsPure(sx.Vector) bool { return false }

// ExecuteCall the Procedure with any number of arguments.
func (ll *LexLambda) ExecuteCall(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
	numParams := len(ll.Params)
	if numargs < numParams {
		env.Kill(numargs)
		return fmt.Errorf("%s: missing arguments: %v", ll.Name, ll.Params[numargs:])
	}
	bindSize := numParams
	if ll.Rest != nil {
		bindSize++
	}
	lexBind := ll.Binding.MakeChildBinding(ll.Name, bindSize)
	args := env.Args(numargs)
	for i, p := range ll.Params {
		err := lexBind.Bind(p, args[i])
		if err != nil {
			env.Kill(numargs)
			return err
		}
	}
	if ll.Rest != nil {
		err := lexBind.Bind(ll.Rest, sx.MakeList(args[numParams:]...))
		if err != nil {
			env.Kill(numargs)
			return err
		}
	} else if numargs > numParams {
		env.Kill(numargs)
		return fmt.Errorf("%s: excess arguments: %v", ll.Name, []sx.Object(args[numParams:]))
	}
	env.Kill(numargs)
	return env.ExecuteTCO(ll.Expr, lexBind)
}

// DefDynS parses a procedure definition with dynamic binding.
var DefDynS = sxeval.Special{
	Name: "defdyn",
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (sxeval.Expr, error) {
		sym, le, err := parseDefProc(pe, args, bind)
		if err != nil {
			return nil, err
		}
		le.Type = dynLambdaType
		return &DefineExpr{Sym: sym, Val: le}, nil
	},
}

// DynLambdaS parses a dynamically scoped procedure specification.
var DynLambdaS = sxeval.Special{
	Name: dynLambdaName,
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (sxeval.Expr, error) {
		expr, err := LambdaS.Fn(pe, args, bind)
		if err == nil {
			le := expr.(*LambdaExpr)
			le.Type = dynLambdaType
			return le, nil
		}
		return expr, err
	},
}

// DynLambda represents the dynamic binding procedure definition form.
type DynLambda struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Expr   sxeval.Expr
}

// IsNil returns true if the object must be treated like a sx.Nil() object.
func (dl *DynLambda) IsNil() bool { return dl == nil }

// IsAtom returns true if the object is atomic.
func (dl *DynLambda) IsAtom() bool { return dl == nil }

// IsEqual returns true if the other object has the same content.
func (dl *DynLambda) IsEqual(other sx.Object) bool { return dl == other }

// String returns a string representation.
func (dl *DynLambda) String() string { return "#<dyn-lambda:" + dl.Name + ">" }

// GoString returns a string representation to be used in Go code.
func (dl *DynLambda) GoString() string { return dl.String() }

// --- Builtin methods to implement sxeval.Callable

// IsPure tests if the Procedure needs an environment value and does not
// produce any other side effects.
func (dl *DynLambda) IsPure(sx.Vector) bool { return false }

// ExecuteCall the Procedure with any number of arguments.
func (dl *DynLambda) ExecuteCall(env *sxeval.Environment, numargs int, bind *sxeval.Binding) error {
	// A DynLambda is just a LexLambda with a different Binding.
	return (&LexLambda{
		Binding: bind,
		Name:    dl.Name,
		Params:  dl.Params,
		Rest:    dl.Rest,
		Expr:    dl.Expr,
	}).ExecuteCall(env, numargs, bind)
}
