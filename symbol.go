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
	"fmt"
	"io"
)

// Symbol represent a symbol value.
type Symbol string

// IsNil return true, if it is a nil symbol value.
func (sy Symbol) IsNil() bool { return false }

func (sy Symbol) IsAtom() bool { return true }

// IsEqual compare two objects.
//
// Two symbols are equal, if the are created by the same factory and have the same same.
func (sy Symbol) IsEqual(other Object) bool {
	otherSy, isSymbol := other.(Symbol)
	return isSymbol && string(sy) == string(otherSy)
}

// String returns the Go string representation.
func (sy Symbol) String() string { return string(sy) }

// Repr returns the object representation.
func (sy Symbol) Repr() string { return Repr(sy) }

// Print write the string representation to the given Writer.
func (sy Symbol) Print(w io.Writer) (int, error) {
	return io.WriteString(w, string(sy))
}

// Name returns the canonical name the symbol factory assigned to the symbol.
func (sy Symbol) Name() string { return string(sy) }

// Factory returns the symbol factory that created this symbol.
func (sy Symbol) Factory() SymbolFactory { return nil }

// GetSymbol returns the object as a symbol if possible.
func GetSymbol(obj Object) (Symbol, bool) {
	if IsNil(obj) {
		return "", false
	}
	sym, ok := obj.(Symbol)
	return sym, ok
}

// SymbolFactory creates new symbols and ensures locally that there is only one symbol with a given string value.
// It encapsulates case-sensitiveness, and is the only way to produce a valid symbol.
type SymbolFactory interface {
	// Make produces a singleton symbol from the given string.
	// If the string denotes an invalid name, an error will be returned.
	Make(string) (Symbol, error)

	// MustMake will produce a singleton symbol and panic if that does not work.
	MustMake(string) Symbol

	// IsValidName returns true, if given name is a valid name for a symbol.
	//
	// The empty string is always an invalid name.
	IsValidName(string) bool

	// Symbols returns a sequence of all symbols managed by this factory.
	Symbols() []Symbol

	// ReprSymbol returns the factory-specific representation of the given symbol.
	ReprSymbol(Symbol) string
}

// FindSymbolFactory searches for a symbol an returns its symbol factory.
//
// Typically, the search is done depth-first.
func FindSymbolFactory(obj Object) SymbolFactory {
	if IsNil(obj) {
		return nil
	}
	return MakeMappedFactory(1)
}

// mappedSymbolFactory create new symbols and ensures their uniqueness with a map.
type mappedSymbolFactory struct{}

// MakeMappedFactory creates a new factory.
func MakeMappedFactory(sizeHint int) SymbolFactory {
	if sizeHint < 7 {
		sizeHint = 7
	}
	return &mappedSymbolFactory{}
}

// IsValidName returns true if name is a vald symbol name.
func (*mappedSymbolFactory) IsValidName(s string) bool { return s != "" }

// Make creates a new symbol.
func (sf *mappedSymbolFactory) Make(s string) (Symbol, error) {
	if !sf.IsValidName(s) {
		return "", fmt.Errorf("symbol name not allowed: %q", s)
	}
	return Symbol(s), nil
}

// MustMake creates a new symbol from a given string.
func (sf *mappedSymbolFactory) MustMake(s string) Symbol {
	sym, err := sf.Make(s)
	if err != nil {
		panic(err)
	}
	return sym
}

// Symbols returns a sequence of all symbols managed by this factory.
func (sf *mappedSymbolFactory) Symbols() []Symbol { return nil }

// ReprSymbol returns the string representation of the given symbol created by this factory.
func (sf *mappedSymbolFactory) ReprSymbol(sy Symbol) string {
	return sy.Repr()
}
