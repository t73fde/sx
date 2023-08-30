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
)

// ToString transforms its argument into its string representation.
func ToString(args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	if err != nil {
		return nil, err
	}
	obj := args[0]
	if s, isString := sx.GetString(obj); isString {
		return s, nil
	}
	return sx.MakeString(obj.Repr()), nil
}

// StringAppend append all its string arguments.
func StringAppend(args []sx.Object) (sx.Object, error) {
	if len(args) == 0 {
		return sx.MakeString(""), nil
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
	return sx.MakeString(sb.String()), nil
}
