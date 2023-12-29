//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sxreader_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"zettelstore.de/sx.fossil/sxreader"
)

func TestReaderName(t *testing.T) {
	testcases := []struct {
		name string
		r    io.Reader
		exp  string
	}{
		{name: "WithStringReader", r: strings.NewReader("test"), exp: "<string>"},
		{name: "WithBytesReader", r: bytes.NewReader([]byte("test")), exp: "<bytes>"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(tc.r)
			if got := rd.Name(); got != tc.exp {
				t.Errorf("Name expected: %s, but got %s", tc.exp, got)
			}
		})
	}
}

// func TestReaderNext(t *testing.T) {
// 	content := "98765EDCBA43210"
// 	rd := sxreader.MakeReader(strings.NewReader(content))
// 	for i := 0; ; i++ {
// 		r, err := rd.NextRune()
// 		if err != nil {
// 			if err != io.EOF {
// 				t.Errorf("Unexpected error %v", err)
// 			} else if r != -1 {
// 				t.Errorf("On EOF rune must be -1, but got %v", r)
// 			}
// 			break
// 		}
// 		if exp := rune(content[i]); r != exp {
// 			t.Errorf("Position: %d, exptected rune %v, but got %v", i, exp, r)
// 		}
// 	}
// }

type readerTestCase struct {
	name    string
	src     string
	exp     string
	mustErr bool
}

func TestReaderInteger(t *testing.T) {
	performReaderTestCases(t, []readerTestCase{
		{name: "zero", src: "0", exp: "0"},
		{name: "double zero", src: "00", exp: "0"},
		{name: "one", src: "1", exp: "1"},
		{name: "double one", src: "11", exp: "11"},
		{name: "WithLeadingSpaces", src: " \t 123", exp: "123"},
		{name: "PositiveInt", src: "+321", exp: "321"},
		{name: "NegativeInt", src: "-6543", exp: "-6543"},
		{name: "WithComment", src: " 234;comment", exp: "234"},
		{name: "TrailingSpace", src: "345 ", exp: "345"},
		{name: "InvalidValue", src: "123x", exp: "123x"},
		{name: "NoNumberSymbol", src: "17-4", exp: "17-4"},
	})
}

func TestReaderSymbol(t *testing.T) {
	performReaderTestCases(t, []readerTestCase{
		{name: "bang zero", src: "!0", exp: "!0"},
		{name: "Ascii", src: "moin", exp: "moin"},
		{name: "Unicode", src: "µ☺", exp: "µ☺"},
		{name: "Single char", src: "+", exp: "+"},
		{name: "ColonSymbol", src: "+:", exp: "+:"},
		{name: "Single char", src: "-", exp: "-"},
		{name: "ColonSymbol", src: "-:", exp: "-:"},
		{name: "NamespaceSymbol", src: "html:body", exp: "html:body"},
	})
}

