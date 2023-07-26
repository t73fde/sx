//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package callable provides syntaxes and builtins to work with callables / functions / procedure.
package callable

import (
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// CallableP returns True, if the given argument is a callable.
func CallableP(args []sxpf.Object) (sxpf.Object, error) {
	if err := builtins.CheckArgs(args, 1, 1); err != nil {
		return nil, err
	}
	_, ok := eval.GetCallable(args[0])
	return sxpf.MakeBoolean(ok), nil
}
