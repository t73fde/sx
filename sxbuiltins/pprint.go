//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

// Provides some function to pretty-print objects.

import (
	"io"
	"os"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Pretty writes the first argument to stdout.
var Pretty = sxeval.Builtin{
	Name:     "pp",
	MinArity: 0,
	MaxArity: 1,
	TestPure: nil,
	Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		if len(args) == 0 {
			return sx.Nil(), nil
		}
		_, err := Print(os.Stdout, args[0])
		return sx.Nil(), err
	},
}

// Print object to given writer in a pretty way.
func Print(w io.Writer, obj sx.Object) (int, error) {
	written, err := doPrint(w, obj, 0)
	if err != nil {
		return written, err
	}
	n, err := w.Write(bEOL)
	return written + n, err
}

func doPrint(w io.Writer, obj sx.Object, indent int) (int, error) {
	if sx.IsNil(obj) {
		return w.Write(bNil)
	}
	if pair, isPair := sx.GetPair(obj); isPair {
		return doPrintList(w, pair, indent)
	}

	return sx.Print(w, obj)
}

func doPrintList(w io.Writer, lst *sx.Pair, indent int) (int, error) {
	written, errWrite := w.Write(bOpen)
	if errWrite != nil {
		return written, errWrite
	}

	pos := 0
	mustIndent := false
	for node := lst; ; pos++ {
		if mustIndent {
			n, err := writeIndent(w, indent+4)
			written += n
			if err != nil {
				return written, err
			}
			mustIndent = false
		}

		n, err := doPrint(w, node.Car(), indent+1)
		written += n
		if err != nil {
			return written, err
		}

		cdr := node.Cdr()
		if sx.IsNil(cdr) {
			break
		}

		if next, isPair := sx.GetPair(cdr); isPair {
			contBytes, mi := calcContinuation(next, pos)
			nContBytes, errContBytes := w.Write(contBytes)
			written += nContBytes
			if errContBytes != nil {
				return written, errContBytes
			}
			node = next
			mustIndent = mi
			continue
		}

		n, err = w.Write(bDot)
		written += n
		if err != nil {
			return written, err
		}

		n, err = doPrint(w, cdr, indent)
		written += n
		if err != nil {
			return written, err
		}
		break
	}
	n, errWrite := w.Write(bClose)
	return written + n, errWrite
}

func calcContinuation(next *sx.Pair, pos int) ([]byte, bool) {
	if pos == 0 && next.Car().IsAtom() {
		return bSpace, false
	}
	return bEOL, true
}

var (
	bEOL   = []byte{'\n'}
	bNil   = []byte{'(', ')'}
	bOpen  = []byte{'('}
	bSpace = []byte{' '}
	bDot   = []byte{' ', '.', ' '}
	bClose = []byte{')'}
)

// 80 spaces
const spaces = "                                                                                "

func writeIndent(w io.Writer, indent int) (int, error) {
	var written int
	for indent > len(spaces) {
		n, err := io.WriteString(w, spaces)
		written += n
		if err != nil {
			return written, err
		}
		indent -= len(spaces)
	}
	n, err := io.WriteString(w, spaces[:indent])
	written += n
	return written, err
}
