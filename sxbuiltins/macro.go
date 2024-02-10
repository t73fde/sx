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

package sxbuiltins

import (
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// DefMacroS parses a macro specfication.
var DefMacroS = sxeval.Special{
	Name: "defmacro",
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		sym, le, err := parseDefProc(pf, args)
		if err != nil {
			return nil, err
		}
		le.IsMacro = true
		return &DefineExpr{Sym: sym, Val: le}, nil
	},
}

// Macro represents the macro definition form.
type Macro struct {
	Env     *sxeval.Environment
	Binding *sxeval.Binding
	Name    string
	Params  []*sx.Symbol
	Rest    *sx.Symbol
	Expr    sxeval.Expr
}

func (m *Macro) IsNil() bool                  { return m == nil }
func (m *Macro) IsAtom() bool                 { return m == nil }
func (m *Macro) IsEqual(other sx.Object) bool { return m == other }
func (m *Macro) String() string               { return "#<macro:" + m.Name + ">" }
func (m *Macro) GoString() string             { return m.String() }

func (m *Macro) Parse(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
	form, err := m.Expand(pf, args)
	if err != nil {
		return nil, err
	}
	return nil, pf.ParseAgain(form)
}

func (m *Macro) Expand(_ *sxeval.ParseEnvironment, args *sx.Pair) (sx.Object, error) {
	var macroArgs sx.Vector
	arg := sx.Object(args)
	for {
		if sx.IsNil(arg) {
			break
		}
		pair, isPair := sx.GetPair(arg)
		if !isPair {
			return nil, sx.ErrImproper{Pair: args}
		}
		macroArgs = append(macroArgs, pair.Car())
		arg = pair.Cdr()
	}

	proc := Procedure{
		Binding: m.Binding,
		Name:    m.Name,
		Params:  m.Params,
		Rest:    m.Rest,
		Expr:    m.Expr,
	}
	return m.Env.Call(&proc, macroArgs)
}

// MacroExpand0 implements one level of macro expansion.
//
// It is mostly used for debugging macros.
var Macroexpand0 = sxeval.Builtin{
	Name:     "macroexpand-0",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		lst, err := GetList(args, 0)
		if err == nil && lst != nil {
			if sym, isSymbol := sx.GetSymbol(lst.Car()); isSymbol {
				if obj, found := env.Resolve(sym); found {
					if macro, isMacro := obj.(*Macro); isMacro {
						return macro.Expand(env.MakeParseFrame(), lst.Tail())
					}
				}
			}
		}
		return lst, err
	},
}
