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
func MacroExpand0(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := sxbuiltins.CheckArgs(args, 1, 1)
	lst, err := sxbuiltins.GetList(err, args, 0)
	if err == nil && lst != nil {
		if sym, isSymbol := sx.GetSymbol(lst.Car()); isSymbol {
			if obj, found := frame.Resolve(sym); found {
				if macro, isMacro := obj.(*Macro); isMacro {
					return macro.Expand(frame, lst.Tail())
				}
			}
		}
	}
	return lst, err
}
