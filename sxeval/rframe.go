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
	Env      Environment // Current environment
	constEnv Environment // Environment that is seen as constant / freezed.
}

// ResolveConst will resolve the symbol in an environment that is assumed not to
// be changed afterwards.
func (rf *ReworkFrame) ResolveConst(sym *sx.Symbol) (sx.Object, bool) {
	if env := rf.Env; env.IsConst(sym) {
		return Resolve(env, sym)
	}
	if constEnv := rf.constEnv; constEnv != nil {
		return Resolve(constEnv, sym)
	}
	return nil, false
}
