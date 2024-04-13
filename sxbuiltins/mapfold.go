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

// Map returns a list, where all member are the result of the given function
// to all original list members.
var Map = sxeval.Builtin{
	Name:     "map",
	MinArity: 2,
	MaxArity: 2,
	TestPure: func(args sx.Vector) bool {
		if len(args) < 2 {
			return false
		}
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
	Fn2: func(env *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		// fn must be checked first, because it is an error, if argument 0 is
		// not a callable, even if the list is empty and fn will never be called.
		fn, err := GetCallable(arg0, 0)
		if err != nil {
			return nil, err
		}
		lst, err := GetList(arg1, 1)
		if err != nil {
			return nil, err
		}
		if sx.IsNil(lst) {
			return sx.Nil(), nil
		}

		val, err := env.Call(fn, sx.Vector{lst.Car()})
		if err != nil {
			return sx.Nil(), err
		}
		result := sx.Cons(val, sx.Nil())
		curr := result
		for {
			cdr2 := lst.Cdr()
			if sx.IsNil(cdr2) {
				break
			}
			pair, isPair := sx.GetPair(cdr2)
			if !isPair {
				val2, err2 := env.Call(fn, sx.Vector{cdr2})
				if err2 != nil {
					return result, err2
				}
				curr.SetCdr(val2)
				break
			}
			val2, err2 := env.Call(fn, sx.Vector{pair.Car()})
			if err2 != nil {
				return result, err2
			}
			curr = curr.AppendBang(val2)
			lst = pair
		}
		return result, nil
	},
}

// Apply calls the given function with the given arguments.
var Apply = sxeval.Builtin{
	Name:     "apply",
	MinArity: 2,
	MaxArity: 2,
	TestPure: nil, // Might be changed in the future
	Fn2: func(env *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
		lst, err := GetList(arg1, 1)
		if err != nil {
			return nil, err
		}
		callExpr := sxeval.CallExpr{
			Proc: sxeval.ObjExpr{Obj: arg0},
			Args: nil,
		}
		if lst == nil {
			return env.ExecuteTCO(&callExpr)
		}
		callExpr.Args = append(callExpr.Args, sxeval.ObjExpr{Obj: lst.Car()})
		for node := lst; ; {
			cdr := node.Cdr()
			if sx.IsNil(cdr) {
				expr := env.MakeReworkEnvironment().Rework(&callExpr)
				return env.ExecuteTCO(expr)
			}
			if next, ok2 := sx.GetPair(cdr); ok2 {
				node = next
				callExpr.Args = append(callExpr.Args, sxeval.ObjExpr{Obj: node.Car()})
				continue
			}
			return nil, sx.ErrImproper{Pair: lst}
		}
	},
}

// Fold will apply the given function pairwise to list of args.
var Fold = sxeval.Builtin{
	Name:     "fold",
	MinArity: 3,
	MaxArity: 3,
	TestPure: nil, // Might be changed in the future
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		fn, err := GetCallable(args[0], 0)
		if err != nil {
			return nil, err
		}
		lst, err := GetList(args[2], 2)
		if err != nil {
			return nil, err
		}
		res := args[1]
		params := sx.Vector{res, res}
		for node := lst; node != nil; {
			params[0] = node.Car()
			params[1] = res
			res, err = env.Call(fn, params)
			if err != nil {
				return nil, err
			}
			next, ok := sx.GetPair(node.Cdr())
			if !ok {
				return nil, sx.ErrImproper{Pair: lst}
			}
			node = next
		}
		return res, nil
	},
}

// FoldReverse will apply the given function reversed pairwise to reversed list of args.
var FoldReverse = sxeval.Builtin{
	Name:     "fold-reverse",
	MinArity: 3,
	MaxArity: 3,
	TestPure: nil, // Might be changed in the future
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		fn, err := GetCallable(args[0], 0)
		if err != nil {
			return nil, err
		}
		lst, err := GetList(args[2], 2)
		if err != nil {
			return nil, err
		}
		rev, err := lst.Reverse()
		if err != nil {
			return nil, err
		}
		res := args[1]
		params := sx.Vector{res, res}
		for node := rev; node != nil; {
			params[0] = node.Car()
			params[1] = res
			res, err = env.Call(fn, params)
			if err != nil {
				return nil, err
			}
			node, _ = sx.GetPair(node.Cdr())
		}
		return res, nil
	},
}
