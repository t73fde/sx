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
	"sync"
)

// Symbol represent a symbol value.
type Symbol struct {
	fac *SymbolFactory
	val string
}

// MakeSymbol creates a symbol from a string.
func MakeSymbol(val string) *Symbol {
	return defaultSymbolFactory.MakeSymbol(val)
}

// GetValue return the string value of the symbol.
func (sym *Symbol) GetValue() string { return sym.val }

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
func (sym *Symbol) GoString() string { return sym.val }

// Print write the string representation to the given Writer.
func (sym *Symbol) Print(w io.Writer) (int, error) {
	// TODO: provide escape of symbol contains non-printable chars.
	return io.WriteString(w, sym.val)
}

// GetSymbol returns the object as a symbol if possible.
func GetSymbol(obj Object) (*Symbol, bool) {
	if IsNil(obj) {
		return nil, false
	}
	sym, ok := obj.(*Symbol)
	return sym, ok
}

// Factory returns the SymbolFactory that created the symbol.
func (sym *Symbol) Factory() *SymbolFactory { return sym.fac }

// SymbolFactory creates an interned symbol.
type SymbolFactory struct {
	parent  *SymbolFactory
	mx      sync.RWMutex
	symbols map[string]*Symbol
}

var defaultSymbolFactory = SymbolFactory{}

// DefaultSymbolFactory returns the symbol factory used by MakeSymbol.
func DefaultSymbolFactory() *SymbolFactory { return &defaultSymbolFactory }

// MakeSymbol builds a symbol with the given string value. It tries to re-use
// symbols, so that symbols can be compared by their reference, not by their
// content.
func (sf *SymbolFactory) MakeSymbol(val string) (sym *Symbol) {
	if val == "" {
		return nil
	}

	var found bool
	for factory := sf; factory != nil; factory = factory.parent {
		factory.mx.RLock()
		if len(factory.symbols) > 0 {
			sym, found = factory.symbols[val]
		}
		factory.mx.RUnlock()
		if found {
			return sym
		}
	}

	sf.mx.Lock()
	if len(sf.symbols) > 0 {
		sym, found = sf.symbols[val]
	}
	if !found {
		sym = &Symbol{fac: sf, val: val}
		if len(sf.symbols) == 0 {
			sf.symbols = map[string]*Symbol{val: sym}
		} else {
			sf.symbols[val] = sym
		}
	}
	sf.mx.Unlock()
	return sym
}

// Size returns the number of symbols created by the factory.
func (sf *SymbolFactory) Size() int {
	sf.mx.RLock()
	result := len(sf.symbols)
	sf.mx.RUnlock()
	return result
}

// NewChild makes a new child factory.
func (sf *SymbolFactory) NewChild() *SymbolFactory {
	return &SymbolFactory{
		parent: sf,
	}
}

// MoveSymbols transfers all symbols of this factory to its parent factory.
func (sf *SymbolFactory) MoveSymbols() {
	parent := sf.parent
	if parent == nil {
		panic("no parent symbol factory")
	}
	sf.mx.Lock()
	parent.mx.Lock()
	var errVal string
	if len(sf.symbols) > 0 {
		if len(parent.symbols) == 0 {
			parent.symbols = make(map[string]*Symbol, len(sf.symbols))
		}
		for val, sym := range sf.symbols {
			if parentSym, found := parent.symbols[val]; found {
				if parentSym != sym {
					errVal = val
					break
				}
				continue
			}
			parent.symbols[val] = sym
		}
		sf.symbols = nil
	}
	parent.mx.Unlock()
	sf.mx.Unlock()
	if errVal != "" {
		panic(errVal + " already in parent")
	}
}
