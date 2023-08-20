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

// EqP returns True iff the two given arguments are identical / the same objects.
func EqP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(args[0] == args[1]), nil
}

// EqlP returns True iff the two given arguments have the same atom value.
func EqlP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(args[0].IsEql(args[1])), nil
}

// EqualP returns True iff the two given arguments have the same value.
func EqualP(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(args[0].IsEqual(args[1])), nil
}
