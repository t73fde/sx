//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sxreader

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"t73f.de/r/sx"
)

// unmatchedDelimiter is a reader macro that signals the error of an
// unmatched delimiter, e.g. a closing parenthesis.
func unmatchedDelimiter(rd *Reader, firstCh rune) (sx.Object, error) {
	return nil, rd.annotateError(delimiterError(firstCh), rd.Position())
}

// readComment is a reader macro that ignores everything until EOL.
func readComment(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	for {
		ch, err := rd.nextRune()
		if err != nil {
			return nil, rd.annotateError(err, beginPos)
		}
		if ch == '\n' {
			return nil, ErrSkip
		}
	}
}

func readNumber(rd *Reader, firstCh rune) (sx.Object, error) {
	beginPos := rd.Position()
	tok, err := rd.readToken(firstCh, rd.isTerminal)
	if err != nil {
		return nil, rd.annotateError(err, beginPos)
	}
	num, err := sx.ParseInteger(tok)
	if err == nil {
		return num, nil
	}
	sym := sx.MakeSymbol(tok)
	return sym, nil
}

func readHash(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	return nil, rd.annotateError(fmt.Errorf("'#' not allowed here"), beginPos)
}

func readDot(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	return nil, rd.annotateError(fmt.Errorf("'.' not allowed here"), beginPos)
}

func readSymbol(rd *Reader, firstCh rune) (sx.Object, error) {
	beginPos := rd.Position()
	tok, err := rd.readToken(firstCh, rd.isTerminal)
	if err != nil {
		return nil, rd.annotateError(err, beginPos)
	}
	sym := sx.MakeSymbol(tok)
	return sym, nil
}

func readString(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	var sb strings.Builder
	for {
		ch, err := rd.nextRune()
		if err != nil {
			if err == io.EOF {
				err = ErrEOF
			}
			return nil, rd.annotateError(err, beginPos)
		}

		if ch == '\\' {
			ch, err = rd.nextRune()
			if err != nil {
				if err == io.EOF {
					err = ErrEOF
				}
				return nil, rd.annotateError(err, beginPos)
			}
			switch ch {
			case '"':
			case '\\':
			case 'n':
				ch = '\n'
			case 'r':
				ch = '\r'
			case 't':
				ch = '\t'
			case 'x':
				ch, err = readRune(rd, 2)
			case 'u':
				ch, err = readRune(rd, 4)
			case 'U':
				ch, err = readRune(rd, 6)
			}
			if err != nil {
				if err == io.EOF {
					err = ErrEOF
				}
				return nil, rd.annotateError(err, beginPos)
			}
		} else if ch == '"' {
			return sx.MakeString(sb.String()), nil
		}

		sb.WriteRune(ch)
	}
}

func readRune(rd *Reader, numDigits int) (rune, error) {
	result := rune(0)
	for range numDigits {
		ch, err := rd.nextRune()
		if err != nil {
			return result, err
		}
		switch ch {
		case '0':
			result <<= 4
		case '1':
			result = (result << 4) + 1
		case '2':
			result = (result << 4) + 2
		case '3':
			result = (result << 4) + 3
		case '4':
			result = (result << 4) + 4
		case '5':
			result = (result << 4) + 5
		case '6':
			result = (result << 4) + 6
		case '7':
			result = (result << 4) + 7
		case '8':
			result = (result << 4) + 8
		case '9':
			result = (result << 4) + 9
		case 'A', 'a':
			result = (result << 4) + 10
		case 'B', 'b':
			result = (result << 4) + 11
		case 'C', 'c':
			result = (result << 4) + 12
		case 'D', 'd':
			result = (result << 4) + 13
		case 'E', 'e':
			result = (result << 4) + 14
		case 'F', 'f':
			result = (result << 4) + 15
		default:
			return result, fmt.Errorf("no hex digit found: %c/%d", ch, ch)
		}
	}
	return result, nil
}

func readList(endCh rune) macroFn {
	return func(rd *Reader, _ rune) (sx.Object, error) {
		beginPos := rd.Position()
		result, err := rd.readList(endCh)
		if err != nil {
			return nil, rd.annotateError(err, beginPos)
		}
		return result, nil
	}
}

func (rd *Reader) readList(endCh rune) (*sx.Pair, error) {
	objs := make(sx.Vector, 0, 32)

	var dotObj sx.Object
	hasDotObj := false

	curLength := uint(0)
	for {
		if maxLength := rd.maxLength; maxLength > 0 {
			if curLength > maxLength {
				return nil, ErrListTooLong
			}
			curLength++
		}
		ch, err := rd.readListCh()
		if err != nil {
			if err == io.EOF {
				return nil, ErrEOF
			}
			return nil, err
		}

		if ch == endCh {
			break
		}
		if ch == '.' {
			ch2, err2 := rd.nextRune()
			if err2 == nil && isSpace(ch2) {
				dotObj, err2 = rd.Read()
				if err2 != nil {
					return nil, err2
				}
				hasDotObj = true
				ch3, err3 := rd.skipSpace()
				if err3 != nil {
					return nil, err
				}
				if ch3 != endCh {
					return nil, fmt.Errorf("'%c' (%v) expected, but got '%c' (%v)", endCh, endCh, ch3, ch3)
				}
				break
			}
			rd.unreadRunes(ch2)
		}
		rd.unreadRunes(ch)

		val, err := rd.Read()
		if err != nil {
			return nil, err
		}
		objs = append(objs, val)
	}
	if hasDotObj {
		lenV := len(objs)
		if lenV == 0 {
			return sx.Cons(sx.Nil(), dotObj), nil
		}
		lenV--
		result := sx.Cons(objs[lenV], dotObj)
		for i := lenV - 1; i >= 0; i-- {
			result = result.Cons(objs[i])
		}
		return result, nil

	}
	return sx.MakeList(objs...), nil
}
func (rd *Reader) readListCh() (rune, error) {
	for {
		ch, err := rd.nextRune()
		if err != nil {
			return 0, err
		}
		if isSpace(ch) {
			continue
		}
		if ch != chComment {
			return ch, nil
		}
		_, err = readComment(rd, ch)
		if err != nil && !errors.Is(err, ErrSkip) {
			return 0, err
		}
	}
}

func readQuote(rd *Reader, _ rune) (sx.Object, error) {
	obj, err := rd.Read()
	if err == nil {
		return sx.Nil().Cons(obj).Cons(sx.SymbolQuote), nil
	}
	if err == io.EOF {
		return obj, ErrEOF
	}
	return obj, err
}

func readQuasiquote(rd *Reader, _ rune) (sx.Object, error) {
	obj, err := rd.Read()
	if err == nil {
		return sx.Nil().Cons(obj).Cons(sx.SymbolQuasiquote), nil
	}
	if err == io.EOF {
		return obj, ErrEOF
	}
	return obj, err
}

func readUnquote(rd *Reader, _ rune) (sx.Object, error) {
	ch, err := rd.nextRune()
	if err != nil {
		return nil, err
	}
	sym := sx.SymbolUnquoteSplicing
	if ch != '@' {
		rd.unreadRunes(ch)
		sym = sx.SymbolUnquote
	}
	obj, err := rd.Read()
	if err == nil {
		return sx.Nil().Cons(obj).Cons(sym), nil
	}
	return obj, err
}
