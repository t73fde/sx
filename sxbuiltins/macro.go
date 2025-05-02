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
	"iter"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// DefMacroS parses a macro specfication.
var DefMacroS = sxeval.Special{
	Name: "defmacro",
	Fn: func(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (sxeval.Expr, error) {
		sym, le, err := parseDefProc(pe, args, bind)
		if err != nil {
			return nil, err
		}
		le.Type = macroType
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

// IsNil returns true if the object must be treated like a sx.Nil() object.
func (m *Macro) IsNil() bool { return m == nil }

// IsAtom returns true if the object is atomic.
func (m *Macro) IsAtom() bool { return m == nil }

// IsEqual returns true if the other object has the same content.
func (m *Macro) IsEqual(other sx.Object) bool { return m == other }

// String returns a string representation.
func (m *Macro) String() string { return "#<macro:" + m.Name + ">" }

// GoString returns a string representation to be used in Go code.
func (m *Macro) GoString() string { return m.String() }

// Parse transforms a macro call into its expanded form. Some kind of
// iterative expansion may happen.
func (m *Macro) Parse(pe *sxeval.ParseEnvironment, args *sx.Pair, bind *sxeval.Binding) (sxeval.Expr, error) {
	if err := m.Expand(pe, args, bind); err != nil {
		return nil, err
	}
	return sxeval.NilExpr, pe.ParseAgain(m.Env.Pop())
}

// Expand the macro in the given call.
func (m *Macro) Expand(_ *sxeval.ParseEnvironment, args *sx.Pair, _ *sxeval.Binding) error {
	numargs := 0
	arg := sx.Object(args)
	for {
		if sx.IsNil(arg) {
			break
		}
		pair, isPair := sx.GetPair(arg)
		if !isPair {
			return sx.ErrImproper{Pair: args}
		}
		m.Env.Push(pair.Car())
		numargs++
		arg = pair.Cdr()
	}

	proc := LexLambda{
		Binding: m.Binding,
		Name:    m.Name,
		Params:  m.Params,
		Rest:    m.Rest,
		Expr:    m.Expr,
	}
	return m.Env.ApplyMacro(proc.Name, &proc, numargs, m.Binding)
}

// -- Disassembler methods

// GetAsmCode returns sequence of pseudo instructions, if possible.
func (m *Macro) GetAsmCode() (iter.Seq[string], bool) { return sxeval.GetAsmCode(m.Expr) }

// Macroexpand0 implements one level of macro expansion.
//
// It is mostly used for debugging macros.
var Macroexpand0 = sxeval.Builtin{
	Name:     "macroexpand-0",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		lst, err := GetList(env.Pop(), 0)
		if err == nil && lst != nil {
			if sym, isSymbol := sx.GetSymbol(lst.Car()); isSymbol {
				if obj, found := bind.Resolve(sym); found {
					if macro, isMacro := obj.(*Macro); isMacro {
						return macro.Expand(env.MakeParseEnvironment(), lst.Tail(), bind)
					}
				}
			}
		}
		return err
	},
}
