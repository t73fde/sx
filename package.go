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
	"errors"
	"fmt"
	"sync"
)

// Package maps symbol names to Symbols.
type Package struct {
	name    string
	mx      sync.RWMutex
	symbols map[string]*Symbol
}

var packageRegistry = map[string]*Package{}

// FindPackage returns the Package with the given name.
func FindPackage(name string) *Package {
	if pkg, found := packageRegistry[name]; found {
		return pkg
	}
	return nil
}

// MakePackage builds a new package.
func MakePackage(name string) (*Package, error) {
	if name == "" {
		return nil, errors.New("no name given for new package")
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

var initPackage = MustMakePackage("init")
var currentPackage = initPackage

// CurrentPackage returns the currently selected package.
func CurrentPackage() *Package { return currentPackage }

// MakeSymbol builds a symbol with the given string value. It tries to re-use
// symbols, so that symbols can be compared by their reference, not by their
// content.
func (pkg *Package) MakeSymbol(name string) *Symbol {
	if name == "" {
		return nil
	}

	sym := pkg.FindSymbol(name)
	if sym != nil {
		return sym
	}

	pkg.mx.Lock()
	sym, found := pkg.symbols[name]
	if !found {
		sym = &Symbol{pkg: pkg, val: name}
		pkg.symbols[name] = sym
	}
	pkg.mx.Unlock()
	return sym
}

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
