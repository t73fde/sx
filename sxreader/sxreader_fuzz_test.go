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

package sxreader_test

import (
	"bytes"
	"io"
	"testing"

	"t73f.de/r/sx/sxreader"
)

// FuzzReaderRead tests reader.Reader.Read() with various data.
//
// Start with: `go test -fuzz=FuzzReaderRead t73f.de/r/sx/sxreader`.
func FuzzReaderRead(f *testing.F) {
	f.Fuzz(func(t *testing.T, src []byte) {
		t.Parallel()
		rd := sxreader.MakeReader(bytes.NewReader(src))
		for {
			_, err := rd.Read()
			if err == io.EOF {
				break
			}
		}
	})
}
