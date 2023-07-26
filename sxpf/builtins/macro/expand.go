//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package macro

import (
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// MacroExpand implements one level of macro expansion.
//
// It is mostly used for debugging macros.
func MacroExpand0(eng *eval.Engine, env sxpf.Environment, args []sxpf.Object) (sxpf.Object, error) {
	err := builtins.CheckArgs(args, 1, 1)
	lst, err := builtins.GetList(err, args, 0)
	if err == nil && lst != nil {
		if sym, isSymbol := sxpf.GetSymbol(lst.Car()); isSymbol {
			if obj, found := sxpf.Resolve(env, sym); found {
				if macro, isMacro := obj.(*Macro); isMacro {
					return macro.Expand(eng, env, lst.Tail())
				}
			}
		}
	}
	return lst, err
}
