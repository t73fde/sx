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
	_ "embed"
	"io"
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/sx/sxeval"
	"t73f.de/r/sx/sxreader"
)

func createTestBinding() *sxeval.Binding {
	bind := sxeval.MakeRootBinding(2)

	symCat := sx.MakeSymbol("cat")
	_ = bind.Bind(symCat, &sxeval.Builtin{
		Name:     "cat",
		MinArity: 0,
		MaxArity: -1,
		TestPure: sxeval.AssertPure,
		Fn0: func(_ *sxeval.Environment, _ *sxeval.Binding) (sx.Object, error) {
			return sx.String{}, nil
		},
		Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Binding) (sx.Object, error) {
			return sx.MakeString(arg.GoString()), nil
		},
		Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
			var sb strings.Builder
			for _, val := range args {
				_, err := sb.WriteString(val.GoString())
				if err != nil {
					return nil, err
				}
			}
			return sx.MakeString(sb.String()), nil
		},
	})

	symHello := sx.MakeSymbol("hello")
	_ = bind.Bind(symHello, sx.MakeString("Hello, World"))
	return bind
}

type testcase struct {
	name string
	src  string
	exp  string
	// mustErr bool
}
type testCases []testcase

func (testcases testCases) Run(t *testing.T, root *sxeval.Binding) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src))
			obj, err := rd.Read()
			if err != nil {
				t.Errorf("Error %v while reading %s", err, tc.src)
				return
			}
			bind := root.MakeChildBinding(tc.name, 0)
			env := sxeval.MakeEnvironment()
			res, err := env.Eval(obj, bind)
			if err != nil {
				t.Error(err) // TODO: temp
				return
			}
			if got := res.String(); got != tc.exp {
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
	_ = sxeval.BindSpecials(root, &sxeval.Special{
		Name: "quote",
		Fn: func(_ *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
			return sxeval.ObjExpr{Obj: args.Car()}, nil
		},
	})
	testcases.Run(t, root)
}

//go:embed tests.sxn
var sxevalTests string

func createBindingForTCO() *sxeval.Binding {
	root := sxeval.MakeRootBinding(32)
	if err := sxbuiltins.LoadPrelude(root); err != nil {
		panic(err)
	}
	if err := sxeval.BindSpecials(root,
		&sxbuiltins.QuoteS, &sxbuiltins.DefVarS, &sxbuiltins.DefunS,
		&sxbuiltins.LambdaS); err != nil {
		panic(err)
	}
	if err := sxeval.BindBuiltins(root,
		&sxbuiltins.Equal, &sxbuiltins.NumLess, &sxbuiltins.NumLessEqual,
		&sxbuiltins.Add, &sxbuiltins.Sub, &sxbuiltins.Mul,
		&sxbuiltins.Map, &sxbuiltins.List,
		&sxbuiltins.NumberP, &sxbuiltins.SymbolP, &sxbuiltins.PairP,
		&sxbuiltins.Cadr, &sxbuiltins.Caddr,
		&sxbuiltins.Error,
	); err != nil {
		panic(err)
	}
	root.Freeze()
	rd := sxreader.MakeReader(strings.NewReader(sxevalTests))
	bind := root.MakeChildBinding("TCO", 128)
	env := sxeval.MakeEnvironment()
	for {
		obj, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		expr, err := parseAndMore(env, obj, bind)
		if err != nil {
			panic(err)
		}
		if _, err = env.Run(expr, bind); err != nil {
			panic(err)
		}
	}
	return bind
}

func TestTailCallOptimization(t *testing.T) {
	t.Parallel()
	testcases := testCases{
		{name: "trivial-even", src: "(even? 0)", exp: "1"},
		{name: "trivial-odd", src: "(odd? 0)", exp: "()"},
		{name: "trivial-map-even", src: "(map even? (list 0 1 2 3 4 5 6))", exp: "(1 () 1 () 1 () 1)"},
		{name: "trivial-map-odd", src: "(map odd? (list 0 1 2 3 4 5 6))", exp: "(() 1 () 1 () 1 ())"},
		{name: "heavy-even", src: "(even? 1000000)", exp: "1"},

		// The following are not TCO tests, but tests for correct implementations.
		{name: "fac10", src: "(fac 10)", exp: "3628800"},
		{name: "faa10", src: "(faa 10 1)", exp: "3628800"},
		{name: "fib20", src: "(fib 6)", exp: "13"},
		{name: "tak-10-5-3", src: "(tak 10 5 2)", exp: "5"},
		{name: "deriv-x", src: "(deriv 'x 'x)", exp: "1"},
		{name: "deriv-c", src: "(deriv 'c 'x)", exp: "0"},
		{name: "deriv-+cx", src: "(deriv '(+ c x) 'x)", exp: "(+ 0 1)"},
		{name: "deriv-*cx", src: "(deriv '(* c x) 'x)", exp: "(+ (* 0 x) (* c 1))"},
		//{name: "deriv-x2", src: "(deriv '(expr x 3) 'x)", exp: "1"},
		{name: "deriv-x2", src: "(deriv '(expt x 3) 'x)", exp: "(* 3 (* (expt x 2) 1))"},
		//{name: "test-deriv", src: "(test-deriv deriv-test-cases)", exp: ""},
	}
	root := createBindingForTCO()
	testcases.Run(t, root)
}
