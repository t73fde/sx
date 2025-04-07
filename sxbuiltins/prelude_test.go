//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import (
	"testing"

	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/sx/sxeval"
)

func TestLoadPrelude(t *testing.T) {
	root := sxeval.MakeRootBinding(128)
	if err := sxbuiltins.LoadPrelude(root); err != nil {
		t.Error(err)
	}
}
