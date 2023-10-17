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
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		return sx.Cons(args[0], args[1]), nil
	},
}

// PairP returns true if the argument is a pair.
var PairP = sxeval.Builtin{
	Name:     "pair?",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		obj := args[0]
		if sx.IsNil(obj) {
			return sx.Nil(), nil
		}
		_, isPair := sx.GetPair(obj)
		return sx.MakeBoolean(isPair), nil
	},
}

// NullP returns true if the argument is nil.
var NullP = sxeval.Builtin{
	Name:     "null?",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		return sx.MakeBoolean(sx.IsNil(args[0])), nil
	},
}

// ListP returns true if the argument is a (proper) list.
var ListP = sxeval.Builtin{
	Name:     "list?",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		return sx.MakeBoolean(sx.IsList(args[0])), nil
	},
}

// Car returns the car of a pair argument.
var Car = sxeval.Builtin{
	Name:     "car",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		pair, err := GetPair(args, 0)
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
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		pair, err := GetPair(args, 0)
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
		IsPure:   true,
		Fn: func(_ *sxeval.Frame, args []sx.Object) (result sx.Object, _ error) {
			pair, err := GetPair(args, 0)
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
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		lst, err := GetList(args, 0)
		if err != nil {
			return nil, err
		}
		return lst.Last()
	},
}

// List returns a list of all arguments.
var List = sxeval.Builtin{
	Name:     sx.ListName,
	MinArity: 0,
	MaxArity: -1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		return sx.MakeList(args...), nil
	},
}

// ListStar returns a list of all arguments, where the last argument is a cons to the second last.
var ListStar = sxeval.Builtin{
	Name:     "list*",
	MinArity: 1,
	MaxArity: -1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		if len(args) == 1 {
			return args[0], nil
		}
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
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		switch len(args) {
		case 0:
			return sx.Nil(), nil
		case 1:
			return args[0], nil
		}
		lastList := len(args) - 1
		lsts := make([]*sx.Pair, lastList)
		for i := 0; i < lastList; i++ {
			lst, err := GetList(args, i)
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
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		lst, err := GetList(args, 0)
		if err != nil {
			return nil, err
		}
		return lst.Reverse()
	},
}

// Length returns the length of the given list.
var Length = sxeval.Builtin{
	Name:     "length",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		lst, err := GetList(args, 0)
		if err != nil {
			return nil, err
		}
		return sx.Int64(int64(lst.Length())), nil
	},
}

// Assoc returns the first pair of the a-list where the second argument is
// equal (e.g. '=) to the pair's car. Otherwise, nil is returned.
var Assoc = sxeval.Builtin{
	Name:     "assoc",
	MinArity: 2,
	MaxArity: 2,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		lst, err := GetList(args, 0)
		if err != nil {
			return nil, err
		}
		return lst.Assoc(args[1]), nil
	},
}
