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

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// Cons returns a cons pair of the two arguments.
var Cons = sxeval.Builtin{
	Name:     "cons",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		env.Set(sx.Cons(env.Top(), arg1))
		return nil
	},
}

// PairP returns true if the argument is a pair.
var PairP = sxeval.Builtin{
	Name:     "pair?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		arg := env.Top()
		if sx.IsNil(arg) {
			return nil
		}
		_, isPair := sx.GetPair(arg)
		env.Set(sx.MakeBoolean(isPair))
		return nil
	},
}

// NullP returns true if the argument is nil.
var NullP = sxeval.Builtin{
	Name:     "null?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Set(sx.MakeBoolean(sx.IsNil(env.Top())))
		return nil
	},
}

// ListP returns true if the argument is a (proper) list.
var ListP = sxeval.Builtin{
	Name:     "list?",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Set(sx.MakeBoolean(sx.IsList(env.Top())))
		return nil
	},
}

// Car returns the car of a pair argument.
var Car = sxeval.Builtin{
	Name:     "car",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		pair, err := GetPair(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(pair.Car())
		return nil
	},
}

// Cdr returns the car of a pair argument.
var Cdr = sxeval.Builtin{
	Name:     "cdr",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		pair, err := GetPair(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(pair.Cdr())
		return nil
	},
}

func makeCxr(spec string) sxeval.Builtin {
	return sxeval.Builtin{
		Name:     "c" + spec + "r",
		MinArity: 1,
		MaxArity: 1,
		TestPure: sxeval.AssertPure,
		Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
			pair, err := GetPair(env.Top(), 0)
			if err != nil {
				env.Kill(1)
				return err
			}
			var result sx.Object
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
					env.Kill(1)
					return fmt.Errorf("pair expected, but got %T/%v", result, result)
				}
			}
			env.Set(result)
			return nil
		},
	}
}

// Car/Cdr call, two to four levels deep.
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
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		lst, err := GetList(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		obj, err := lst.Last()
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(obj)
		return nil
	},
}

// List returns a list of all arguments.
var List = sxeval.Builtin{
	Name:     sx.SymbolList.String(),
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Push(sx.Nil())
		return nil
	},
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Set(sx.Cons(env.Top(), sx.Nil()))
		return nil
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		obj := sx.MakeList(env.Args(numargs)...)
		env.Kill(numargs - 1)
		env.Set(obj)
		return nil
	},
}

// ListStar returns a list of all arguments, where the last argument is a cons to the second last.
var ListStar = sxeval.Builtin{
	Name:     "list*",
	MinArity: 1,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn1:      func(_ *sxeval.Environment, _ *sxeval.Binding) error { return nil },
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		argPos := len(args) - 2
		result := sx.Cons(args[argPos], args[argPos+1])
		for argPos > 0 {
			argPos--
			result = sx.Cons(args[argPos], result)
		}
		env.Kill(numargs - 1)
		env.Set(result)
		return nil
	},
}

// Append returns a list where all list arguments are concatenated.
var Append = sxeval.Builtin{
	Name:     "append",
	MinArity: 0,
	MaxArity: -1,
	TestPure: sxeval.AssertPure,
	Fn0: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		env.Push(sx.Nil())
		return nil
	},
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		lst, err := GetList(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(lst)
		return nil
	},
	Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
		args := env.Args(numargs)
		lastList := len(args) - 1
		lsts := make([]*sx.Pair, lastList)
		for i := range lastList {
			lst, err := GetList(args[i], i)
			if err != nil {
				env.Kill(numargs)
				return err
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
					env.Kill(numargs)
					return sx.ErrImproper{Pair: lst}
				}
				node = next
			}
		}
		curr.SetCdr(args[lastList])
		env.Kill(numargs - 1)
		env.Set(sentinel.Cdr())
		return nil
	},
}

// Reverse returns a reversed list.
var Reverse = sxeval.Builtin{
	Name:     "reverse",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		lst, err := GetList(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		obj, err := lst.Reverse()
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(obj)
		return nil
	},
}

// Assoc returns the first pair of the a-list where the second argument is
// equal (e.g. '=) to the pair's car. Otherwise, nil is returned.
var Assoc = sxeval.Builtin{
	Name:     "assoc",
	MinArity: 2,
	MaxArity: 2,
	TestPure: sxeval.AssertPure,
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		lst, err := GetList(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(lst.Assoc(arg1))
		return nil
	},
}

// All returns a true value, if all elements of the list evaluate to true.
// Otherwise it returns a false value.
var All = sxeval.Builtin{
	Name:     "all",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		obj, err := anyAll(env.Top(), sx.IsFalse, false)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(obj)
		return nil
	},
}

// Any returns a true value, if any element of the list evaluates to true.
// Otherwise it returns a false value.
var Any = sxeval.Builtin{
	Name:     "any",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		obj, err := anyAll(env.Top(), sx.IsTrue, true)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(obj)
		return nil
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
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		lst, err := GetList(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(sx.Collect(lst.Values()))
		return nil
	},
}
