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

// Contains all list-related builtins

import (
	"fmt"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Cons returns a cons pair of the two arguments.
var Cons = sxeval.Builtin{
	Name:     "cons",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		return sx.Cons(arg0, arg1), nil
	},
}

// PairP returns true if the argument is a pair.
var PairP = sxeval.Builtin{
	Name:     "pair?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		if sx.IsNil(arg) {
			return sx.Nil(), nil
		}
		_, isPair := sx.GetPair(arg)
		return sx.MakeBoolean(isPair), nil
	},
}

// NullP returns true if the argument is nil.
var NullP = sxeval.Builtin{
	Name:     "null?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return sx.MakeBoolean(sx.IsNil(arg)), nil
	},
}

// ListP returns true if the argument is a (proper) list.
var ListP = sxeval.Builtin{
	Name:     "list?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return sx.MakeBoolean(sx.IsList(arg)), nil
	},
}

// Car returns the car of a pair argument.
var Car = sxeval.Builtin{
	Name:     "car",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		pair, err := GetPair(arg, 0)
		if err != nil {
			return nil, err
		}
		return pair.Car(), nil
	},
}

// Cdr returns the car of a pair argument.
var Cdr = sxeval.Builtin{
	Name:     "cdr",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		pair, err := GetPair(arg, 0)
		if err != nil {
			return nil, err
		}
		return pair.Cdr(), nil
	},
}

func makeCxr(spec string) sxeval.Builtin {
	return sxeval.Builtin{
		Name:     "c" + spec + "r",
		MinArity: 1,
		MaxArity: 1,
		TestPure: sxeval.AssertPure,
		Fn1: func(_ *sxeval.Environment, arg sx.Object) (result sx.Object, _ error) {
			pair, err := GetPair(arg, 0)
			if err != nil {
				return nil, err
			}
			i := len(spec) - 1
			for {
				switch spec[i] {
				case 'a':
					result = pair.Car()
				case 'd':
					result = pair.Cdr()
				default:
					panic(spec)
				}
				if i <= 0 {
					break
				}
				i--
				var isPair bool
				pair, isPair = sx.GetPair(result)
				if !isPair {
					return nil, fmt.Errorf("pair expected, but got %T/%v", result, result)
				}
			}
			return result, nil
		},
	}
}

var (
	Caar = makeCxr("aa")
	Cadr = makeCxr("ad")
	Cdar = makeCxr("da")
	Cddr = makeCxr("dd")

	Caaar = makeCxr("aaa")
	Caadr = makeCxr("aad")
	Cadar = makeCxr("ada")
	Caddr = makeCxr("add")
	Cdaar = makeCxr("daa")
	Cdadr = makeCxr("dad")
	Cddar = makeCxr("dda")
	Cdddr = makeCxr("ddd")

	Caaaar = makeCxr("aaaa")
	Caaadr = makeCxr("aaad")
	Caadar = makeCxr("aada")
	Caaddr = makeCxr("aadd")
	Cadaar = makeCxr("adaa")
	Cadadr = makeCxr("adad")
	Caddar = makeCxr("adda")
	Cadddr = makeCxr("addd")
	Cdaaar = makeCxr("daaa")
	Cdaadr = makeCxr("daad")
	Cdadar = makeCxr("dada")
	Cdaddr = makeCxr("dadd")
	Cddaar = makeCxr("ddaa")
	Cddadr = makeCxr("ddad")
	Cdddar = makeCxr("ddda")
	Cddddr = makeCxr("dddd")
)

// Last returns the last element of a list
var Last = sxeval.Builtin{
	Name:     "last",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		lst, err := GetList(arg, 0)
		if err != nil {
			return nil, err
		}
		return lst.Last()
	},
}

// List returns a list of all arguments.
var List = sxeval.Builtin{
	Name:     sx.SymbolList.String(),
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(_ *sxeval.Environment) (sx.Object, error) {
		return sx.Nil(), nil
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return sx.Cons(arg, sx.Nil()), nil
	},
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		return sx.Cons(arg0, sx.Cons(arg1, sx.Nil())), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		return sx.MakeList(args...), nil
	},
}

