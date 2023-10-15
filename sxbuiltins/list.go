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
)

// ConsOld returns a cons pair of the two arguments.
func ConsOld(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sx.Cons(args[0], args[1]), nil
}

// PairPold returns True if the argument is a pair.
func PairPold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	obj := args[0]
	if sx.IsNil(obj) {
		return sx.Nil(), nil
	}
	_, isPair := sx.GetPair(obj)
	return sx.MakeBoolean(isPair), nil
}

// NullPold returns True if the argument is nil.
func NullPold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsNil(args[0])), nil
}

// ListPold returns True if the argument is a (proper) list.
func ListPold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsList(args[0])), nil
}

// CarOld returns the car of a pair argument.
func CarOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	pair, err := GetPair(err, args, 0)
	if err != nil {
		return nil, err
	}
	return pair.Car(), nil
}

// CdrOld returns the cdr of a pair argument.
func CdrOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	pair, err := GetPair(err, args, 0)
	if err != nil {
		return nil, err
	}
	return pair.Cdr(), nil
}

func cxr(args []sx.Object, spec string) (result sx.Object, _ error) {
	err := CheckArgs(args, 1, 1)
	pair, err := GetPair(err, args, 0)
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
}

func CaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aa") }
func CadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "ad") }
func CdarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "da") }
func CddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "dd") }

func CaaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aaa") }
func CaadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aad") }
func CadarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "ada") }
func CaddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "add") }
func CdaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "daa") }
func CdadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "dad") }
func CddarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "dda") }
func CdddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "ddd") }

func CaaaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aaaa") }
func CaaadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aaad") }
func CaadarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aada") }
func CaaddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "aadd") }
func CadaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "adaa") }
func CadadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "adad") }
func CaddarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "adda") }
func CadddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "addd") }
func CdaaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "daaa") }
func CdaadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "daad") }
func CdadarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "dada") }
func CdaddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "dadd") }
func CddaarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "ddaa") }
func CddadrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "ddad") }
func CdddarOld(args []sx.Object) (sx.Object, error) { return cxr(args, "ddda") }
func CddddrOld(args []sx.Object) (sx.Object, error) { return cxr(args, "dddd") }

// LastOld returns the last element of a list
func LastOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Last()
}

// ListOld returns a list of all arguments.
func ListOld(args []sx.Object) (sx.Object, error) { return sx.MakeList(args...), nil }

// ListStarOld returns a list of all arguments, when the last argument is a cons to the second last.
func ListStarOld(args []sx.Object) (sx.Object, error) {
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

// AppendOld returns a list where all list arguments are concatenated.
func AppendOld(args []sx.Object) (sx.Object, error) {
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

// ReverseOld returns a reversed list.
func ReverseOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Reverse()
}

// LengthOld returns the length of the given list.
func LengthOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return sx.Int64(int64(lst.Length())), nil
}

// AssocOld returns the first pair of the a-list where the second argument is eql?
// to the pair's car. Otherwise, nil is returned.
func AssocOld(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 2, 2)
	lst, err := GetList(err, args, 0)
	if err != nil {
		return nil, err
	}
	return lst.Assoc(args[1]), nil
}
