//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package callable

import (
	"log"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// Map returns a list, where all member are the result of the given function to all original list members.
func Map(eng *eval.Engine, env sxpf.Environment, args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 2, 2)
	fn, err := builtins.GetCallable(err, args, 0)
	lst, err := builtins.GetList(err, args, 1)
	if err != nil {
		return nil, err
	}

	if sxpf.IsNil(lst) {
		return sxpf.Nil(), nil
	}
	val, err := eng.Call(env, fn, []sxpf.Object{lst.Car()})
	if err != nil {
		return sxpf.Nil(), nil
	}
	result := sxpf.Cons(val, sxpf.Nil())
	curr := result
	for {
		cdr2 := lst.Cdr()
		if sxpf.IsNil(cdr2) {
			break
		}
		pair, isPair := sxpf.GetPair(cdr2)
		if !isPair {
			val2, err2 := eng.Call(env, fn, []sxpf.Object{cdr2})
			if err2 != nil {
				return result, err2
			}
			curr.SetCdr(val2)
			break
		}
		val2, err2 := eng.Call(env, fn, []sxpf.Object{pair.Car()})
		if err2 != nil {
			return result, err2
		}
		curr = curr.AppendBang(val2)
		lst = pair
	}
	return result, nil
}

// Apply calls the given function with the given arguments.
func Apply(eng *eval.Engine, env sxpf.Environment, args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 2, 2)
	lst, err := builtins.GetList(err, args, 1)
	if err != nil {
		return nil, err
	}
	expr := eval.CallExpr{
		Proc: eval.ObjExpr{Obj: args[0]},
		Args: nil,
	}
	if lst == nil {
		return eng.ExecuteTCO(env, &expr)
	}
	expr.Args = append(expr.Args, eval.ObjExpr{Obj: lst.Car()})
	for node := lst; ; {
		cdr := node.Cdr()
		if sxpf.IsNil(cdr) {
			return eng.ExecuteTCO(env, &expr)
		}
		if next, ok2 := sxpf.GetPair(cdr); ok2 {
			node = next
			expr.Args = append(expr.Args, eval.ObjExpr{Obj: node.Car()})
			continue
		}
		return nil, sxpf.ErrImproper{Pair: lst}
	}
}

// Fold will apply the given function pairwise to list of args.
func Fold(eng *eval.Engine, env sxpf.Environment, args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 3, 3)
	fn, err := builtins.GetCallable(err, args, 0)
	lst, err := builtins.GetList(err, args, 2)
	if err != nil {
		return nil, err
	}
	res := args[1]
	params := []sxpf.Object{res, res}
	for node := lst; node != nil; {
		params[0] = node.Car()
		params[1] = res
		res, err = eng.Call(env, fn, params)
		if err != nil {
			return nil, err
		}
		next, ok := sxpf.GetPair(node.Cdr())
		if !ok {
			return nil, sxpf.ErrImproper{Pair: lst}
		}
		node = next
	}
	return res, nil
}

// FoldReverse will apply the given function reversed pairwise to reversed list of args.
func FoldReverse(eng *eval.Engine, env sxpf.Environment, args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 3, 3)
	fn, err := builtins.GetCallable(err, args, 0)
	lst, err := builtins.GetList(err, args, 2)
	if err != nil {
		return nil, err
	}
	rev, err := lst.Reverse()
	if err != nil {
		return nil, err
	}
	log.Println("REVL", rev)
	res := args[1]
	params := []sxpf.Object{res, res}
	for node := rev; node != nil; {
		params[0] = node.Car()
		params[1] = res
		res, err = eng.Call(env, fn, params)
		if err != nil {
			return nil, err
		}
		node, _ = sxpf.GetPair(node.Cdr())
	}
	return res, nil
}
