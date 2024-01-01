//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sx

// Names of quotation symbols.
//
// Used in packages sxbuiltins, sxreader.
const (
	QuoteSymbol           = Symbol("quote")
	QuasiquoteSymbol      = Symbol("quasiquote")
	UnquoteSymbol         = Symbol("unquote")
	UnquoteSplicingSymbol = Symbol("unquote-splicing")
)

// ListName is the name of the list function.
const ListName = "list"
