//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package reader_test

import (
	"bytes"
	"io"
	"testing"

	"zettelstore.de/sx.fossil/sxpf/reader"
)

// FuzzReaderRead tests reader.Reader.Read() with various data.
//
// Start with: `go test -fuzz=FuzzReaderRead zettelstore.de/sx.fossil/sxpf/reader`.
func FuzzReaderRead(f *testing.F) {
	f.Fuzz(func(t *testing.T, src []byte) {
		t.Parallel()
		rd := reader.MakeReader(bytes.NewReader(src))
		for {
			_, err := rd.Read()
			if err == io.EOF {
				break
			}
		}
	})
}
