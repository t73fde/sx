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
	Fn1: func(env *sxeval.Environment, arg sx.Object, bind *sxeval.Binding) (sx.Object, error) {
		expr, err := env.Parse(arg, bind)
		if err != nil {
			return nil, err
		}
		return sxeval.MakeExprObj(expr), nil
	},
	Fn: func(env *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		bind, err := GetBinding(args[1], 1)
		if err != nil {
			return nil, err
		}
		expr, err := env.Parse(args[0], bind)
		if err != nil {
			return nil, err
		}
		return sxeval.MakeExprObj(expr), nil
	},
}

// UnparseExpression produces a form object of a given expression.
var UnparseExpression = sxeval.Builtin{
	Name:     "unparse-expression",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		expr, err := GetExprObj(arg, 0)
		if err != nil {
			return nil, err
		}
		return expr.GetExpr().Unparse(), nil
	},
}

// RunExpression executes the given compiled expression, optionally within
// an environment.
var RunExpression = sxeval.Builtin{
	Name:     "run-expression",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, arg sx.Object, bind *sxeval.Binding) (sx.Object, error) {
		expr, err := GetExprObj(arg, 0)
		if err != nil {
			return nil, err
		}
		obj, err := env.Run(expr.GetExpr(), bind)
		return obj, err

	},
	Fn: func(env *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		expr, err := GetExprObj(args[0], 0)
		if err != nil {
			return nil, err
		}
		bind, err := GetBinding(args[1], 1)
		if err != nil {
			return nil, err
		}
		obj, err := env.Run(expr.GetExpr(), bind)
		return obj, err
	},
}

// Eval evaluates the given object, in an optional environment.
var Eval = sxeval.Builtin{
	Name:     "eval",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(env *sxeval.Environment, arg sx.Object, bind *sxeval.Binding) (sx.Object, error) {
		expr, err := getEvalExpr(env, arg, bind)
		if err != nil {
			return nil, err
		}
		return env.Run(expr, bind)
	},
	Fn: func(env *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
		bind, err := GetBinding(args[1], 1)
		if err != nil {
			return nil, err
		}
		expr, err := getEvalExpr(env, args[0], bind)
		if err != nil {
			return nil, err
		}
		return env.Run(expr, bind)
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
	Fn1: func(env *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		exprObj, err := GetExprObj(arg, 0)
		if err != nil {
			return nil, err
		}
		cexpr, err := env.Compile(exprObj.GetExpr())
		if err != nil {
			return nil, err
		}
		return sxeval.MakeExprObj(cexpr), nil
	},
}

// Disassemble returns a sequence of compiled code.
var Disassemble = sxeval.Builtin{
	Name:     "disassemble",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
		if objAsm, isObjAsm := arg.(sxeval.Disassembler); isObjAsm {
			if stmts, found := objAsm.GetAsmCode(); found {
				var lb sx.ListBuilder
				for s := range stmts {
					lb.Add(sx.MakeString(s))
				}
				return lb.List(), nil
			}
		}
		return sx.Nil(), nil
	},
}
