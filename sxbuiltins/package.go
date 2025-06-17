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

package sxbuiltins

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// CurrentPackage returns the current active default package.
var CurrentPackage = sxeval.Builtin{
	Name:     "current-package",
	MinArity: 0,
	MaxArity: 0,
	Fn0: func(*sxeval.Environment, *sxeval.Frame) (sx.Object, error) {
		return sx.CurrentPackage(), nil
	},
}

// PackageList returns a list of all defined packages.
var PackageList = sxeval.Builtin{
	Name:     "package-list",
	MinArity: 0,
	MaxArity: 0,
	Fn0: func(*sxeval.Environment, *sxeval.Frame) (sx.Object, error) {
		var lb sx.ListBuilder
		for pkg := range sx.AllPackages() {
			lb.Add(pkg)
		}
		return lb.List(), nil
	},
}

// FindPackage returns the package with the given name.
var FindPackage = sxeval.Builtin{
	Name:     "find-package",
	MinArity: 1,
	MaxArity: 1,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		s, err := GetString(arg, 0)
		if err != nil {
			return nil, err
		}
		return sx.FindPackage(s.GetValue()), nil
	},
}

// PackageSymbols returns a list of all symbols managed by the given package.
// If no package is specified, the current active package is used.
var PackageSymbols = sxeval.Builtin{
	Name:     "package-symbols",
	MinArity: 0,
	MaxArity: 1,
	Fn0: func(*sxeval.Environment, *sxeval.Frame) (sx.Object, error) {
		return packageSymbols(sx.CurrentPackage())
	},
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		pkg, err := GetPackage(arg, 0)
		if err != nil {
			return nil, err
		}
		return packageSymbols(pkg)
	},
}

func packageSymbols(pkg *sx.Package) (sx.Object, error) {
	var lb sx.ListBuilder
	for sym := range pkg.AllSymbols() {
		lb.Add(sym)
	}
	return lb.List(), nil
}
