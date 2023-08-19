// -----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
// -----------------------------------------------------------------------------

package sxeval

import (
	"fmt"

	"zettelstore.de/sx.fossil"
)

// ParseFrame is a parsing environment.
type ParseFrame struct {
	engine *Engine
	env    Environment
	parser Parser
}

func (frame *ParseFrame) IsEql(other *ParseFrame) bool {
	if frame == other {
		return true
	}
	if frame == nil || other == nil {
		return false
	}
	if frame.engine != other.engine {
		return false
	}
	return frame.env.IsEql(other.env)
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
		engine: pf.engine,
		env:    MakeChildEnvironment(pf.env, name, baseSize),
		parser: pf.parser,
	}
}

func (pf *ParseFrame) Bind(sym *sx.Symbol, obj sx.Object) error {
	env, err := pf.env.Bind(sym, obj)
	pf.env = env
	return err
}

func (pf *ParseFrame) Resolve(sym *sx.Symbol) (sx.Object, bool) {
	return Resolve(pf.env, sym)
}
func (pf *ParseFrame) Environment() Environment { return pf.env }
