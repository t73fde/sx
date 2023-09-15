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
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// DefMacroS parses a macro specfication and assigns it to a value.
func DefMacroS(frame *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	if args == nil {
		return nil, sxeval.ErrNoArgs
	}
	sym, isSymbol := sx.GetSymbol(args.Car())
	if !isSymbol {
		return nil, fmt.Errorf("not a symbol: %T/%v", args.Car(), args.Car())
	}
	args = args.Tail()
	if args == nil {
		return nil, fmt.Errorf("parameter spec and body missing")
	}
	le, err := ParseProcedure(frame, sx.Repr(sym), args.Car(), args.Cdr())
	if err != nil {
		return nil, err
	}
	me := &MacroExpr{
		Name:    le.Name,
		Params:  le.Params,
		Rest:    le.Rest,
		ExprSeq: le.ExprSeq,
	}
	return &DefineExpr{Sym: sym, Val: me}, nil
}

type MacroExpr struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	ExprSeq
}

func (me *MacroExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	me.ExprSeq.Rework(rf)
	return me
}
func (me *MacroExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	return &Macro{
		Frame:   frame,
		PFrame:  frame.MakeParseFrame(),
		Name:    me.Name,
		Params:  me.Params,
		Rest:    me.Rest,
		ExprSeq: me.ExprSeq,
	}, nil
}
func (me *MacroExpr) IsEqual(other sxeval.Expr) bool {
	if me == other {
		return true
	}
	if otherM, ok := other.(*MacroExpr); ok && otherM != nil {
		return me.Name == otherM.Name &&
			sxeval.EqualSymbolSlice(me.Params, otherM.Params) &&
			me.Rest.IsEqual(otherM.Rest) &&
			me.ExprSeq.IsEqual(&otherM.ExprSeq)
	}
	return false
}
func (me *MacroExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{MACRO ")
	if err != nil {
		return length, err
	}
	l, err := io.WriteString(w, me.Name)
	length += l
	if err != nil {
		return length, err
	}
	for _, p := range me.Params {
		l2, err2 := fmt.Fprintf(w, " %v", p)
		length += l2
		if err2 != nil {
			return length, err2
		}
	}
	if me.Rest == nil {
		l, err = io.WriteString(w, " none ")
	} else {
		l, err = fmt.Fprintf(w, " %v ", me.Rest)
	}
	length += l
	if err != nil {
		return length, err
	}
	l, err = me.ExprSeq.Print(w)
	length += l
	return length, err
}

// Macro represents the macro definition form.
type Macro struct {
	Frame  *sxeval.Frame
	PFrame *sxeval.ParseFrame
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	ExprSeq
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
			m.ExprSeq.IsEqual(&otherM.ExprSeq)
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
		PFrame:  m.PFrame,
		Name:    m.Name,
		Params:  m.Params,
		Rest:    m.Rest,
		ExprSeq: m.ExprSeq,
	}
	return m.Frame.MakeCalleeFrame().Call(&proc, macroArgs)
}

// MacroExpand0 implements one level of macro expansion.
//
// It is mostly used for debugging macros.
func MacroExpand0(frame *sxeval.Frame, args []sx.Object) (sx.Object, error) {
	err := CheckArgs(args, 1, 1)
	lst, err := GetList(err, args, 0)
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
}
