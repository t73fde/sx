//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package callable provides syntaxes and builtins to work with callables / functions / procedure.
package callable

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
)

// CallableP returns True, if the given argument is a callable.
func CallableP(args []sx.Object) (sx.Object, error) {
	if err := sxbuiltins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := sxeval.GetCallable(args[0])
	return sx.MakeBoolean(ok), nil
}
