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

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

var Defined = sxeval.Builtin{
	Name:     "defined?",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		return sx.MakeBoolean(!sx.IsUndefined(args[0])), nil
	},
}
