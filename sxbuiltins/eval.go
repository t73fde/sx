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
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

var ParseExpression = sxeval.Builtin{
	Name:     "parse-expression",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		realEnv, err := adaptEnvironment(env, args, 1)
		if err != nil {
			return nil, err
		}
		expr, err := realEnv.Parse(args[0])
		if err != nil {
			return nil, err
		}
		return sxeval.MakeExprObj(expr), nil
	},
}

var ReworkExpression = sxeval.Builtin{
	Name:     "rework-expression",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		expr, err := GetExprObj(args, 0)
		if err != nil {
			return nil, err
		}
		realEnv, err := adaptEnvironment(env, args, 1)
		if err != nil {
			return nil, err
		}
		reworkedExpr := realEnv.Rework(expr.GetExpr())
		return sxeval.MakeExprObj(reworkedExpr), nil
	},
}

var RunExpression = sxeval.Builtin{
	Name:     "run-expression",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		expr, err := GetExprObj(args, 0)
		if err != nil {
			return nil, err
		}
		realEnv, err := adaptEnvironment(env, args, 1)
		if err != nil {
			return nil, err
		}
		obj, err := realEnv.Run(expr.GetExpr())
		return obj, err
	},
}

var Compile = sxeval.Builtin{
	Name:     "compile",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		realEnv, err := adaptEnvironment(env, args, 1)
		if err != nil {
			return nil, err
		}
		expr, err := realEnv.Compile(args[0])
		if err != nil {
			return nil, err
		}
		return sxeval.MakeExprObj(expr), nil
	},
}

var Eval = sxeval.Builtin{
	Name:     "eval",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn: func(env *sxeval.Environment, args sx.Vector) (sx.Object, error) {
		realEnv, err := adaptEnvironment(env, args, 1)
		if err != nil {
			return nil, err
		}

		var expr sxeval.Expr
		exprObj, isExpr := sxeval.GetExprObj(args[0])
		if isExpr {
			expr = exprObj.GetExpr()
		} else {
			expr, err = realEnv.Compile(args[0])
			if err != nil {
				return nil, err
			}
		}

		obj, err := realEnv.Run(expr)
		return obj, err
	},
}

func adaptEnvironment(env *sxeval.Environment, args sx.Vector, pos int) (*sxeval.Environment, error) {
	if pos < len(args) {
		if sx.IsNil(args[pos]) {
			return env, nil
		}
		bind, err := GetBinding(args, pos)
		if err != nil {
			return nil, err
		}
		return env.RebindExecutionEnvironment(bind), nil
	}
	return env, nil
}
