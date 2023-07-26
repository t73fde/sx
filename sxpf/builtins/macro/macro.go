//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package macro

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins/callable"
	"zettelstore.de/sx.fossil/sxpf/builtins/define"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

// MacroS parses a macro specification.
//
// Syntactically, it is the same as a procedure specification (aka lambda).
func MacroS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	le, err := callable.LambdaS(eng, env, args)
	if err != nil {
		return nil, err
	}
	return makeMacroExpr(le.(*callable.LambdaExpr)), err
}

func makeMacroExpr(le *callable.LambdaExpr) eval.Expr {
	return &MacroExpr{
		Name:   le.Name,
		Params: le.Params,
		Rest:   le.Rest,
		Front:  le.Front,
		Last:   le.Last,
	}
}

// DefMacroS parses a macro specfication and assigns it to a value.
func DefMacroS(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	if args == nil {
		return nil, eval.ErrNoArgs
	}
	sym, isSymbol := sxpf.GetSymbol(args.Car())
	if !isSymbol {
		return nil, fmt.Errorf("not a symbol: %T/%v", args.Car(), args.Car())
	}
	args = args.Tail()
	if args == nil {
		return nil, fmt.Errorf("parameter spec and body missing")
	}
	le, err := callable.ParseProcedure(eng, env, sxpf.Repr(sym), args.Car(), args.Cdr())
	if err != nil {
		return nil, err
	}
	return &define.DefineExpr{Sym: sym, Val: makeMacroExpr(le)}, nil
}

type MacroExpr struct {
	Name   string
	Params []*sxpf.Symbol
	Rest   *sxpf.Symbol
	Front  []eval.Expr // all expressions, but the last
	Last   eval.Expr
}

func (me *MacroExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	return &Macro{
		Env:    env,
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
	l, err = eval.PrintFrontLast(w, me.Front, me.Last)
	length += l
	return length, err
}
func (me *MacroExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	for i, expr := range me.Front {
		me.Front[i] = expr.Rework(ro, env)
	}
	me.Last = me.Last.Rework(ro, env)
	return me
}

// Macro represents the macro definition form.
type Macro struct {
	Env    sxpf.Environment
	Name   string
	Params []*sxpf.Symbol
	Rest   *sxpf.Symbol
	Front  []eval.Expr // all expressions, but the last
	Last   eval.Expr
}

func (m *Macro) IsNil() bool  { return m == nil }
func (m *Macro) IsAtom() bool { return m == nil }
func (m *Macro) IsEql(other sxpf.Object) bool {
	if m == other {
		return true
	}
	if m.IsNil() {
		return sxpf.IsNil(other)
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
		return m.Env.IsEql(otherF.Env)
	}
	return false
}
func (m *Macro) IsEqual(other sxpf.Object) bool { return m.IsEql(other) }
func (m *Macro) String() string                 { return m.Repr() }
func (m *Macro) Repr() string                   { return sxpf.Repr(m) }
func (m *Macro) Print(w io.Writer) (int, error) {
	return sxpf.WriteStrings(w, "#<macro:", m.Name, ">")
}
func (m *Macro) Parse(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
	form, err := m.Expand(eng, env, args)
	if err != nil {
		return nil, err
	}
	return nil, eval.ErrParseAgain{Env: env, Form: form}
}

func (m *Macro) Expand(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (sxpf.Object, error) {
	var macroArgs []sxpf.Object
	arg := sxpf.Object(args)
	for {
		if sxpf.IsNil(arg) {
			break
		}
		pair, isPair := sxpf.GetPair(arg)
		if !isPair {
			return nil, sxpf.ErrImproper{Pair: args}
		}
		macroArgs = append(macroArgs, pair.Car())
		arg = pair.Cdr()
	}

	proc := callable.Procedure{
		Env:    m.Env,
		Name:   m.Name,
		Params: m.Params,
		Rest:   m.Rest,
		Front:  m.Front,
		Last:   m.Last,
	}

	return eng.Call(env, &proc, macroArgs)
}
