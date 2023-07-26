//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package equiv contains function to test for equivalence of objects.
package equiv

import (
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
)

// EqP returns True iff the two given arguments are identical / the same objects.
func EqP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(args[0] == args[1]), nil
}

// EqlP returns True iff the two given arguments have the same atom value.
func EqlP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(args[0].IsEql(args[1])), nil
}

// EqualP returns True iff the two given arguments have the same value.
func EqualP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 2, 2); err != nil {
		return nil, err
	}
	return sxpf.MakeBoolean(args[0].IsEqual(args[1])), nil
}
