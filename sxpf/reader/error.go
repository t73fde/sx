//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package reader

import (
	"errors"
	"fmt"
)

// ErrSkip signals a caller that the read value should be skipped.
// Typically, the read method is called again.
// Mostly used to implement comments at the lexical level.
var ErrSkip = errors.New("skip s-expression")

// ErrEOF is returned when reader ends prematurely to indicate
// that more data is needed to complete the current s-expression.
var ErrEOF = errors.New("unexpected EOF while reading")

// ErrNumberFormat is returned when a reader macro encounters a invalid number.
var ErrNumberFormat = errors.New("invalid number format")

// ErrPairFormat signals an invalid pair.
var ErrPairFormat = errors.New("invalid pair format")

// Error is returned when reading fails due to some issue.
// Use errors.Is() with Cause to check for specific underlying errors.
type Error struct {
	Cause error
	Begin Position
	End   Position
}

// Is returns true if the other error is same as the cause of this error.
func (e Error) Is(other error) bool { return errors.Is(e.Cause, other) }

// Unwrap returns the underlying cause of the error.
func (e Error) Unwrap() error { return e.Cause }

func (e Error) Error() string {
	return fmt.Sprintf("ReaderError %d-%d: %v", e.End.Line, e.End.Col, e.Cause)
}

func (e Error) Format(s fmt.State, _ rune) {
	if s.Flag('#') {
		fmt.Fprintf(s, "File \"%s\" line %d, column %d to line %d, column %d\n",
			e.Begin.Name, e.Begin.Line, e.Begin.Col, e.End.Line, e.End.Col)
	}
	fmt.Fprint(s, e.Error())
}

type delimiterError rune

func (r delimiterError) Error() string { return fmt.Sprintf("unmatched delimiter '%c'", r) }
