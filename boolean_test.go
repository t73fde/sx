//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"zettelstore.de/sx.fossil"
)

func TestBoolean(t *testing.T) {
	t.Parallel()
	if sx.IsTrue(sx.String("")) {
		t.Error("Empty string is True")
	}
}
