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

import "zettelstore.de/sx.fossil"

// UndefinedPold returns true, if the given value is an undefined value.
func UndefinedPold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(sx.IsUndefined(args[0])), nil
}

// DefinedPold returns true, if the given value is not an undefined value.
func DefinedPold(args []sx.Object) (sx.Object, error) {
	if err := CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	return sx.MakeBoolean(!sx.IsUndefined(args[0])), nil
}
