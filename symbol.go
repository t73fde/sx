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

import "io"

// Symbol represent a symbol value.
type Symbol string

// IsNil return false, since a symbol is never nil.
func (sy Symbol) IsNil() bool { return false }

func (sy Symbol) IsAtom() bool { return true }

// IsEqual compare two symbols.
func (sy Symbol) IsEqual(other Object) bool {
	otherSy, isSymbol := other.(Symbol)
	return isSymbol && string(sy) == string(otherSy)
}

// String returns the Go string representation.
func (sy Symbol) String() string { return string(sy) }

// Print write the string representation to the given Writer.
func (sy Symbol) Print(w io.Writer) (int, error) {
	// TODO: provide escape of symbol contains non-printable chars.
	return io.WriteString(w, string(sy))
}

// Name returns the canonical name.
func (sy Symbol) Name() string { return string(sy) }

// GetSymbol returns the object as a symbol if possible.
func GetSymbol(obj Object) (Symbol, bool) {
	if IsNil(obj) {
		return "", false
	}
	sym, ok := obj.(Symbol)
	return sym, ok
}
