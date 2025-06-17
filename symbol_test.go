//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"t73f.de/r/sx"
)

func TestSymbolBind(t *testing.T) {
	sym := sx.MakeSymbol("sym")
	if val, ok := sym.Bound(); ok {
		t.Errorf("no value bound to symbol %v, but value returned: %v", sym, val)
	}

	if err := sym.Bind(sx.Nil()); err != nil {
		t.Errorf("got error, when binding the symbol to nil")
	}
	if val, ok := sym.Bound(); !ok || !sx.IsNil(val) {
		if !ok {
			t.Errorf("unable to get bound value from symbol %v", sym)
		} else {
			t.Errorf("expected symbol %v with bound value nil, but got %T/%v", sym, val, val)
		}
	}

	if err := sym.Bind(sym); err != nil {
		t.Errorf("got error, when binding the symbol to itself")
	}

	if val, ok := sym.Bound(); !ok || !sym.IsEqual(val) {
		if !ok {
			t.Errorf("unable to get bound value from symbol %v", sym)
		} else {
			t.Errorf("expected symbol %v as bound value, but got %T/%v", sym, val, val)
		}
	}

	sym.Freeze()
	if !sym.IsFrozen() {
		t.Errorf("symbol %s must be frozen, but is not", sym)
	}
	err := sym.Bind(sx.Int64(17))
	if err == nil {
		t.Errorf("no error when binding to frozen symbol")
	} else if ferr, ok := err.(sx.ErrSymbolFrozen); !ok {
		t.Errorf("expected frozen error, but got %T/%v", err, err)
	} else if ferr.Symbol != sym {
		t.Errorf("got frozen error, but expexpectedted symbol %v, but got %v", sym, ferr.Symbol)
	}

	if val, ok := sym.Bound(); !ok || !sym.IsEqual(val) {
		if !ok {
			t.Errorf("unable to get bound value from symbol %v", sym)
		} else {
			t.Errorf("expected symbol %v as bound value, but got %T/%v", sym, val, val)
		}
	}
}
