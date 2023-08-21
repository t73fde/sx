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

import "zettelstore.de/sx.fossil"

// Cons returns a cons pair of the two arguments.
func Cons(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sx.Cons(args[0], args[1]), nil
}

// PairP returns True if the argument is a pair.
func PairP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	obj := args[0]
	if sx.IsNil(obj) {
		return sx.False, nil
	}
	_, isPair := sx.GetPair(obj)
	return sx.MakeBoolean(isPair), nil
}

// NullP returns True if the argument is nil.
func NullP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsNil(args[0])), nil
}

// ListP returns True if the argument is a (proper) list.
func ListP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsList(args[0])), nil
}

// Car returns the car of a pair argument.
func Car(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	pair, err := GetPair(err, args, 0)
	if err != nil {
		return nil, err
	}
	return pair.Car(), nil
}

// Cdr returns the cdr of a pair argument.
func Cdr(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	pair, err := GetPair(err, args, 0)
	if err != nil {
		return nil, err
	}
	return pair.Cdr(), nil
}

// Last returns the last element of a list
func Last(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Last()
}

// List returns a list of all arguments.
func List(args []sx.Object) (sx.Object, error) { return sx.MakeList(args...), nil }

// ListStar returns a list of all arguments, when the last argument is a cons to the second last.
func ListStar(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 0); err != nil {
		return nil, err
	}
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
}

// Append returns a list where all list arguments are concatenated.
func Append(args []sx.Object) (sx.Object, error) {
	switch len(args) {
	case 0:
		return sx.Nil(), nil
	case 1:
		return args[0], nil
	}
	lastList := len(args) - 1
	lsts := make([]*sx.Pair, lastList)
	for i := 0; i < lastList; i++ {
		lst, err := GetList(nil, args, i)
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
}

// Reverse returns a reversed list.
func Reverse(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Reverse()
}

// Length returns the length of the given list.
func Length(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return sx.Int64(int64(lst.Length())), nil
}
