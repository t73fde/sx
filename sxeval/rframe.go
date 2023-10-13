//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval

import "zettelstore.de/sx.fossil"

// ReworkFrame guides the Expr.Rework operation.
type ReworkFrame struct {
	Env Environment // Current environment
}

// ResolveConst will resolve the symbol in an environment that is assumed not
// to b echanged afterwards.
func (rf *ReworkFrame) ResolveConst(sym *sx.Symbol) (sx.Object, bool) {
	if env := rf.Env; IsConstantBinding(env, sym) {
		return Resolve(env, sym)
	}
	return nil, false
}
