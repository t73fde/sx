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

package sx

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// String represents a string object.
type String struct{ val string }

// MakeString creates a String from a string.
func MakeString(s string) String { return String{s} }

// GetValue returns the string value of a String.
func (s String) GetValue() string { return s.val }

// IsNil return true, if it is a nil string value.
func (String) IsNil() bool { return false }

// IsAtom always returns true because a string is an atomic value.
func (String) IsAtom() bool { return true }

// IsTrue returns true if string can be interpreted as a "true" value.
func (s String) IsTrue() bool { return s.val != "" }

// IsEqual compares two objects for equivalence.
func (s String) IsEqual(other Object) bool {
	otherS, ok := other.(String)
	return ok && s.IsEqualString(otherS)

}

// IsEqualString compares two strings.
func (s String) IsEqualString(other String) bool { return s.val == other.val }

// String returns the string representation.
func (s String) String() string {
	var sb strings.Builder
	if _, err := s.Print(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

// GoString returns the go string representation.
func (s String) GoString() string { return s.val }

var (
	quote        = []byte{'"'}
	escQuote     = []byte{'\\', '"'}
	escBackslash = []byte{'\\', '\\'}
	escTab       = []byte{'\\', 't'}
	escLF        = []byte{'\\', 'n'}
	escCR        = []byte{'\\', 'r'}
	esc2         = []byte{'\\', 'x', 0, 0}
	esc4         = []byte{'\\', 'u', 0, 0, 0, 0}
	esc6         = []byte{'\\', 'U', 0, 0, 0, 0, 0, 0}
	encHex       = []byte("0123456789ABCDEF")
)

// Print write the string representation to the given Writer.
func (s String) Print(w io.Writer) (int, error) {
	last := 0
	length, err := w.Write(quote)
	if err != nil {
		return length, err
	}
	var esc []byte
	for i := 0; i < len(s.val); {
		ch, size := rune(s.val[i]), 1
		if ch >= utf8.RuneSelf {
			ch, size = utf8.DecodeRuneInString(s.val[i:])
		}
		switch ch {
		case '"':
			esc = escQuote
		case '\\':
			esc = escBackslash
		case '\t':
			esc = escTab
		case '\n':
			esc = escLF
		case '\r':
			esc = escCR
		default:
			if unicode.IsGraphic(ch) {
				i += size
				continue
			}
			if ch <= 0xff {
				esc = esc2
				esc[2] = encHex[ch>>4]
				esc[3] = encHex[ch&0xF]
			} else if ch <= 0xffff {
				esc = esc4
				esc[2] = encHex[(ch>>12)&0xF]
				esc[3] = encHex[(ch>>8)&0xF]
				esc[4] = encHex[(ch>>4)&0xF]
				esc[5] = encHex[ch&0xF]
			} else {
				esc = esc6
				esc[2] = encHex[(ch>>20)&0xF]
				esc[3] = encHex[(ch>>16)&0xF]
				esc[4] = encHex[(ch>>12)&0xF]
				esc[5] = encHex[(ch>>8)&0xF]
				esc[6] = encHex[(ch>>4)&0xF]
				esc[7] = encHex[ch&0xF]
			}
		}
		l, err2 := io.WriteString(w, s.val[last:i])
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = w.Write(esc)
		length += l
		if err2 != nil {
			return length, err2
		}
		i += size
		last = i
	}
	if last <= len(s.val) {
		l, err2 := io.WriteString(w, s.val[last:])
		length += l
		if err2 != nil {
			return length, err2
		}
	}
	l, err := w.Write(quote)
	return length + l, err
}

// GetString returns the object as a string, if possible
func GetString(obj Object) (String, bool) {
	if IsNil(obj) {
		return String{}, false
	}
	s, ok := obj.(String)
	return s, ok
}
