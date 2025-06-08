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
	name string   // symbol name
	pkg  *Package // home package
}

// MakeSymbol creates a symbol from a string.
func MakeSymbol(name string) *Symbol {
	return CurrentPackage().MakeSymbol(name)
}

// GetValue return the string value of the symbol.
func (sym *Symbol) GetValue() string { return sym.name }

// IsNil may return true if a symbol pointer is nil.
func (sym *Symbol) IsNil() bool { return sym == nil }

// IsAtom always returns true because a symbol is an atomic value.
func (*Symbol) IsAtom() bool { return true }

// IsEqual compare the symbol with an object.
func (sym *Symbol) IsEqual(other Object) bool {
	if sym.IsNil() {
		return IsNil(other)
	}
	if IsNil(other) {
		return false
	}
	otherSy, isSymbol := other.(*Symbol)
	return isSymbol && sym.IsEqualSymbol(otherSy)
}

// IsEqualSymbol compare two symbols.
func (sym *Symbol) IsEqualSymbol(other *Symbol) bool { return sym == other }

// String returns the string representation.
func (sym *Symbol) String() string {
	var sb strings.Builder
	if _, err := sym.Print(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

// GoString returns the go string representation.
func (sym *Symbol) GoString() string { return sym.name }

// Print write the string representation to the given Writer.
func (sym *Symbol) Print(w io.Writer) (length int, err error) {
	if pkg := sym.pkg; pkg != CurrentPackage() {
		if sym.pkg != keywordPackage {
			length, err = io.WriteString(w, pkg.name)
			if err != nil {
				return length, err
			}
		}
		l, err := io.WriteString(w, ":")
		length += l
		if err != nil {
			return length, err
		}
	}
	// TODO: provide escape of symbol contains non-printable chars.
	l, err := io.WriteString(w, sym.name)
	return length + l, err
}

// GetSymbol returns the object as a symbol if possible.
func GetSymbol(obj Object) (*Symbol, bool) {
	if IsNil(obj) {
		return nil, false
	}
	sym, ok := obj.(*Symbol)
	return sym, ok
}

// Package returns the Package that created the symbol.
func (sym *Symbol) Package() *Package { return sym.pkg }
