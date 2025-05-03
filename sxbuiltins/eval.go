//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// ParseExpression parses the given object into an expression (to be
// executed / run later).
var ParseExpression = sxeval.Builtin{
	Name:     "parse-expression",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		expr, err := env.Parse(env.Top(), bind)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(sxeval.MakeExprObj(expr))
		return nil
	},
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		bind, err := GetBinding(env.Pop(), 1)
		if err != nil {
			env.Kill(1)
			return err
		}
		expr, err := env.Parse(env.Top(), bind)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(sxeval.MakeExprObj(expr))
		return nil
	},
}

// UnparseExpression produces a form object of a given expression.
var UnparseExpression = sxeval.Builtin{
	Name:     "unparse-expression",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		expr, err := GetExprObj(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(expr.GetExpr().Unparse())
		return nil
	},
}

// ExecuteExpression executes the given compiled expression, optionally within
// an environment.
var ExecuteExpression = sxeval.Builtin{
	Name:     "execute-expression",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		expr, err := GetExprObj(env.Pop(), 0)
		if err != nil {
			return err
		}
		return env.Execute(expr.GetExpr(), bind)
	},
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		arg1 := env.Pop()
		expr, err := GetExprObj(env.Pop(), 0)
		if err != nil {
			return err
		}
		bind, err := GetBinding(arg1, 1)
		if err != nil {
			return err
		}
		return env.Execute(expr.GetExpr(), bind)
	},
}

// Eval evaluates the given object, in an optional environment.
var Eval = sxeval.Builtin{
	Name:     "eval",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
		expr, err := getEvalExpr(env, env.Pop(), bind)
		if err != nil {
			return err
		}
		return env.Execute(expr, bind)
	},
	Fn: func(env *sxeval.Environment, _ int, _ *sxeval.Binding) error {
		bind, err := GetBinding(env.Pop(), 1)
		if err != nil {
			env.Kill(1)
			return err
		}
		expr, err := getEvalExpr(env, env.Pop(), bind)
		if err != nil {
			return err
		}
		return env.Execute(expr, bind)
	},
}

func getEvalExpr(env *sxeval.Environment, arg sx.Object, bind *sxeval.Binding) (sxeval.Expr, error) {
	if exprObj, isExpr := sxeval.GetExprObj(arg); isExpr {
		return exprObj.GetExpr(), nil
	}
	return env.Parse(arg, bind)
}

// Compile an expression oject.
var Compile = sxeval.Builtin{
	Name:     "compile",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		exprObj, err := GetExprObj(env.Top(), 0)
		if err != nil {
			env.Kill(1)
			return err
		}
		cexpr, err := env.Compile(exprObj.GetExpr())
		if err != nil {
			env.Kill(1)
			return err
		}
		env.Set(sxeval.MakeExprObj(cexpr))
		return nil
	},
}

// Disassemble returns a sequence of compiled code.
var Disassemble = sxeval.Builtin{
	Name:     "disassemble",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, _ *sxeval.Binding) error {
		if objAsm, isObjAsm := env.Top().(sxeval.Disassembler); isObjAsm {
			if stmts, found := objAsm.GetAsmCode(); found {
				var lb sx.ListBuilder
				for s := range stmts {
					lb.Add(sx.MakeString(s))
				}
				env.Set(lb.List())
				return nil
			}
		}
		env.Set(sx.Nil())
		return nil
	},
}
