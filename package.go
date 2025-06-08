//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sx

import (
	"fmt"
	"regexp"
	"sync"
)

// Package maps symbol names to Symbols.
type Package struct {
	name    string
	mx      sync.RWMutex
	symbols map[string]*Symbol
}

// IsNil may return true if a symbol pointer is nil.
func (pkg *Package) IsNil() bool { return pkg == nil }

// IsAtom always returns true because a symbol is an atomic value.
func (pkg *Package) IsAtom() bool { return pkg == nil }

// IsEqual compare the symbol with an object.
func (pkg *Package) IsEqual(other Object) bool {
	if pkg.IsNil() {
		return IsNil(other)
	}
	if IsNil(other) {
		return false
	}
	otherPkg, isPackage := other.(*Package)
	return isPackage && pkg.IsEqualPackage(otherPkg)
}

// IsEqualPackage compare two packages.
func (pkg *Package) IsEqualPackage(other *Package) bool { return pkg == other }

// String returns the string representation.
func (pkg *Package) String() string {
	return fmt.Sprintf("#<package:%s>", pkg.name)
}

// GoString returns the go string representation.
func (pkg *Package) GoString() string { return pkg.name }

// GetPackage returns the object as a package if possible.
func GetPackage(obj Object) (*Package, bool) {
	if IsNil(obj) {
		return nil, false
	}
	pkg, ok := obj.(*Package)
	return pkg, ok
}

var packageRegistry = map[string]*Package{}

// FindPackage returns the package with the given name.
func FindPackage(name string) *Package {
	if pkg, found := packageRegistry[name]; found {
		return pkg
	}
	return nil
}

var validPackageName = regexp.MustCompile("^[A-Za-z][0-9A-Za-z-]*$")

// MakePackage builds a new package.
func MakePackage(name string) (*Package, error) {
	if !validPackageName.MatchString(name) {
		return nil, fmt.Errorf("invalid package name: %q", name)
	}
	if pkg := FindPackage(name); pkg != nil {
		return nil, fmt.Errorf("package %q already made", name)
	}
	pkg := &Package{
		name:    name,
		symbols: map[string]*Symbol{},
	}
	packageRegistry[name] = pkg
	return pkg, nil
}

// MustMakePackage builds a new package and panics if something went wrong.
func MustMakePackage(name string) *Package {
	pkg, err := MakePackage(name)
	if err == nil {
		return pkg
	}
	panic(err)
}

// Predefined package names.
const (
	InitName    = "INIT"
	KeywordName = "KEYWORD"
)

var keywordPackage = MustMakePackage(KeywordName)

// KeywordPackage return the package that manages keyword symbols.
func KeywordPackage() *Package { return keywordPackage }

var initPackage = MustMakePackage(InitName)
var currentPackage = initPackage

// CurrentPackage returns the currently selected package.
func CurrentPackage() *Package { return currentPackage }

// MakeSymbol builds a symbol with the given string value. It tries to re-use
// symbols, so that symbols can be compared by their reference, not by their
// content.
func (pkg *Package) MakeSymbol(name string) *Symbol {
	sym := pkg.FindSymbol(name)
	if sym != nil {
		return sym
	}

	if name == "" {
		return nil
	}
	pkg.mx.Lock()
	sym, found := pkg.symbols[name]
	if !found {
		sym = &Symbol{pkg: pkg, name: name}
		pkg.symbols[name] = sym
	}
	pkg.mx.Unlock()
	return sym
}

// FindSymbol returns the symbol with the given name.
func (pkg *Package) FindSymbol(name string) *Symbol {
	if name == "" {
		return nil
	}
	pkg.mx.RLock()
	sym, found := pkg.symbols[name]
	pkg.mx.RUnlock()
	if found {
		return sym
	}
	return nil
}

// Size returns the number of symbols created by the package.
func (pkg *Package) Size() int {
	pkg.mx.RLock()
	result := len(pkg.symbols)
	pkg.mx.RUnlock()
	return result
}
