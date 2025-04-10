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

package sxbuiltins_test

import "testing"

func TestLambda(t *testing.T) {
	t.Parallel()
	tcsLambda.Run(t)
}

var tcsLambda = tTestCases{
	{
		name:    "err-callable?-0",
		src:     "(callable?)",
		exp:     "{[{callable?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "callable?-1", src: "(callable? 1)", exp: "()"},
	{name: "callable?-callable?", src: "(callable? callable?)", exp: "T"},
	{name: "callable?-map", src: "(callable? map)", exp: "T"},
	{name: "callable-lambda", src: "(callable? (lambda x 0))", exp: "T"},

	{name: "err-defun-0", src: "(defun)", exp: "{[{defun: no arguments given}]}", withErr: true},
	{name: "err-defun-nosym", src: "(defun 1)", exp: "{[{defun: not a symbol: sx.Int64/1}]}", withErr: true},
	{name: "err-defun-a", src: "(defun a)", exp: "{[{defun: parameter spec and body missing}]}", withErr: true},
	{name: "defun", src: "(defun a () 1)", exp: "#<lambda:a>", nocompare: true},
	{name: "defun-eval", src: "((defun a () 1))", exp: "1"},
	{name: "defun-eval-arg", src: "((defun a (a) a) 1)", exp: "1"},

	{
		name:    "err-lambda-0",
		src:     "(lambda)",
		exp:     "{[{lambda: parameter spec and body missing}]}",
		withErr: true,
	},
	{
		name:    "err-lambda-x",
		src:     "(lambda x)",
		exp:     "{[{lambda: missing body}]}",
		withErr: true},
	{
		name:    "err-lambda-1",
		src:     "(lambda 1)",
		exp:     "{[{lambda: only symbol and list allowed in parameter spec, but got: sx.Int64/1}]}",
		withErr: true,
	},
	{name: "lambda-1", src: "((lambda x 1))", exp: "1"},
	{name: "lambda-x-x-rest", src: "((lambda x x) 1)", exp: "(1)"},
	{name: "lambda-x-x", src: "((lambda (x) x) 1)", exp: "1"},
	{name: "lambda-adder", src: "(((lambda (n) (lambda (x) (+ n x))) 3) 4)", exp: "7"},
	{name: "lambda-add", src: "((lambda (x y) (+ x y)) 3 4)", exp: "7"},
	{name: "lambda-add-apply", src: "((lambda x (apply + x)) 3 4 5)", exp: "12"},
	{
		name:    "err-lambda-add-many",
		src:     "((lambda (x y) (+ x y)) 3 4 5)",
		exp:     "{[{(x y): excess arguments: [5]}]}",
		withErr: true,
	},
	{
		name:    "err-lambda-add-less",
		src:     "((lambda (x y) (+ x y)) 3)",
		exp:     "{[{(x y): missing arguments: [y]}]}",
		withErr: true,
	},
	{name: "lambda-front", src: "((lambda x x 3) 1)", exp: "3"},
	{name: "lambda-car-1", src: "((lambda (x . y) x) 1)", exp: "1"},
	{name: "lambda-car-2", src: "((lambda (x . y) x) 1 2)", exp: "1"},
	{name: "lambda-cdr-1", src: "((lambda (x . y) y) 1)", exp: "()"},
	{name: "lambda-cdr-2", src: "((lambda (x . y) y) 1 2)", exp: "(2)"},
	{
		name:    "err-lambda-arg-nosym",
		src:     "(lambda (1) 1)",
		exp:     "{[{lambda: symbol in list expected, but got sx.Int64/1}]}",
		withErr: true,
	},

	{name: "lambda-name", src: "(lambda \"adbmal\" x x)", exp: "#<lambda:adbmal>", nocompare: true},
	{
		name:    "err-lambda-name-body",
		src:     "(lambda \"adbmal\" x)",
		exp:     "{[{lambda: missing body}]}",
		withErr: true,
	},
	{
		name:    "err-lambda-name-param",
		src:     "(lambda \"adbmal\")",
		exp:     "{[{lambda: parameter spec and body missing}]}",
		withErr: true,
	},

	{
		name:      "lambda-lex-resolve",
		src:       "(defvar y 3) (defun fn (x) (+ x y)) (let ((y 17)) (fn 4))",
		exp:       "3 #<lambda:fn> 7",
		nocompare: true,
	},
	{
		name:      "lambda-dyn-resolve",
		src:       "(defvar y 3) (defdyn fn (x) (+ x y)) (let ((y 17)) (fn 4))",
		exp:       "3 #<dyn-lambda:fn> 21",
		nocompare: true,
	},
}
