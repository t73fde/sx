//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package macro

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins/callable"
	"zettelstore.de/sx.fossil/sxbuiltins/define"
	"zettelstore.de/sx.fossil/sxeval"
)

// MacroS parses a macro specification.
//
// Syntactically, it is the same as a procedure specification (aka lambda).
func MacroS(frame *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
	le, err := callable.LambdaS(frame, args)
	if err != nil {
		return nil, err
	}
	return makeMacroExpr(le.(*callable.LambdaExpr)), err
}

func makeMacroExpr(le *callable.LambdaExpr) sxeval.Expr {
	return &MacroExpr{
		Name:   le.Name,
		Params: le.Params,
		Rest:   le.Rest,
		Front:  le.Front,
		Last:   le.Last,
	}
}

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
	le, err := callable.ParseProcedure(frame, sx.Repr(sym), args.Car(), args.Cdr())
	if err != nil {
		return nil, err
	}
	return &define.DefineExpr{Sym: sym, Val: makeMacroExpr(le)}, nil
}

type MacroExpr struct {
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Front  []sxeval.Expr // all expressions, but the last
	Last   sxeval.Expr
}

func (me *MacroExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	return &Macro{
		Frame:  frame,
		PFrame: frame.MakeParseFrame(),
		Name:   me.Name,
		Params: me.Params,
		Rest:   me.Rest,
		Front:  me.Front,
		Last:   me.Last,
	}, nil
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
	l, err = sxeval.PrintFrontLast(w, me.Front, me.Last)
	length += l
	return length, err
}
func (me *MacroExpr) Rework(ro *sxeval.ReworkOptions, env sxeval.Environment) sxeval.Expr {
	for i, expr := range me.Front {
		me.Front[i] = expr.Rework(ro, env)
	}
	me.Last = me.Last.Rework(ro, env)
	return me
}

// Macro represents the macro definition form.
type Macro struct {
	Frame  *sxeval.Frame
	PFrame *sxeval.ParseFrame
	Name   string
	Params []*sx.Symbol
	Rest   *sx.Symbol
	Front  []sxeval.Expr // all expressions, but the last
	Last   sxeval.Expr
}

func (m *Macro) IsNil() bool  { return m == nil }
func (m *Macro) IsAtom() bool { return m == nil }
func (m *Macro) IsEql(other sx.Object) bool {
	if m == other {
		return true
	}
	if m.IsNil() {
		return sx.IsNil(other)
	}
	if otherF, ok := other.(*Macro); ok {
		if m.Name != otherF.Name || m.Rest != otherF.Rest || m.Last != otherF.Last {
			return false
		}
		if len(m.Params) != len(otherF.Params) || len(m.Front) != len(otherF.Front) {
			return false
		}
		for i, p := range m.Params {
			if p != otherF.Params[i] {
				return false
			}
		}
		for i, e := range m.Front {
			if e != otherF.Front[i] {
				return false
			}
		}
		return m.PFrame.IsEql(otherF.PFrame)
	}
	return false
}
func (m *Macro) IsEqual(other sx.Object) bool { return m.IsEql(other) }
func (m *Macro) String() string               { return m.Repr() }
func (m *Macro) Repr() string                 { return sx.Repr(m) }
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

func (m *Macro) Expand(pf *sxeval.ParseFrame, args *sx.Pair) (sx.Object, error) {
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

	proc := callable.Procedure{
		PFrame: m.PFrame,
		Name:   m.Name,
		Params: m.Params,
		Rest:   m.Rest,
		Front:  m.Front,
		Last:   m.Last,
	}

	return m.Frame.MakeCalleeFrame().Call(&proc, macroArgs)
}
