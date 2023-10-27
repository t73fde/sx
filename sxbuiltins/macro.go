//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// DefMacroS parses a macro specfication.
var DefMacroS = sxeval.Special{
	Name: "defmacro",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
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
	Frame  *sxeval.Frame
	PFrame *sxeval.ParseFrame
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Front  []sxeval.Expr
	Last   sxeval.Expr
}

func (m *Macro) IsNil() bool  { return m == nil }
func (m *Macro) IsAtom() bool { return m == nil }
func (m *Macro) IsEqual(other sx.Object) bool {
	if m == other {
		return true
	}
	if m.IsNil() {
		return sx.IsNil(other)
	}
	if otherM, ok := other.(*Macro); ok {
		// Don't compare Name, because they are always different, but that does not matter.
		return m.PFrame.IsEqual(otherM.PFrame) &&
			sxeval.EqualSymbolSlice(m.Params, otherM.Params) &&
			m.Rest.IsEqual(otherM.Rest) &&
			sxeval.EqualExprSlice(m.Front, otherM.Front) &&
			m.Last.IsEqual(otherM.Last)
	}
	return false
}
func (m *Macro) String() string { return m.Repr() }
func (m *Macro) Repr() string   { return sx.Repr(m) }
func (m *Macro) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<macro:", m.Name, ">")
}
func (m *Macro) Parse(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	form, err := m.Expand(pf, args)
	if err != nil {
		return nil, err
	}
	return nil, pf.ParseAgain(form)
}

func (m *Macro) Expand(_ *sxeval.ParseFrame, args *sx.Pair) (sx.Object, error) {
	var macroArgs []sx.Object
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
		PFrame: m.PFrame,
		Name:   m.Name,
		Params: m.Params,
		Rest:   m.Rest,
		Front:  m.Front,
		Last:   m.Last,
	}
	return m.Frame.MakeCalleeFrame().Call(&proc, macroArgs)
}

// MacroExpand0 implements one level of macro expansion.
//
// It is mostly used for debugging macros.
var Macroexpand0 = sxeval.Builtin{
	Name:     "macroexpand-0",
	MinArity: 1,
	MaxArity: 1,
	IsPure:   false,
	Fn: func(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
		lst, err := GetList(args, 0)
		if err == nil && lst != nil {
			if sym, isSymbol := sx.GetSymbol(lst.Car()); isSymbol {
				if obj, found := frame.Resolve(sym); found {
					if macro, isMacro := obj.(*Macro); isMacro {
						return macro.Expand(frame.MakeParseFrame(), lst.Tail())
					}
				}
			}
		}
		return lst, err
	},
}
