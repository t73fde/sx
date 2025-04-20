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

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/sx/sxeval"
	"t73f.de/r/sx/sxreader"
)

// Contains tests of all builtins in sub-packages.

type (
	tTestCase struct {
		name    string
		src     string
		exp     string
		withErr bool
	}
	tTestCases []tTestCase
)

func (tcs tTestCases) Run(t *testing.T) {
	t.Helper()
	root := createBinding()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			rd := sxreader.MakeReader(strings.NewReader(tc.src))

			var sb strings.Builder
			bind := root.MakeChildBinding(tc.name, 0)
			env := sxeval.MakeEnvironment()
			for {
				obj, err := rd.Read()
				if err != nil {
					if err == io.EOF {
						break
					}
					if tc.withErr {
						sb.WriteString(fmt.Errorf("{[{%w}]}", err).Error())
						continue
					}
					t.Errorf("Error %v while reading %s", err, tc.src)
					return
				}
				res, err := env.Eval(obj, bind)
				if err != nil {
					if tc.withErr {
						sb.WriteString(fmt.Errorf("{[{%w}]}", err).Error())
						continue
					}
					t.Errorf("unexpected error: %v", fmt.Errorf("%w", err))
					return
				} else if tc.withErr {
					t.Errorf("should fail, but got: %v", res)
					return
				}
				if sb.Len() > 0 {
					sb.WriteByte(' ')
				}
				_, _ = sx.Print(&sb, res)
			}
			if got := sb.String(); got != tc.exp {
				t.Errorf("%s should result in %q, but got %q", tc.src, tc.exp, got)
			}
		})
	}

}

func createBinding() *sxeval.Binding {
	numBuiltins := len(specials) + len(builtins) + len(objects)
	root := sxeval.MakeRootBinding(numBuiltins)
	_ = sxeval.BindSpecials(root, specials...)
	_ = sxeval.BindBuiltins(root, builtins...)
	root.Freeze()
	env := root.MakeChildBinding("vars", len(objects))
	for _, obj := range objects {
		if err := env.Bind(sx.MakeSymbol(obj.name), obj.obj); err != nil {
			panic(err)
		}
	}
	return env
}

