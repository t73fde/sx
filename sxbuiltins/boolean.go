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

// Contains builtins and syntax for boolean values.

import "zettelstore.de/sx.fossil"

// Boolean returns the given value interpreted as a boolean.
func Boolean(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsTrue(args[0])), nil
}

// Not negates the given value interpreted as a boolean.
func Not(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(!sx.IsTrue(args[0])), nil
}
