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

package sxeval_test

import (
	"io"
	"strings"
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxeval"
	"zettelstore.de/sx.fossil/sxreader"
)

func createTestBinding() *sxeval.Binding {
	bind := sxeval.MakeRootBinding(2)

	symCat := sx.Symbol("cat")
	bind.Bind(symCat, &sxeval.Builtin{
		Name:     "cat",
		MinArity: 0,
		MaxArity: -1,
		TestPure: sxeval.AssertPure,
		Fn: func(_ *sxeval.Environment, args []sx.Object) (sx.Object, error) {
			var sb strings.Builder
			for _, val := range args {
				var s string
				if sv, ok := val.(sx.String); ok {
					s = string(sv)
				} else {
					s = val.String()
				}

				_, err := sb.WriteString(s)
				if err != nil {
					return nil, err
				}
			}
			return sx.String(sb.String()), nil
		},
	})

	symHello := sx.Symbol("hello")
	bind.Bind(symHello, sx.String("Hello, World"))
	return bind
}

type testcase struct {
	name string
	src  string
	exp  string
	// mustErr bool
}
type testCases []testcase

func (testcases testCases) Run(t *testing.T, engine *sxeval.Engine) {
	root := engine.GetToplevelBinding()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src))
			obj, err := rd.Read()
			if err != nil {
				t.Errorf("Error %v while reading %s", err, tc.src)
				return
			}
			bind := sxeval.MakeChildBinding(root, tc.name, 0)
			env := sxeval.MakeExecutionEnvironment(engine, nil, bind)
			res, err := env.Eval(obj)
			if err != nil {
				t.Error(err) // TODO: temp
				return
			}
			if got := res.Repr(); got != tc.exp {
				t.Errorf("%s should result in %q, but got %q", tc.src, tc.exp, got)
			}
		})
	}
}

func TestEval(t *testing.T) {
	t.Parallel()

	testcases := testCases{
		{name: "nil", src: `()`, exp: "()"},
		{name: "zero", src: "0", exp: "0"},
		{name: "hello", src: "hello", exp: `"Hello, World"`},
		{name: "cat-empty", src: `(cat)`, exp: `""`},
		{name: "cat-123", src: "(cat 1 2 3)", exp: `"123"`},
		{name: "cat-hello-sx", src: `(cat hello ": sx")`, exp: `"Hello, World: sx"`},
		// {name: "err-binding", src: "moin", mustErr: true},
		// {name: "err-callable", src: "(hello)", mustErr: true},
	}
	root := createTestBinding()
	engine := sxeval.MakeEngine(root)
	engine.BindSpecial(&sxeval.Special{
		Name: "quote",
		Fn: func(_ *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
			return sxeval.ObjExpr{Obj: args.Car()}, nil
		},
	})
	testcases.Run(t, engine)
}

var sxPrelude = `;; Indirekt recursive definition of even/odd
(defun even? (n) (if (= n 0) 1 (odd? (- n 1))))
(defun odd? (n) (if (= n 0) () (even? (- n 1))))

;; Naive implementation of fac
(defun fac (n) (if (= n 0) 1 (* n (fac (- n 1)))))
`

func createEngineForTCO() *sxeval.Engine {
	root := sxeval.MakeRootBinding(6)
	engine := sxeval.MakeEngine(root)
	engine.BindSpecial(&sxbuiltins.DefunS)
	engine.BindSpecial(&sxbuiltins.IfS)
	engine.BindBuiltin(&sxbuiltins.Equal)
	engine.BindBuiltin(&sxbuiltins.Sub)
	engine.BindBuiltin(&sxbuiltins.Mul)
	engine.BindBuiltin(&sxbuiltins.Map)
	engine.BindBuiltin(&sxbuiltins.List)
	root.Freeze()
	rd := sxreader.MakeReader(strings.NewReader(sxPrelude))
	bind := sxeval.MakeChildBinding(root, "TCO", 128)
	env := sxeval.MakeExecutionEnvironment(engine, nil, bind)
	for {
		obj, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		_, err = env.Eval(obj)
		if err != nil {
			panic(err)
		}
	}
	engine.SetToplevelBinding(bind)
	return engine
}

func TestTailCallOptimization(t *testing.T) {
	t.Parallel()
	testcases := testCases{
		{name: "trivial-even", src: "(even? 0)", exp: "1"},
		{name: "trivial-odd", src: "(odd? 0)", exp: "()"},
		{name: "trivial-map-even", src: "(map even? (list 0 1 2 3 4 5 6))", exp: "(1 () 1 () 1 () 1)"},
		{name: "trivial-map-odd", src: "(map odd? (list 0 1 2 3 4 5 6))", exp: "(() 1 () 1 () 1 ())"},
		{name: "heavy-even", src: "(even? 1000000)", exp: "1"},

		// The following is not a TCO test, but a test for a correct fac implementation.
		{name: "fac20", src: "(fac 20)", exp: "2432902008176640000"},
	}
	engine := createEngineForTCO()
	testcases.Run(t, engine)
}
