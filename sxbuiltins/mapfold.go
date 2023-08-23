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
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Map returns a list, where all member are the result of the given function to all original list members.
func Map(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 2, 2)
	fn, err := GetCallable(err, args, 0)
	lst, err := GetList(err, args, 1)
	if err != nil {
		return nil, err
	}

	if sx.IsNil(lst) {
		return sx.Nil(), nil
	}
	val, err := frame.Call(fn, []sx.Object{lst.Car()})
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
			val2, err2 := frame.Call(fn, []sx.Object{cdr2})
			if err2 != nil {
				return result, err2
			}
			curr.SetCdr(val2)
			break
		}
		val2, err2 := frame.Call(fn, []sx.Object{pair.Car()})
		if err2 != nil {
			return result, err2
		}
		curr = curr.AppendBang(val2)
		lst = pair
	}
	return result, nil
}

// Apply calls the given function with the given arguments.
func Apply(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 2, 2)
	lst, err := GetList(err, args, 1)
	if err != nil {
		return nil, err
	}
	expr := sxeval.CallExpr{
		Proc: sxeval.ObjExpr{Obj: args[0]},
		Args: nil,
	}
	if lst == nil {
		return frame.ExecuteTCO(&expr)
	}
	expr.Args = append(expr.Args, sxeval.ObjExpr{Obj: lst.Car()})
	for node := lst; ; {
		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			return frame.ExecuteTCO(&expr)
		}
		if next, ok2 := sx.GetPair(cdr); ok2 {
			node = next
			expr.Args = append(expr.Args, sxeval.ObjExpr{Obj: node.Car()})
			continue
		}
		return nil, sx.ErrImproper{Pair: lst}
	}
}

// Fold will apply the given function pairwise to list of args.
func Fold(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 3, 3)
	fn, err := GetCallable(err, args, 0)
	lst, err := GetList(err, args, 2)
	if err != nil {
		return nil, err
	}
	res := args[1]
	params := []sx.Object{res, res}
	for node := lst; node != nil; {
		params[0] = node.Car()
		params[1] = res
		res, err = frame.Call(fn, params)
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
}

// FoldReverse will apply the given function reversed pairwise to reversed list of args.
func FoldReverse(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 3, 3)
	fn, err := GetCallable(err, args, 0)
	lst, err := GetList(err, args, 2)
	if err != nil {
		return nil, err
	}
	rev, err := lst.Reverse()
	if err != nil {
		return nil, err
	}
	res := args[1]
	params := []sx.Object{res, res}
	for node := rev; node != nil; {
		params[0] = node.Car()
		params[1] = res
		res, err = frame.Call(fn, params)
		if err != nil {
			return nil, err
		}
		node, _ = sx.GetPair(node.Cdr())
	}
	return res, nil
}
