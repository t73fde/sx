//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

// Contains function to test for equivalence of objects.

import "zettelstore.de/sx.fossil"

// IdenticalOld returns True iff the two given arguments are identical / the same objects.
func IdenticalOld(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 0); err != nil {
		return nil, err
	}
	for i := 1; i < len(args); i++ {
		if args[0] != args[i] {
			return sx.Nil(), nil
		}
	}
	return sx.MakeBoolean(true), nil
}

// EqualOld returns True iff the two given arguments have the same value.
func EqualOld(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 0); err != nil {
		return nil, err
	}
	for i := 1; i < len(args); i++ {
		if !args[0].IsEqual(args[i]) {
			return sx.Nil(), nil
		}
	}
	return sx.MakeBoolean(true), nil
}
