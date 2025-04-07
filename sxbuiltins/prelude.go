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
	_ "embed"
	"io"
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
	"t73f.de/r/sx/sxreader"
)

//go:embed prelude.sxn
var prelude string

// LoadPrelude reads and evaluates the standard prelude. In addition some symbols
// (NIL, T, UNDEFINED) are bound, as well as needed builtins and special forms.
func LoadPrelude(root *sxeval.Binding) error {
	if err := root.Bind(sx.MakeSymbol("UNDEFINED"), sx.MakeUndefined()); err != nil {
		return err
	}
	if err := root.Bind(sx.MakeSymbol("NIL"), sx.Nil()); err != nil {
		return err
	}
	if err := root.Bind(sx.MakeSymbol("T"), sx.MakeSymbol("T")); err != nil {
		return err
	}

	if err := sxeval.BindSpecials(root, &DefMacroS, &QuasiquoteS, &IfS, &LetS, &BeginS); err != nil {
		return err
	}
	if err := sxeval.BindBuiltins(root, &NullP, &Caar, &Cadar, &Cdr, &Car, &Equal); err != nil {
		return err
	}

	rd := sxreader.MakeReader(strings.NewReader(prelude))
	env := sxeval.MakeEnvironment()
	for {
		form, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err = env.Eval(form, root)
		if err != nil {
			return err
		}
	}
}
