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

package sxhtml

import (
	"fmt"
	"io"
	"strings"
)

type printer struct {
	w      io.Writer
	length int
	err    error
}

func (pr *printer) printString(s string) {
	if pr.err == nil {
		l, err := io.WriteString(pr.w, s)
		pr.length += l
		pr.err = err
	}
}
func (pr *printer) printStrings(sl ...string) {
	if pr.err == nil {
		for _, s := range sl {
			l, err := io.WriteString(pr.w, s)
			pr.length += l
			if err != nil {
				pr.err = err
				return
			}
		}
	}
}

const (
	htmlQuot = "&quot;" // longer than "&#34;", but often requested in standards
	htmlAmp  = "&amp;"
	htmlNull = "\uFFFD"
)

var (
	htmlEscapes = []string{
		`&`, htmlAmp,
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, htmlQuot,
		"\000", htmlNull,
	}
	htmlEscaper = strings.NewReplacer(htmlEscapes...)
)

func (pr *printer) printHTML(s string) {
	if pr.err == nil {
		l, err := htmlEscaper.WriteString(pr.w, s)
		pr.length += l
		pr.err = err
	}
}

var commentEscaper = strings.NewReplacer("--", "-&#45;")

func (pr *printer) printComment(s string) {
	if pr.err == nil {
		l, err := commentEscaper.WriteString(pr.w, s)
		pr.length += l
		pr.err = err
	}
}

func (pr *printer) printAttributeValue(s string) {
	last := 0
	var html string
	for i := range len(s) {
		switch s[i] {
		case '\000':
			html = htmlNull
		case '"':
			html = htmlQuot
		case '&':
			html = htmlAmp
		default:
			continue
		}
		pr.printStrings(s[last:i], html)
		last = i + 1
	}
	pr.printString(s[last:])
}

func urlEscape(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 32)
	written := 0
	for i, n := 0, len(s); i < n; i++ {
		ch := s[i]
		switch ch {
		case '!', '#', '$', '&', '*', '+', ',', '/', ':', ';', '=', '?', '@', '[', ']':
			continue
		case '-', '.', '_', '~':
			continue
		case '%':
			if i+2 < n && isHex(s[i+1]) && isHex(s[i+2]) {
				// If already an %-encoding, do not encode '%' twice.
				continue
			}
		default:
			if 'a' <= ch && ch <= 'z' || '0' <= ch && ch <= '9' || 'A' <= ch && ch <= 'Z' {
				continue
			}
		}
		sb.WriteString(s[written:i])
		fmt.Fprintf(&sb, "%%%02x", ch)
		written = i + 1

	}
	if written == 0 {
		return s
	}
	sb.WriteString(s[written:])
	return sb.String()
}

func isHex(ch byte) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}
