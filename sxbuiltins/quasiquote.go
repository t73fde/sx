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

// Quasiquote implementation is a little bit too simple as it does not support
// nested quasiquotes.

import (
	"errors"
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// QuasiquoteS parses a form that is quasi-quotated
var QuasiquoteS = sxeval.Special{
	Name: sx.QuasiquoteSymbol.Name(),
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
		if sx.IsNil(args) {
			return nil, sxeval.ErrNoArgs
		}
		if !sx.IsNil(args.Cdr()) {
			return nil, fmt.Errorf("more than one argument: %v", args)
		}
		qqp := qqParser{
			pframe: pf,
		}
		return qqp.parseQQ(args.Car())
	},
}

// UnquoteS parses the unquote symbol (and returns an error, because it is
// not allowed outside a quasiquote).
var UnquoteS = sxeval.Special{
	Name: sx.UnquoteSymbol.Name(),
	Fn: func(*sxeval.ParseFrame, *sx.Pair) (sxeval.Expr, error) {
		return nil, errNotAllowedOutsideQQ
	},
}

// UnquoteSplicingS parses the unquote-splicing symbol (and returns an error,
// because it is not allowed outside a quasiquote).
var UnquoteSplicingS = sxeval.Special{
	Name: sx.UnquoteSplicingSymbol.Name(),
	Fn: func(*sxeval.ParseFrame, *sx.Pair) (sxeval.Expr, error) {
		return nil, errNotAllowedOutsideQQ
	},
}

var errNotAllowedOutsideQQ = errors.New("not allowed outside " + sx.QuasiquoteSymbol.Name())

type qqParser struct {
	pframe *sxeval.ParseFrame
}

func (qqp *qqParser) parse(obj sx.Object) (sxeval.Expr, error) { return qqp.pframe.Parse(obj) }

func (qqp *qqParser) parseQQ(obj sx.Object) (sxeval.Expr, error) {
	pair, isPair := sx.GetPair(obj)
	if !isPair || pair == nil {
		// `basic is the same as (quote basic), for any form basic that is not a list.
		return sxeval.ObjExpr{Obj: obj}, nil
	}
	first := pair.Car()
	if sym, isSymbol := sx.GetSymbol(first); isSymbol {
		if sx.UnquoteSymbol.IsEqual(sym) {
			form, err := getUnquoteObj(sym, pair)
			if err != nil {
				return nil, err
			}
			// `,form is the same as form, for any form.
			return qqp.parse(form)
		}
		if sx.QuasiquoteSymbol.IsEqual(sym) {
			form, err := getUnquoteObj(sym, pair)
			if err != nil {
				return nil, err
			}
			// If the backquote syntax is nested, the innermost backquoted form should be expanded first.
			// This means that if several commas occur in a row, the leftmost one belongs to the innermost backquote.
			expr, err := qqp.parseQQ(form)
			if err != nil {
				return nil, err
			}
			return listArgs([]sxeval.Expr{sxeval.ObjExpr{Obj: sym}, expr}), err
		}
		if sx.UnquoteSplicingSymbol.IsEqual(sym) {
			// `,@form has undefined consequences.
			return nil, fmt.Errorf("(%v %v) is not allowed", sx.QuasiquoteSymbol, obj)
		}
	}
	args, err := qqp.parseList(pair)
	if err != nil {
		return nil, err
	}
	return combineArgs(args), nil
}

