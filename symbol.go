//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sx

import (
	"fmt"
	"sync"
)

// Symbol represent a symbol value.
//
// Every symbol can store metadata with the help of Cons(). It can be retrieved using Assoc().
type Symbol struct {
	cname   string // Canonical name
	factory SymbolFactory
	alist   *Pair
}

// IsNil return true, if it is a nil symbol value.
func (sy *Symbol) IsNil() bool { return sy == nil }

func (sy *Symbol) IsAtom() bool { return true }

// IsEqual compare two objects.
//
// Two symbols are equal, if the are created by the same factory and have the same same.
func (sy *Symbol) IsEqual(other Object) bool {
	if sy == other {
		return true
	}
	if sy == nil {
		return IsNil(other)
	}
	if otherSy, ok := other.(*Symbol); ok {
		return sy.factory == otherSy.factory && sy.cname == otherSy.cname
	}
	return false
}

// String returns the Go string representation.
func (sy *Symbol) String() string { return sy.cname }

// Repr returns the object representation.
func (sy *Symbol) Repr() string { return sy.factory.ReprSymbol(sy) }

// Name returns the canonical name the symbol factory assigned to the symbol.
func (sy *Symbol) Name() string { return sy.cname }

// Factory returns the symbol factory that created this symbol.
func (sy *Symbol) Factory() SymbolFactory { return sy.factory }

// Cons a key to a value, to store metadata for the symbol.
func (sy *Symbol) Cons(key, obj Object) *Pair {
	p := Cons(key, obj)
	sy.alist = sy.alist.Cons(p)
	return p
}

// Assoc retrieves the formerly bound key/value pair.
func (sy *Symbol) Assoc(key Object) *Pair { return sy.alist.Assoc(key) }

// GetSymbol returns the object as a symbol if possible.
func GetSymbol(obj Object) (*Symbol, bool) {
	if IsNil(obj) {
		return nil, false
	}
	sym, ok := obj.(*Symbol)
	return sym, ok
}

// SymbolFactory creates new symbols and ensures locally that there is only one symbol with a given string value.
// It encapsulates case-sensitiveness, and is the only way to produce a valid symbol.
type SymbolFactory interface {
	// Make produces a singleton symbol from the given string.
	// If the string denotes an invalid name, an error will be returned.
	Make(string) (*Symbol, error)

	// MustMake will produce a singleton symbol and panic if that does not work.
	MustMake(string) *Symbol

	// IsValidName returns true, if given name is a valid name for a symbol.
	//
	// The empty string is always an invalid name.
	IsValidName(string) bool

	// Symbols returns a sequence of all symbols managed by this factory.
	Symbols() []*Symbol

	// ReprSymbol returns the factory-specific representation of the given symbol.
	ReprSymbol(*Symbol) string
}

// FindSymbolFactory searches for a symbol an returns its symbol factory.
//
// Typically, the search is done depth-first.
func FindSymbolFactory(obj Object) SymbolFactory {
	if IsNil(obj) {
		return nil
	}
	switch v := obj.(type) {
	case *Symbol:
		return v.Factory()
	case *Pair:
		for n := v; ; {
			if sf := FindSymbolFactory(n.Car()); sf != nil {
				return sf
			}
			cdr := n.cdr
			if IsNil(cdr) {
				break
			}
			tail, ok := cdr.(*Pair)
			if !ok {
				return FindSymbolFactory(cdr)
			}
			n = tail
		}
	}
	return nil
}

// mappedSymbolFactory create new symbols and ensures their uniqueness with a map.
type mappedSymbolFactory struct {
	mu      sync.RWMutex
	symbols map[string]*Symbol
}

// MakeMappedFactory creates a new factory.
func MakeMappedFactory(sizeHint int) SymbolFactory {
	if sizeHint < 7 {
		sizeHint = 7
	}
	return &mappedSymbolFactory{
		symbols: make(map[string]*Symbol, sizeHint),
	}
}

// IsValidName returns true if name is a vald symbol name.
func (*mappedSymbolFactory) IsValidName(s string) bool { return s != "" }

// Make creates a new symbol.
func (sf *mappedSymbolFactory) Make(s string) (*Symbol, error) {
	if !sf.IsValidName(s) {
		return nil, fmt.Errorf("symbol name not allowed: %q", s)
	}
	sf.mu.RLock()
	sym, found := sf.symbols[s]
	sf.mu.RUnlock()
	if found {
		return sym, nil
	}
	sym = &Symbol{
		cname:   s,
		factory: sf,
		alist:   nil,
	}
	sf.mu.Lock()
	sf.symbols[s] = sym
	sf.mu.Unlock()
	return sym, nil
}

// MustMake creates a new symbol from a given string.
func (sf *mappedSymbolFactory) MustMake(s string) *Symbol {
	sym, err := sf.Make(s)
	if err != nil {
		panic(err)
	}
	return sym
}

// Symbols returns a sequence of all symbols managed by this factory.
func (sf *mappedSymbolFactory) Symbols() []*Symbol {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	result := make([]*Symbol, 0, len(sf.symbols))
	for _, sym := range sf.symbols {
		result = append(result, sym)
	}
	return result
}

// ReprSymbol returns the string representation of the given symbol created by this factory.
func (sf *mappedSymbolFactory) ReprSymbol(sy *Symbol) string {
	if sf != sy.factory {
		panic("ReprSymbol called by symbol created with other factory")
	}
	return sy.cname
}
