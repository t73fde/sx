// -----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
// -----------------------------------------------------------------------------

package sxeval

import (
	"fmt"

	"zettelstore.de/sx.fossil"
)

// ParseFrame is a parsing environment.
type ParseFrame struct {
	sf      sx.SymbolFactory
	binding *Binding
	parser  Parser
}

func (pf *ParseFrame) Parse(obj sx.Object) (Expr, error) {
	return pf.parser.Parse(pf, obj)

}

func (pf *ParseFrame) ParseAgain(form sx.Object) error {
	return errParseAgain{pf: pf, form: form}
}

// errParseAgain is a non-error error signalling that the given form should be
// parsed again in the given environment.
type errParseAgain struct {
	pf   *ParseFrame
	form sx.Object
}

func (e errParseAgain) Error() string { return fmt.Sprintf("Again: %T/%v", e.form, e.form) }

func (pf *ParseFrame) MakeChildFrame(name string, baseSize int) *ParseFrame {
	return &ParseFrame{
		sf:      pf.sf,
		binding: MakeChildBinding(pf.binding, name, baseSize),
		parser:  pf.parser,
	}
}

func (pf *ParseFrame) SymbolFactory() sx.SymbolFactory { return pf.sf }

func (pf *ParseFrame) Bind(sym sx.Symbol, obj sx.Object) error { return pf.binding.Bind(sym, obj) }

func (pf *ParseFrame) Resolve(sym sx.Symbol) (sx.Object, bool) {
	return pf.binding.Resolve(sym)
}
func (pf *ParseFrame) Binding() *Binding { return pf.binding }
