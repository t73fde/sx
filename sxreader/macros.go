//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxreader

import (
	"fmt"
	"io"
	"strings"

	"zettelstore.de/sx.fossil"
)

// UnmatchedDelimiter is a reader macro that signals the error of an
// unmatched delimiter, e.g. a closing parenthesis.
func UnmatchedDelimiter(rd *Reader, firstCh rune) (sx.Object, error) {
	return nil, rd.AnnotateError(delimiterError(firstCh), rd.Position())
}

// ReadComment is a reader macro that ignores everything until EOL.
func ReadComment(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	for {
		ch, err := rd.NextRune()
		if err != nil {
			return nil, rd.AnnotateError(err, beginPos)
		}
		if ch == '\n' {
			return nil, ErrSkip
		}
	}
}

func readNumber(rd *Reader, firstCh rune) (sx.Object, error) {
	beginPos := rd.Position()
	tok, err := rd.ReadToken(firstCh, rd.IsTerminal)
	if err != nil {
		return nil, rd.AnnotateError(err, beginPos)
	}
	num, err := sx.ParseInteger(tok)
	if err == nil {
		return num, nil
	}
	sym, err := rd.symFac.Make(tok)
	if err != nil {
		return nil, rd.AnnotateError(err, beginPos)
	}
	return sym, nil
}

func readDot(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	return nil, rd.AnnotateError(fmt.Errorf("'.' not allowed here"), beginPos)
}

func readSymbol(rd *Reader, firstCh rune) (sx.Object, error) {
	if rd.symFac == nil {
		return sx.Nil(), fmt.Errorf("symbol factory of reader not set")
	}
	beginPos := rd.Position()
	tok, err := rd.ReadToken(firstCh, rd.IsTerminal)
	if err != nil {
		return nil, rd.AnnotateError(err, beginPos)
	}
	sym, err := rd.symFac.Make(tok)
	if err != nil {
		return nil, rd.AnnotateError(err, beginPos)
	}
	return sym, nil
}

func readKeyword(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	tok, err := rd.ReadToken(0, rd.IsTerminal)
	if err != nil {
		return nil, rd.AnnotateError(err, beginPos)
	}
	return sx.Keyword(tok), nil
}

func readString(rd *Reader, _ rune) (sx.Object, error) {
	beginPos := rd.Position()
	var sb strings.Builder
	for {
		ch, err := rd.NextRune()
		if err != nil {
			if err == io.EOF {
				err = ErrEOF
			}
			return nil, rd.AnnotateError(err, beginPos)
		}

		if ch == '\\' {
			ch, err = rd.NextRune()
			if err != nil {
				if err == io.EOF {
					err = ErrEOF
				}
				return nil, rd.AnnotateError(err, beginPos)
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
				return nil, rd.AnnotateError(err, beginPos)
			}
		} else if ch == '"' {
			return sx.String(sb.String()), nil
		}

		sb.WriteRune(ch)
	}
}

func readRune(rd *Reader, numDigits int) (rune, error) {
	result := rune(0)
	for i := 0; i < numDigits; i++ {
		ch, err := rd.NextRune()
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

func readList(endCh rune) Macro {
	return func(rd *Reader, _ rune) (sx.Object, error) {
		beginPos := rd.Position()
		result, err := rd.readList(endCh)
		if err != nil {
			return nil, rd.AnnotateError(err, beginPos)
		}
		return result, nil
	}
}

func (rd *Reader) readList(endCh rune) (*sx.Pair, error) {
	objs := make([]sx.Object, 0, 32)

	var dotObj sx.Object
	hasDotObj := false

	curLength := uint(0)
	for {
		if curLength > rd.maxLength {
			return nil, ErrListTooLong
		}
		curLength++
		ch, err := rd.SkipSpace()
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
			ch2, err2 := rd.NextRune()
			if err2 == nil && rd.IsSpace(ch2) {
				dotObj, err2 = rd.Read()
				if err2 != nil {
					return nil, err2
				}
				hasDotObj = true
				ch3, err3 := rd.SkipSpace()
				if err3 != nil {
					return nil, err
				}
				if ch3 != endCh {
					return nil, fmt.Errorf("'%v' expected, but got %v", endCh, ch3)
				}
				break
			}
			rd.Unread(ch2)
		}
		rd.Unread(ch)

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
