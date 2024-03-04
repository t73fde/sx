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

package sxreader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"zettelstore.de/sx.fossil"
)

// Reader consumes characters from a stream and parses them into s-expressions.
type Reader struct {
	rr      io.RuneReader
	err     error
	name    string
	buf     []rune
	line    int
	col     int
	prevCol int
	macros  macroMap

	maxDepth, curDepth uint
	maxLength          uint
}

// macroFn is a function that reads according to its own syntax.
type macroFn func(*Reader, rune) (sx.Object, error)

// macroMap maps rune to read macros.
type macroMap map[rune]macroFn

// Position stores the positional information about a value within the reader.
type Position struct {
	Name string
	Line int
	Col  int
}

func (rp *Position) String() string {
	name := rp.Name
	if name == "" {
		name = "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", name, rp.Line, rp.Col)
}

// Option is a function to modify the default reader when it is made.
type Option func(*Reader)

// WithNestingLimit sets the maximum nesting for a object.
func WithNestingLimit(depth uint) Option {
	return func(rd *Reader) { rd.maxDepth = depth }
}

// WithListLimit sets the maximum length of a list.
func WithListLimit(length uint) Option {
	return func(rd *Reader) { rd.maxLength = length }
}

// DefaultNestingLimit specifies the default value for `WithNestingLimit`.
const DefaultNestingLimit = 1000

// DefaultListLimit specifies the default valur for `WithListLimit`.
const DefaultListLimit = 10000

// MakeReader creates a new reader.
func MakeReader(r io.Reader, opts ...Option) *Reader {
	rd := Reader{
		rr:      bufio.NewReader(r),
		err:     nil,
		name:    inferReaderName(r),
		buf:     []rune{},
		line:    0,
		col:     0,
		prevCol: 0,
		macros: macroMap{
			'"':  readString,
			'#':  readHash,
			'\'': readQuote,
			'(':  readList(')'),
			')':  unmatchedDelimiter,
			',':  readUnquote,
			'.':  readDot,
			';':  readComment,
			'`':  readQuasiquote,
		},
		maxDepth:  DefaultNestingLimit,
		maxLength: DefaultListLimit,
	}
	for _, opt := range opts {
		opt(&rd)
	}
	return &rd
}
func inferReaderName(r io.Reader) string {
	switch tr := r.(type) {
	case *strings.Reader:
		return "<string>"
	case *bytes.Reader:
		return "<bytes>"
	default:
		return fmt.Sprintf("<%T>", tr)
	}
}

// Name return the name of the underlying reader.
func (rd *Reader) Name() string { return rd.name }

// nextRune returns the next rune from the reader and advances the reader.
func (rd *Reader) nextRune() (rune, error) {
	if rd.err != nil {
		return -1, rd.err
	}
	var ch rune
	if len(rd.buf) > 0 {
		ch = rd.buf[0]
		if len(rd.buf) > 1 {
			rd.buf = rd.buf[1:]
		} else {
			rd.buf = nil
		}
	} else {
		var err error
		ch, _, err = rd.rr.ReadRune()
		if err != nil {
			rd.err = err
			return -1, err
		}
	}

	if ch == '\n' {
		rd.line++
		rd.prevCol = rd.col
		rd.col = 0
	} else {
		rd.col++
	}
	return ch, nil
}

// unreadRunes returns runes consumed from the reader back to it.
func (rd *Reader) unreadRunes(chs ...rune) {
	hasNewline := false
	for _, ch := range chs {
		if ch == '\n' {
			hasNewline = true
		}
	}

	if hasNewline {
		rd.line--
		rd.col = rd.prevCol
	} else {
		rd.col--
	}
	rd.buf = append(chs, rd.buf...)
}

// Position returns information about the current position of the reader.
func (rd *Reader) Position() Position {
	return Position{
		Name: rd.name,
		Line: rd.line + 1,
		Col:  rd.col,
	}
}

// skipSpace skips all space rune from the reader and return the first non-space rune.
func (rd *Reader) skipSpace() (rune, error) {
	for {
		ch, err := rd.nextRune()
		if err != nil {
			return -1, err
		}
		if !isSpace(ch) {
			return ch, nil
		}
	}
}

// isSpace returns true, if the rune is a space character.
func isSpace(ch rune) bool {
	return (ch <= ' ' && ch >= 0) || unicode.IsSpace(ch)
}

// isTerminal returns true, if rune terminates the current token, according to the given read macro map.
func (rd *Reader) isTerminal(ch rune) bool {
	_, found := rd.macros[ch]
	return found || unicode.In(ch, unicode.C, unicode.Z) // C=Control, Z=Separator
}

// Read one s-expression and return it.
func (rd *Reader) Read() (sx.Object, error) {
	if rd.curDepth > rd.maxDepth {
		return nil, ErrTooDeeplyNested
	}
	rd.curDepth++
	defer func() { rd.curDepth-- }()
	for {
		val, err := rd.readValue()
		if err == nil {
			return val, nil
		}
		if !errors.Is(err, ErrSkip) {
			return nil, err
		}
	}
}

// ReadAll s-expressions until EOF.
func (rd *Reader) ReadAll() (objs sx.Vector, _ error) {
	for {
		val, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				return objs, nil
			}
			return objs, err
		}
		objs = append(objs, val)
	}
}

// ErrTooDeeplyNested is returned, if the reader should read an object that
// is too deeply nested.
var ErrTooDeeplyNested = errors.New("too deeply nested")

// ErrListTooLong is returned, if the reader should read a list that is too long.
var ErrListTooLong = errors.New("list too long")

func (rd *Reader) readValue() (sx.Object, error) {
	ch, err := rd.skipSpace()
	if err != nil {
		return nil, err
	}
	if isNumber(ch) {
		return readNumber(rd, ch)
	}
	if ch == '+' {
		ch2, err2 := rd.nextRune()
		if err2 != io.EOF {
			if err2 != nil {
				return nil, err2
			}
			if isNumber(ch2) {
				return readNumber(rd, ch2)
			}
			rd.unreadRunes(ch2)
		}
	} else if ch == '-' {
		ch2, err2 := rd.nextRune()
		if err2 != io.EOF {
			if err2 != nil {
				return nil, err2
			}
			rd.unreadRunes(ch2)
			if isNumber(ch2) {
				return readNumber(rd, ch)
			}
		}
	}

	if m, found := rd.macros[ch]; found {
		return m(rd, ch)
	}
	return readSymbol(rd, ch)
}

func isNumber(ch rune) bool { return '0' <= ch && ch <= '9' }

// readToken reads a sequence of non-terminal runes from the reader.
// if initCh > ' ', it is included as the first char.
func (rd *Reader) readToken(firstCh rune, isTerminal func(rune) bool) (string, error) {
	var sb strings.Builder
	if firstCh > ' ' {
		sb.WriteRune(firstCh)
	}
	for {
		ch, err := rd.nextRune()
		if err != nil {
			if err == io.EOF {
				return sb.String(), nil
			}
			return sb.String(), err
		}

		if isTerminal(ch) {
			rd.unreadRunes(ch)
			return sb.String(), nil
		}

		sb.WriteRune(ch)
	}
}

// annotateError adds error information (reader.Name, position) to the given error.
func (rd *Reader) annotateError(err error, begin Position) error {
	if err == io.EOF || err == ErrSkip {
		return err
	}
	readerErr, ok := err.(Error)
	if !ok {
		readerErr.Cause = err
	}
	readerErr.Begin = begin
	readerErr.End = rd.Position()
	return readerErr
}