func TestReaderString(t *testing.T) {
	performReaderTestCases(t, []readerTestCase{
		{name: "Empty", src: `""`, exp: `""`},
		{name: "Simple", src: `"moin"`, exp: `"moin"`},
		{name: "EscQuote", src: `"moin\""`, exp: `"moin\""`},
		{name: "EscEsc", src: `"moin\\"`, exp: `"moin\\"`},
		{name: "EscTab", src: `"moin\t"`, exp: `"moin\t"`},
		{name: "EscCRLF", src: `"mo\r\nin"`, exp: `"mo\r\nin"`},
		{name: "Esc2hex", src: `"\x41"`, exp: `"A"`},
		{name: "Esc2HexCR", src: `"\x0a"`, exp: `"\n"`},
		{name: "Esc2Hex1B", src: `"\x1b"`, exp: `"\x1B"`},
		{name: "Esc4hex", src: `"\u0041"`, exp: `"A"`},
		{name: "Esc4HexNL", src: `"\u000d"`, exp: `"\r"`},
		{name: "Esc4HexNonGraphic", src: `"\ue8Fe"`, exp: `"\uE8FE"`},
		{name: "Esc6hex", src: `"\U000041"`, exp: `"A"`},
		{name: "Esc6hexNonGraphic", src: `"\U0e0000"`, exp: `"\U0E0000"`},
		{name: "EscUnknown", src: `"moin\x"`, exp: `ReaderError 1-8: no hex digit found: "/34`, mustErr: true},
		{name: "MissingQuote", src: `"moin`, exp: "ReaderError 1-5: unexpected EOF", mustErr: true},
		{name: "EscEOF", src: `"moin\`, exp: "ReaderError 1-6: unexpected EOF", mustErr: true},
	})
}

func TestReadList(t *testing.T) {
	performReaderTestCases(t, []readerTestCase{
		{name: "empty list", src: "()", exp: "()"},
		{name: "empty list with spaces", src: " ( )", exp: "()"},
		{name: "one value", src: "( 1 )", exp: "(1)"},
		{name: "two values", src: "( 1 2)", exp: "(1 2)"},
		{name: "list of two nils", src: "(()())", exp: "(() ())"},
		{name: "unbalanced", src: ")", exp: "ReaderError 1-1: unmatched delimiter ')'", mustErr: true},
		{name: "EOF", src: "(1 2", exp: "ReaderError 1-4: unexpected EOF", mustErr: true},
		{name: "WithComment", src: "(1 ; one\n a\n µ)", exp: "(1 a µ)"},
		{name: "SimpleDot", src: "(1 . 2)", exp: "(1 . 2)"},
		{name: "NilDot", src: "(. 2)", exp: "(() . 2)"},
		{name: "DotList", src: "(1 . (2 3))", exp: "(1 2 3)"},
		{name: "DotEmpty", src: "(1 .)", exp: "ReaderError 1-4: '.' not allowed here", mustErr: true},
		{name: "DotEmpty", src: "(1 . )", exp: "ReaderError 1-6: unmatched delimiter ')'", mustErr: true},
	})
}

func TestReadComment(t *testing.T) {
	performReaderTestCases(t, []readerTestCase{
		{name: "comment only", src: ";", exp: "EOF", mustErr: true},
		{name: "double comment only", src: ";;", exp: "EOF", mustErr: true},
		{name: "triple comment only", src: ";;;", exp: "EOF", mustErr: true},
		{name: "triple comment only spaces", src: " ; ; ; ", exp: "EOF", mustErr: true},
		{name: "comment text", src: "; abc", exp: "EOF", mustErr: true},
		{name: "double comment text", src: ";; def", exp: "EOF", mustErr: true},
		{name: "triple comment text", src: ";;;ghi", exp: "EOF", mustErr: true},
		{name: "triple comment text spaces", src: " ; a ; b ;c ", exp: "EOF", mustErr: true},
		{name: "comment only after int", src: "3;", exp: "3"},
		{name: "comment text after int", src: "3 ; three", exp: "3"},
		{name: "new line in commented list", src: "(3 ;; three\n  4)", exp: "(3 4)"},
		{name: "", src: "(\n3\n;;; line\n4\n)", exp: "(3 4)"},
	})
}

func TestReadHash(t *testing.T) {
	performReaderTestCases(t, []readerTestCase{
		{name: "hash only", src: "#", exp: "ReaderError 1-1: '#' not allowed here", mustErr: true},
	})
}

func performReaderTestCases(t *testing.T, testcases []readerTestCase) {
	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src), sxreader.WithDefaultSymbolFactory)
			val, err := rd.Read()
			if err != nil {
				got := err.Error()
				if tc.mustErr {
					if got != tc.exp {
						t.Errorf("Input: %q, expected error %q, but got %q", tc.src, tc.exp, got)
					}
				} else {
					t.Errorf("Input: %q resulted in unexpected error:\n%s", tc.src, fmt.Sprintf("%#s", err))
				}
			} else {
				got := val.Repr()
				if tc.mustErr {
					t.Errorf("Input: %q should result in error %q, but got value %q", tc.src, tc.exp, got)
				} else if got != tc.exp {
					t.Errorf("Input: %q, expected %q, but got %q", tc.src, tc.exp, got)
				}
			}
		})
	}
}

func TestReadAll(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		src string
		exp string
	}{
		{"", "[]"},
		{"1", "[1]"},
		{"1 2", "[1 2]"},
		{"(1 2 3) a (x y z)", "[(1 2 3) a (x y z)]"},
	}
	for i, tc := range testcases {
		rd := sxreader.MakeReader(strings.NewReader(tc.src))
		objs, err := rd.ReadAll()
		if err != nil {
			t.Errorf("%d: error while reading: %v", i, err)
			continue
		}
		if got := fmt.Sprintf("%v", objs); got != tc.exp {
			t.Errorf("%d: %q expected, but got %q", i, tc.exp, got)
		}
	}
}

func TestReaderLimits(t *testing.T) {
	t.Parallel()
	err := checkNested(sxreader.DefaultNestingLimit, sxreader.DefaultNestingLimit)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkNested(sxreader.DefaultNestingLimit, sxreader.DefaultNestingLimit+1)
	if !errors.Is(err, sxreader.ErrTooDeeplyNested) {
		t.Errorf("%v, but got %v", sxreader.ErrTooDeeplyNested, err)
	}
	err = checkLength(sxreader.DefaultListLimit, sxreader.DefaultListLimit)
	if err != nil {
		t.Error(err)
		return
	}
	err = checkLength(sxreader.DefaultListLimit, sxreader.DefaultListLimit+1)
	if !errors.Is(err, sxreader.ErrListTooLong) {
		t.Errorf("%v, but got %v", sxreader.ErrListTooLong, err)
	}
}

func checkNested(maxDepth, depth int) error {
	inp := strings.Repeat("(", depth) + "1" + strings.Repeat(")", depth)
	rd := sxreader.MakeReader(strings.NewReader(inp), sxreader.WithNestingLimit(uint(maxDepth)))
	if _, err := rd.Read(); err != nil {
		return err
	}
	if _, err := rd.Read(); err != io.EOF {
		return fmt.Errorf("io.EOF exprected, but got %v", err)
	}
	return nil
}

func checkLength(maxLength, length int) error {
	inp := "(" + strings.Repeat(" 7", length) + " )"
	rd := sxreader.MakeReader(strings.NewReader(inp), sxreader.WithListLimit(uint(maxLength)))
	if _, err := rd.Read(); err != nil {
		return err
	}
	return nil
}