var specials = []*sxeval.Special{
	&sxbuiltins.QuoteS, &sxbuiltins.QuasiquoteS, // quote, quasiquote
	&sxbuiltins.UnquoteS, &sxbuiltins.UnquoteSplicingS, // unquote, unquote-splicing
	&sxbuiltins.DefVarS,                     // defvar
	&sxbuiltins.DefunS, &sxbuiltins.LambdaS, // defun, lambda
	&sxbuiltins.DefDynS, &sxbuiltins.DynLambdaS, // defdyn, dyn-lambda
	&sxbuiltins.DefMacroS,                  //  defmacro
	&sxbuiltins.LetS, &sxbuiltins.LetStarS, // let, let*
	&sxbuiltins.SetXS,  // set!
	&sxbuiltins.IfS,    // if
	&sxbuiltins.BeginS, // begin
}
var builtins = []*sxeval.Builtin{
	&sxbuiltins.Equal,                    // =
	&sxbuiltins.Identical,                // ==
	&sxbuiltins.SymbolP,                  // symbol?
	&sxbuiltins.NullP,                    // null?
	&sxbuiltins.Cons,                     // cons
	&sxbuiltins.PairP, &sxbuiltins.ListP, // pair?, list?
	&sxbuiltins.Car, &sxbuiltins.Cdr, // car, cdr
	&sxbuiltins.Caar, &sxbuiltins.Cadr, &sxbuiltins.Cdar, &sxbuiltins.Cddr,
	&sxbuiltins.Caaar, &sxbuiltins.Caadr, &sxbuiltins.Cadar, &sxbuiltins.Caddr,
	&sxbuiltins.Cdaar, &sxbuiltins.Cdadr, &sxbuiltins.Cddar, &sxbuiltins.Cdddr,
	&sxbuiltins.Caaaar, &sxbuiltins.Caaadr, &sxbuiltins.Caadar, &sxbuiltins.Caaddr,
	&sxbuiltins.Cadaar, &sxbuiltins.Cadadr, &sxbuiltins.Caddar, &sxbuiltins.Cadddr,
	&sxbuiltins.Cdaaar, &sxbuiltins.Cdaadr, &sxbuiltins.Cdadar, &sxbuiltins.Cdaddr,
	&sxbuiltins.Cddaar, &sxbuiltins.Cddadr, &sxbuiltins.Cdddar, &sxbuiltins.Cddddr,
	&sxbuiltins.Last,                       // last
	&sxbuiltins.List, &sxbuiltins.ListStar, // list, list*
	&sxbuiltins.Append,               // append
	&sxbuiltins.Reverse,              // reverse
	&sxbuiltins.Assoc,                // assoc
	&sxbuiltins.All, &sxbuiltins.Any, // all, any
	&sxbuiltins.Map,                           // map
	&sxbuiltins.Apply,                         // apply
	&sxbuiltins.Fold, &sxbuiltins.FoldReverse, // fold, fold-reverse
	&sxbuiltins.NumberP,                               // number?
	&sxbuiltins.Add, &sxbuiltins.Sub, &sxbuiltins.Mul, // +, -, *
	&sxbuiltins.Div, &sxbuiltins.Mod, // div, mod
	&sxbuiltins.NumLess, &sxbuiltins.NumLessEqual, // <, <=
	&sxbuiltins.NumGreater, &sxbuiltins.NumGreaterEqual, // >, >=
	&sxbuiltins.ToString, &sxbuiltins.Concat, // ->string, concat
	&sxbuiltins.Vector, &sxbuiltins.VectorP, // vector, vector?
	&sxbuiltins.VectorSetBang,                   // vset!
	&sxbuiltins.List2Vector,                     // list->vector
	&sxbuiltins.Length, &sxbuiltins.LengthEqual, // length, length=
	&sxbuiltins.LengthLess, &sxbuiltins.LengthGreater, // length<, length>
	&sxbuiltins.Nth,               // nth
	&sxbuiltins.Sequence2List,     // seq->list
	&sxbuiltins.CallableP,         // callable?
	&sxbuiltins.Macroexpand0,      // macroexpand-0
	&sxbuiltins.DefinedP,          // defined?
	&sxbuiltins.CurrentBinding,    // current-binding
	&sxbuiltins.ParentBinding,     // parent-binding
	&sxbuiltins.Bindings,          // bindings
	&sxbuiltins.BoundP,            // bound?
	&sxbuiltins.BindingLookup,     // binding-lookup
	&sxbuiltins.BindingResolve,    // binding-resolve
	&sxbuiltins.Pretty,            // pp
	&sxbuiltins.Error,             // error
	&sxbuiltins.NotBoundError,     // not-bound-error
	&sxbuiltins.ParseExpression,   // parse-expression
	&sxbuiltins.UnparseExpression, // unparse-expression
	&sxbuiltins.RunExpression,     // run-expression
	&sxbuiltins.Eval,              // eval
}

var objects = []struct {
	name string
	obj  sx.Object
}{
	{"NIL", sx.Nil()}, {"TRUE", sx.Int64(1)}, {"FALSE", sx.Nil()},
	{"ZERO", sx.Int64(0)}, {"ONE", sx.Int64(1)}, {"TWO", sx.Int64(2)},

	{"b", sx.Int64(11)},
	{"c", sx.MakeList(sx.Int64(22), sx.Int64(33))},
	{"d", sx.MakeList(sx.Int64(44), sx.Int64(55))},
	{"x", sx.Int64(3)}, {"y", sx.Int64(5)},
	{"lang0", sx.String{}}, {"lang1", sx.MakeString("de-DE")},
}