// combineArgs optimizes some cases for (append ...).
//
// (append) --> ()
// (append x) --> x
// (append (x) (y) ...) --> (list x y ...) OR (list* x y ...), if all but the last element
// are not spliced. (list* ...) will be used, if the last element was spliced, (list ...) if not.
//
// In addition, for some sequences in (append ... (x) (y) ...), these will be simplified
// into (append ... (list x y) ...).
func combineArgs(args []sxeval.Expr) sxeval.Expr {
	if len(args) == 0 {
		return sxeval.NilExpr
	}
	if len(args) == 1 {
		return args[0]
	}
	mleCount := countMLE(args)
	if mleCount < len(args)-1 {
		newArgs := collectMakeList(args)
		return &sxeval.BuiltinCallExpr{Proc: &Append, Args: newArgs}
	}
	newArgs := make([]sxeval.Expr, len(args))
	for i := 0; i < mleCount; i++ {
		newArgs[i] = args[i].(MakeListExpr).Elem
	}
	if mleCount < len(args) {
		newArgs[mleCount] = args[mleCount]
		return listStarArgs(newArgs)
	}
	return listArgs(newArgs)
}
func countMLE(args []sxeval.Expr) int {
	for i := 0; i < len(args); i++ {
		if _, isMLE := args[i].(MakeListExpr); !isMLE {
			return i
		}
	}
	return len(args)
}
func collectMakeList(args []sxeval.Expr) []sxeval.Expr {
	result := make([]sxeval.Expr, 0, len(args))
	var makeLists []sxeval.Expr
	for _, arg := range args {
		if mle, isMLE := arg.(MakeListExpr); isMLE {
			makeLists = append(makeLists, mle.Elem)
		} else {
			if len(makeLists) > 0 {
				result = append(result, listArgs(makeLists))
			}
			makeLists = nil
			result = append(result, arg)
		}
	}
	if len(makeLists) > 0 {
		result = append(result, listArgs(makeLists))
	}
	return result
}

// listArgs optimizes some cases for (list ...).
//
// (list)           --> ()
// (list 'x 'y ...) --> '(x y ...)
//
// In addition, arguments of the form (list 'x) are transformed into '(x) before optimization.
func listArgs(args []sxeval.Expr) sxeval.Expr {
	if len(args) == 0 {
		return sxeval.NilExpr
	}
	if countQuote(args) < len(args) {
		lstArgs := resolveMakeListQuoted(args)
		return &sxeval.BuiltinCallExpr{Proc: &List, Args: lstArgs}
	}
	lstArgs := make([]sx.Object, len(args))
	for i, arg := range args {
		if oe, isObj := arg.(sxeval.ObjExpr); isObj {
			lstArgs[i] = oe.Obj
		} else {
			lstArgs[i] = sx.MakeList(arg.(MakeListExpr).Elem.(sxeval.ObjExpr).Obj)
		}
	}
	return sxeval.ObjExpr{Obj: sx.MakeList(lstArgs...)}
}
func countQuote(args []sxeval.Expr) int {
	for i, arg := range args {
		if _, isObj := arg.(sxeval.ObjExpr); isObj {
			continue
		}
		if mle, isMLE := arg.(MakeListExpr); isMLE {
			if _, isObj := mle.Elem.(sxeval.ObjExpr); isObj {
				continue
			}
		}
		return i
	}
	return len(args)
}

// resolveMakeListQuoted changes arguments.
//
// It basically transforms a (list 'x) into '(x).
// It does not work on arbitrary (list ...)-calls, but only those with one arg.
func resolveMakeListQuoted(args []sxeval.Expr) []sxeval.Expr {
	result := make([]sxeval.Expr, len(args))
	for i, arg := range args {
		if mle, isMLE := arg.(MakeListExpr); isMLE {
			if oe, isObj := mle.Elem.(sxeval.ObjExpr); isObj {
				result[i] = sxeval.ObjExpr{Obj: sx.MakeList(oe.Obj)}
				continue
			}
		}
		result[i] = arg
	}
	return result
}

// listStarArgs optimizes some cases for (list* ...).
//
// (list*)     --> ()
// (list* x)   --> x
// (list* x y) --> (cons x y)
func listStarArgs(args []sxeval.Expr) sxeval.Expr {
	switch len(args) {
	case 0:
		return sxeval.NilExpr
	case 1:
		lstArgs := resolveMakeListQuoted(args)
		return lstArgs[0]
	case 2:
		lstArgs := resolveMakeListQuoted(args)
		return &sxeval.BuiltinCallExpr{Proc: &Cons, Args: lstArgs}
	default:
		lstArgs := resolveMakeListQuoted(args)
		return &sxeval.BuiltinCallExpr{Proc: &ListStar, Args: lstArgs}
	}
}

