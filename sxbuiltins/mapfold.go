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
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// ----- Notes
//
// (map fn lst ...)
//    works on lists;
//    shortest list limits result;
//    if one list is improper, all lists must have the same shape
//
// (seq-map fn seq ...)
//    works on sequences;
//    shortest sequence limits result (typically a list);
//    improper lists are ignored
//
// (for-each fn seq ...)
//    called for the side effect, returns some undefined value
//
// (filter pred seq)
//    result is typically a list
//
// (map-filter fn-pred lst ...)
//    fn-pred returns cons (val, bool), where bool==nil states to include val into result;
//    cons can be re-used for building the result;
//    no additional allocs compared to (map fn ...)
//
// -----

// Map returns a list, where all member are the result of the given function
// to all original list members.
var Map = sxeval.Builtin{
	Name:     "map",
	MinArity: 2,
	MaxArity: 2,
	TestPure: func(args sx.Vector) bool {
		// fn must be checked first, because it is an error, if argument 0 is
		// not a callable, even if the list is empty and fn will never be called.
		fn, err := GetCallable(args[0], 0)
		if err != nil {
			return false
		}
		lst, err := GetList(args[1], 1)
		if err != nil {
			return false
		}
		if sx.IsNil(lst) {
			return true
		}

		for {
			if !fn.IsPure(sx.Vector{lst.Car()}) {
				return false
			}
			cdr := lst.Cdr()
			if sx.IsNil(cdr) {
				return true
			}
			pair, isPair := sx.GetPair(cdr)
			if !isPair {
				return fn.IsPure(sx.Vector{cdr})
			}
			lst = pair
		}
	},
	Fn: func(env *sxeval.Environment, _ int, bind *sxeval.Binding) error {
		// fn must be checked first, because it is an error, if argument 0 is
		// not a callable, even if the list is empty and fn will never be called.
		arg1 := env.Pop()
		fn, err := GetCallable(env.Pop(), 0)
		if err != nil {
			return err
		}
		lst, err := GetList(arg1, 1)
		if err != nil {
			return err
		}
		if sx.IsNil(lst) {
			env.Push(sx.Nil())
			return nil
		}

		env.Push(lst.Car())
		if err = env.Apply(fn, 1, bind); err != nil {
			return err
		}
		result := sx.Cons(env.Pop(), sx.Nil())
		curr := result
		for {
			cdr2 := lst.Cdr()
			if sx.IsNil(cdr2) {
				break
			}
			pair, isPair := sx.GetPair(cdr2)
			if !isPair {
				env.Push(cdr2)
				if err2 := env.Apply(fn, 1, bind); err2 != nil {
					return err2
				}
				curr.SetCdr(env.Pop())
				break
			}
			env.Push(pair.Car())
			if err2 := env.Apply(fn, 1, bind); err2 != nil {
				return err2
			}
			curr = curr.AppendBang(env.Pop())
			lst = pair
		}
		env.Push(result)
		return nil
	},
}

// Apply calls the given function with the given arguments.
var Apply = sxeval.Builtin{
	Name:     "apply",
	MinArity: 2,
	MaxArity: 2,
	TestPure: nil, // Might be changed in the future
	Fn: func(env *sxeval.Environment, _ int, bind *sxeval.Binding) error {
		arg1 := env.Pop()
		fn, err := GetCallable(env.Pop(), 0)
		if err != nil {
			return err
		}
		lst, err := GetList(arg1, 1)
		if err != nil {
			return err
		}
		if lst == nil {
			return env.Apply(fn, 0, bind)
		}

		env.Push(lst.Car())
		argCount := 1
		for node := lst; ; {
			cdr := node.Cdr()
			if sx.IsNil(cdr) {
				return env.Apply(fn, argCount, bind)
			}
			if next, isPair := sx.GetPair(cdr); isPair {
				node = next
				env.Push(node.Car())
				argCount++
				continue
			}
			env.Kill(argCount)
			return sx.ErrImproper{Pair: lst}
		}
	},
}

// Fold will apply the given function pairwise to list of args.
var Fold = sxeval.Builtin{
	Name:     "fold",
	MinArity: 3,
	MaxArity: 3,
	TestPure: nil, // Might be changed in the future
	Fn: func(env *sxeval.Environment, _ int, bind *sxeval.Binding) error {
		arg2 := env.Pop()
		res := env.Pop()
		fn, err := GetCallable(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		lst, err := GetList(arg2, 2)
		if err != nil {
			env.Kill(1)
			return err
		}
		for node := lst; node != nil; {
			env.Push(node.Car())
			env.Push(res)
			if err = env.Apply(fn, 2, bind); err != nil {
				env.Kill(1)
				return err
			}
			res = env.Pop()
			next, ok := sx.GetPair(node.Cdr())
			if !ok {
				env.Kill(1)
				return sx.ErrImproper{Pair: lst}
			}
			node = next
		}
		env.Set(res)
		return nil
	},
}

// FoldReverse will apply the given function reversed pairwise to reversed list of args.
var FoldReverse = sxeval.Builtin{
	Name:     "fold-reverse",
	MinArity: 3,
	MaxArity: 3,
	TestPure: nil, // Might be changed in the future
	Fn: func(env *sxeval.Environment, _ int, bind *sxeval.Binding) error {
		arg2 := env.Pop()
		res := env.Pop()
		fn, err := GetCallable(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		lst, err := GetList(arg2, 2)
		if err != nil {
			env.Kill(1)
			return err
		}
		rev, err := lst.Reverse()
		if err != nil {
			env.Kill(1)
			return err
		}
		for node := rev; node != nil; {
			env.Push(node.Car())
			env.Push(res)
			if err = env.Apply(fn, 2, bind); err != nil {
				env.Kill(1)
				return err
			}
			res = env.Pop()
			node, _ = sx.GetPair(node.Cdr())
		}
		env.Set(res)
		return nil
	},
}
