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
	"strings"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// ToString transforms its argument into its string representation.
var ToString = sxeval.Builtin{
	Name:     "->string",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		obj := args[0]
		if s, isString := sx.GetString(obj); isString {
			return s, nil
		}
		return sx.String(obj.Repr()), nil
	},
}

// StringAppend append all its string arguments.
var StringAppend = sxeval.Builtin{
	Name:     "string-append",
	MinArity: 0,
	MaxArity: -1,
	IsPure:   true,
	Fn: func(_ *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		if len(args) == 0 {
			return sx.String(""), nil
		}
		s, err := GetString(nil, args, 0)
		if err != nil {
			return nil, err
		}
		if len(args) == 1 {
			return s, nil
		}
		var sb strings.Builder
		sb.WriteString(s.String())
		for i := 1; i < len(args); i++ {
			s, err = GetString(err, args, i)
			if err != nil {
				return nil, err
			}
			sb.WriteString(s.String())
		}
		return sx.String(sb.String()), nil
	},
}
