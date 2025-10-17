//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sx

// MakeBoolean creates a new Boolean object.
func MakeBoolean(b bool) Object {
	if b {
		return T
	}
	return Nil()
}

// T is the default true object.
var T = MakeSymbol("T")

func init() {
	_ = T.Bind(T)
	T.Freeze()
}

// IsTrue returns true, if the given object will be interpreted as "true" in a boolean context.
func IsTrue(obj Object) bool { return obj != nil && obj.IsTrue() }

// IsFalse returns true, if the given object will be interpreted as "false" in a boolean context.
func IsFalse(obj Object) bool { return obj == nil || !obj.IsTrue() }
