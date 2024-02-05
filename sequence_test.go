//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"zettelstore.de/sx.fossil"
)

func TestSequence(t *testing.T) {
	lst := sx.Sequence(sx.MakeList(sx.Int64(1), sx.Int64(2), sx.Int64(3)))
	vec := sx.Sequence(sx.Vector{sx.Int64(1), sx.Int64(2), sx.Int64(3)})

	lstIter := lst.Iterator()
	vecIter := vec.Iterator()
	for {
		lstHas := lstIter.HasElement()
		vecHas := vecIter.HasElement()
		if lstHas != vecHas {
			t.Errorf("lstHas %v != vecHas %v", lstHas, vecHas)
			break
		}
		if !lstHas || !vecHas {
			break
		}
		lstElem := lstIter.Element()
		vecElem := vecIter.Element()
		if !lstElem.IsEqual(vecElem) {
			t.Errorf("lstElem %v != vecElem %v", lstElem, vecElem)
			break
		}
		lstAdv := lstIter.Advance()
		vecAdv := vecIter.Advance()
		if lstAdv != vecAdv {
			t.Errorf("lstAdv %v != vecAdv %v", lstAdv, vecAdv)
			break
		}
	}
	exp := sx.MakeUndefined()
	if got := lstIter.Element(); !exp.IsEqual(got) {
		t.Errorf("lstIter.Elements() should be %v, but got %v", exp, got)
	}
	if got := vecIter.Element(); !exp.IsEqual(got) {
		t.Errorf("vecIter.Elements() should be %v, but got %v", exp, got)
	}
}
