//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sx

// Names of quotation symbols.
//
// Used in packages sxbuiltins, sxreader.
var (
	SymbolQuote           = MakeSymbol("quote")
	SymbolQuasiquote      = MakeSymbol("quasiquote")
	SymbolUnquote         = MakeSymbol("unquote")
	SymbolUnquoteSplicing = MakeSymbol("unquote-splicing")
)

// SymbolList is the symbol of the (list ...) function.
var SymbolList = MakeSymbol("list")
