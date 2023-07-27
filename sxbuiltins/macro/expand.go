//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package macro

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
)

// MacroExpand implements one level of macro expansion.
//
// It is mostly used for debugging macros.
func MacroExpand0(eng *sxeval.Engine, env sx.Environment, args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 1, 1)
	lst, err := sxbuiltins.GetList(err, args, 0)
	if err == nil && lst != nil {
		if sym, isSymbol := sx.GetSymbol(lst.Car()); isSymbol {
			if obj, found := sx.Resolve(env, sym); found {
				if macro, isMacro := obj.(*Macro); isMacro {
					return macro.Expand(eng, env, lst.Tail())
				}
			}
		}
	}
	return lst, err
}
