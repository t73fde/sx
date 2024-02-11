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

package sx

import (
	"io"
	"strings"
)

// Symbol represent a symbol value.
type Symbol struct {
	val string
}

// MakeSymbol creates a symbol from a string.
func MakeSymbol(val string) *Symbol {
	if val == "" {
		return nil
	}
	return &Symbol{val}
}

// GetValue return the string value of the symbol.
func (sym *Symbol) GetValue() string { return sym.val }

// IsNil return false, since a symbol is never nil.
func (sy *Symbol) IsNil() bool { return sy == nil }

func (*Symbol) IsAtom() bool { return true }

// IsEqual compare two symbols.
func (sy *Symbol) IsEqual(other Object) bool {
	if sy == nil {
		return IsNil(other)
	}
	if IsNil(other) {
		return false
	}
	otherSy, isSymbol := other.(*Symbol)
	return isSymbol && (sy == otherSy || sy.val == otherSy.val)
}

// String returns the string representation.
func (sy *Symbol) String() string {
	var sb strings.Builder
	sy.Print(&sb)
	return sb.String()
}

// GoString returns the go string representation.
func (sy *Symbol) GoString() string { return sy.val }

// Print write the string representation to the given Writer.
func (sy *Symbol) Print(w io.Writer) (int, error) {
	// TODO: provide escape of symbol contains non-printable chars.
	return io.WriteString(w, sy.val)
}

// GetSymbol returns the object as a symbol if possible.
func GetSymbol(obj Object) (*Symbol, bool) {
	if IsNil(obj) {
		return nil, false
	}
	sym, ok := obj.(*Symbol)
	return sym, ok
}
