//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package list contains all list-related builtins
package list

import (
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
)

// Cons returns a cons pair of the two arguments.
func Cons(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sxpf.Cons(args[0], args[1]), nil
}

// PairP returns True if the argument is a pair.
func PairP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	obj := args[0]
	if sxpf.IsNil(obj) {
		return sxpf.False, nil
	}
	_, isPair := sxpf.GetPair(obj)
	return sxpf.MakeBoolean(isPair), nil
}

// NullP returns True if the argument is nil.
func NullP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(sxpf.IsNil(args[0])), nil
}

// ListP returns True if the argument is a (proper) list.
func ListP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(sxpf.IsList(args[0])), nil
}

// Car returns the car of a pair argument.
func Car(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 1)
	pair, err := builtins.GetPair(err, args, 0)
	if err != nil {
		return nil, err
	}
	return pair.Car(), nil
}

// Cdr returns the cdr of a pair argument.
func Cdr(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 1)
	pair, err := builtins.GetPair(err, args, 0)
	if err != nil {
		return nil, err
	}
	return pair.Cdr(), nil
}

// Last returns the last element of a list
func Last(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 1)
	lst, err := builtins.GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Last()
}

// List returns a list of all arguments.
func List(args []sxpf.Object) (sxpf.Object, error) { return sxpf.MakeList(args...), nil }

// ListStar returns a list of all arguments, when the last argument is a cons to the second last.
func ListStar(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 0); err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return args[0], nil
	}
	argPos := len(args) - 2
	result := sxpf.Cons(args[argPos], args[argPos+1])
	for argPos > 0 {
		argPos--
		result = sxpf.Cons(args[argPos], result)
	}
	return result, nil
}

// Append returns a list where all list arguments are concatenated.
func Append(args []sxpf.Object) (sxpf.Object, error) {
	switch len(args) {
	case 0:
		return sxpf.Nil(), nil
	case 1:
		return args[0], nil
	}
	lastList := len(args) - 1
	lsts := make([]*sxpf.Pair, lastList)
	for i := 0; i < lastList; i++ {
		lst, err := builtins.GetList(nil, args, i)
		if err != nil {
			return nil, err
		}
		lsts[i] = lst
	}
	sentinel := sxpf.Pair{}
	curr := &sentinel
	for _, lst := range lsts {
		for node := lst; node != nil; {
			curr = curr.AppendBang(node.Car())
			next, isPair := sxpf.GetPair(node.Cdr())
			if !isPair {
				return nil, sxpf.ErrImproper{Pair: lst}
			}
			node = next
		}
	}
	curr.SetCdr(args[lastList])
	return sentinel.Cdr(), nil
}

// Reverse returns a reversed list.
func Reverse(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 1)
	lst, err := builtins.GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Reverse()
}

// Length returns the length of the given list.
func Length(args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 1)
	lst, err := builtins.GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return sxpf.Int64(int64(lst.Length())), nil
}