// ListStar returns a list of all arguments, where the last argument is a cons to the second last.
var ListStar = sxeval.Builtin{
	Name:     "list*",
	MinArity: 1,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return arg, nil
	},
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		return sx.Cons(arg0, arg1), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		argPos := len(args) - 2
		result := sx.Cons(args[argPos], args[argPos+1])
		for argPos > 0 {
			argPos--
			result = sx.Cons(args[argPos], result)
		}
		return result, nil
	},
}

// Append returns a list where all list arguments are concatenated.
var Append = sxeval.Builtin{
	Name:     "append",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(_ *sxeval.Environment) (sx.Object, error) {
		return sx.Nil(), nil
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		lst, err := GetList(arg, 0)
		return lst, err
	},
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		lst0, err := GetList(arg0, 0)
		if err != nil {
			return nil, err
		}
		lst1, err := GetList(arg1, 1)
		if err != nil {
			return nil, err
		}
		if lst0 == nil {
			return lst1, nil
		}

		result := sx.Cons(lst0.Car(), lst0.Cdr())
		prev := result
		for {
			curr := prev.Cdr()
			if next, isPair := sx.GetPair(curr); isPair {
				if next == nil {
					prev.SetCdr(lst1)
					return result, nil
				}
				copy := sx.Cons(next.Car(), next.Cdr())
				prev.SetCdr(copy)
				prev = copy
				continue
			}
			return nil, sx.ErrImproper{Pair: lst0}
		}
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		if len(args) == 0 {
			return sx.Nil(), nil
		}
		lastList := len(args) - 1
		lsts := make([]*sx.Pair, lastList)
		for i := range lastList {
			lst, err := GetList(args[i], i)
			if err != nil {
				return nil, err
			}
			lsts[i] = lst
		}
		sentinel := sx.Pair{}
		curr := &sentinel
		for _, lst := range lsts {
			for node := lst; node != nil; {
				curr = curr.AppendBang(node.Car())
				next, isPair := sx.GetPair(node.Cdr())
				if !isPair {
					return nil, sx.ErrImproper{Pair: lst}
				}
				node = next
			}
		}
		curr.SetCdr(args[lastList])
		return sentinel.Cdr(), nil
	},
}

// Reverse returns a reversed list.
var Reverse = sxeval.Builtin{
	Name:     "reverse",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		lst, err := GetList(arg, 0)
		if err != nil {
			return nil, err
		}
		return lst.Reverse()
	},
}

// Assoc returns the first pair of the a-list where the second argument is
// equal (e.g. '=) to the pair's car. Otherwise, nil is returned.
var Assoc = sxeval.Builtin{
	Name:     "assoc",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		lst, err := GetList(arg0, 0)
		if err != nil {
			return nil, err
		}
		return lst.Assoc(arg1), nil
	},
}

// All returns a true value, if all elements of the list evaluate to true.
// Otherwise it returns a false value.
var All = sxeval.Builtin{
	Name:     "all",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return anyAll(arg, sx.IsFalse, false)
	},
}

// All returns a true value, if any element of the list evaluate to true.
// Otherwise it returns a false value.
var Any = sxeval.Builtin{
	Name:     "any",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		return anyAll(arg, sx.IsTrue, true)
	},
}

// anyAll is a helper function for builtins any and all.
func anyAll(arg sx.Object, pred func(sx.Object) bool, found bool) (sx.Object, error) {
	lst, err := GetList(arg, 0)
	if err != nil {
		return nil, err
	}
	for node := lst; node != nil; {
		if pred(node.Car()) {
			return sx.MakeBoolean(found), nil
		}
		cdr := node.Cdr()
		next, isPair := sx.GetPair(cdr)
		if !isPair {
			return sx.MakeBoolean(pred(cdr) == found), nil
		}
		node = next
	}
	return sx.MakeBoolean(!found), nil
}

// List2Vector returns the given proper list as a vector.
var List2Vector = sxeval.Builtin{
	Name:     "list->vector",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
		lst, err := GetList(arg, 0)
		if err != nil {
			return nil, err
		}
		return lst.AsVector(), nil
	},
}