func (qqp *qqParser) parseList(lst *sx.Pair) ([]sxeval.Expr, error) {
	length, prevPair, lastPair := analyseList(lst)
	if length == 0 {
		return nil, nil
	}
	numArgs, realArgs := length, length
	var form sxeval.Expr
	if prevPair != nil {
		if sym, isSymbol := sx.GetSymbol(prevPair.Car()); isSymbol {
			if sx.UnquoteSymbol.IsEqual(sym) {
				obj, err := getUnquoteObj(sym, prevPair)
				if err != nil {
					return nil, err
				}
				// `(x1 x2 x3 ... xn . ,form) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] form)
				expr, err := qqp.parse(obj)
				if err != nil {
					return nil, err
				}
				numArgs--
				realArgs -= 2
				form = expr
			} else if sx.UnquoteSplicingSymbol.IsEqual(sym) {
				// `(x1 x2 x3 ... xn . ,@form) has undefined consequences.
				return nil, fmt.Errorf("%v not allowed", lst)
			}
		}
	}
	if form == nil {
		last := lastPair.Cdr()
		if !sx.IsNil(last) {
			// `(x1 x2 x3 ... xn . atom) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] (quote atom))
			form = sxeval.ObjExpr{Obj: last}
			numArgs++
		}
	}

	args := make([]sxeval.Expr, numArgs)
	for node, i := lst, 0; i < realArgs; i++ {
		elem := node.Car()
		node = node.Tail()
		if elemList, isPair := sx.GetPair(elem); isPair && elemList != nil {
			if sym, isSymbol := sx.GetSymbol(elemList.Car()); isSymbol {
				if sx.UnquoteSymbol.IsEqual(sym) {
					// -- [,form] is interpreted as (list form)
					obj, err := getUnquoteObj(sym, elemList)
					if err != nil {
						return nil, err
					}
					expr, err := qqp.parse(obj)
					if err != nil {
						return nil, err
					}
					args[i] = MakeListExpr{expr}
					continue
				}
				if sx.UnquoteSplicingSymbol.IsEqual(sym) {
					// -- [,@form] is interpreted as form.
					obj, err := getUnquoteObj(sym, elemList)
					if err != nil {
						return nil, err
					}
					expr, err := qqp.parse(obj)
					if err != nil {
						return nil, err
					}
					args[i] = expr
					continue
				}
			}
		}
		// -- [form] is interpreted as (list `form), which contains a backquoted form that must then be further interpreted.
		expr, err := qqp.parseQQ(elem)
		if err != nil {
			return nil, err
		}
		args[i] = MakeListExpr{expr}
	}

	if form != nil {
		// `(x1 x2 x3 ... xn . ,form) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] form)
		// or
		// `(x1 x2 x3 ... xn . atom) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] (quote atom))
		args[realArgs] = form
	}
	return args, nil
}

func analyseList(lst *sx.Pair) (int, *sx.Pair, *sx.Pair) {
	length := 0
	prevObj, lastPair := sx.Nil(), sx.Nil()
	for node := lst; node != nil; {
		length++
		prevObj = lastPair
		lastPair = node
		next, isPair := sx.GetPair(node.Cdr())
		if !isPair {
			break
		}
		node = next
	}
	return length, prevObj, lastPair
}

func getUnquoteObj(sym sx.Symbol, lst *sx.Pair) (sx.Object, error) {
	args, isPair := sx.GetPair(lst.Cdr())
	if !isPair {
		return nil, sx.ErrImproper{Pair: lst}
	}
	if args == nil {
		return nil, fmt.Errorf("missing argument for %s", sym.Name())
	}
	obj := args.Car()
	rest := args.Cdr()
	if sx.IsNil(rest) {
		return obj, nil
	}
	return nil, fmt.Errorf("additional arguments %v for %s", rest.Repr(), sym.Name())
}

type MakeListExpr struct{ Elem sxeval.Expr }

func (mle MakeListExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	mle.Elem = mle.Elem.Rework(rf)
	return mle
}
func (mle MakeListExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	subEnv := env.NewDynamicEnvironment()
	elem, err := subEnv.Execute(mle.Elem)
	if err != nil {
		return nil, err
	}
	return sx.Cons(elem, nil), nil
}
func (mle MakeListExpr) IsEqual(other sxeval.Expr) bool {
	if mle == other {
		return true
	}
	if otherM, ok := other.(*MakeListExpr); ok && otherM != nil {
		return mle.Elem.IsEqual(otherM.Elem)
	}
	return false
}

func (mle MakeListExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{MAKELIST ")
	if err != nil {
		return length, err
	}
	l, err := mle.Elem.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err

}
