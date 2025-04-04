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
		Fn0: func(_ *sxeval.Environment) (sx.Object, error) {
			return sx.String{}, nil
		},
		Fn1: func(_ *sxeval.Environment, arg sx.Object) (sx.Object, error) {
			return sx.MakeString(arg.GoString()), nil
		},
		Fn2: func(_ *sxeval.Environment, arg0, arg1 sx.Object) (sx.Object, error) {
			return sx.MakeString(arg0.GoString() + arg1.GoString()), nil
		},
		Fn: func(_ *sxeval.Environment, args sx.Vector) (sx.Object, error) {
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
			env := sxeval.MakeExecutionEnvironment(bind)
			res, err := env.Eval(obj)
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
	_ = root.BindSpecial(&sxeval.Special{
		Name: "quote",
		Fn: func(_ *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
			return sxeval.ObjExpr{Obj: args.Car()}, nil
		},
	})
	testcases.Run(t, root)
}

var sxPrelude = `;; Some helpers
(defvar NIL ())
(defvar T 'T)
(defmacro not (x) (list 'if x NIL T))

;; Indirekt recursive definition of even/odd
(defvar odd? ()) ; define symbol odd? to make lookup in even? faster, because it is known.
(defun even? (n) (if (= n 0) 1 (odd? (- n 1))))
(defun odd? (n) (if (= n 0) () (even? (- n 1))))

;; Naive implementation of fac
(defun fac (n) (if (= n 0) 1 (* n (fac (- n 1)))))

;; Accumulator based implementation of fac
(defun faa (n acc) (if (= n 0) acc (faa (- n 1) (* acc n))))

;; Naive fibonacci
(defun fib (n) (if (<= n 1) 1 (+ (fib (- n 1)) (fib (- n 2)))))

;; Takeuchi benchmark
(defun tak (x y z)
  (if (not (< y x))
      z
      (tak (tak (- x 1) y z)
           (tak (- y 1) z x)
           (tak (- z 1) x y))))
`

func createBindingForTCO() *sxeval.Binding {
	root := sxeval.MakeRootBinding(32)
	_ = root.BindSpecial(&sxbuiltins.QuoteS)
	_ = root.BindSpecial(&sxbuiltins.DefVarS)
	_ = root.BindSpecial(&sxbuiltins.DefMacroS)
	_ = root.BindSpecial(&sxbuiltins.DefunS)
	_ = root.BindSpecial(&sxbuiltins.IfS)
	_ = root.BindBuiltin(&sxbuiltins.Equal)
	_ = root.BindBuiltin(&sxbuiltins.NumLess)
	_ = root.BindBuiltin(&sxbuiltins.NumLessEqual)
	_ = root.BindBuiltin(&sxbuiltins.Add)
	_ = root.BindBuiltin(&sxbuiltins.Sub)
	_ = root.BindBuiltin(&sxbuiltins.Mul)
	_ = root.BindBuiltin(&sxbuiltins.Map)
	_ = root.BindBuiltin(&sxbuiltins.List)
	root.Freeze()
	rd := sxreader.MakeReader(strings.NewReader(sxPrelude))
	bind := root.MakeChildBinding("TCO", 128)
	env := sxeval.MakeExecutionEnvironment(bind)
	for {
		obj, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if _, err = env.Eval(obj); err != nil {
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

		// The following is not a TCO test, but a test for a correct fac implementation.
		{name: "fac20", src: "(fac 10)", exp: "3628800"},

		// The following is not a TCO test, but a test for a correct faa implementation.
		{name: "faa20", src: "(faa 10 1)", exp: "3628800"},

		// The following is not a TCO test, but a test for a correct fac implementation.
		{name: "fib20", src: "(fib 6)", exp: "13"},
	}
	root := createBindingForTCO()
	testcases.Run(t, root)
}
