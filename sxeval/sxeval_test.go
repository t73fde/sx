//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval_test

import (
	"strings"
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
	"zettelstore.de/sx.fossil/sxreader"
)

func createTestEnv(sf sx.SymbolFactory) sxeval.Environment {
	env := sxeval.MakeRootEnvironment()

	symCat := sf.MustMake("cat")
	env.Bind(symCat, sxeval.BuiltinA(func(args []sx.Object) (sx.Object, error) {
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
	}))

	symHello := sf.MustMake("hello")
	env.Bind(symHello, sx.String("Hello, World"))
	return env
}

type testcase struct {
	name string
	src  string
	exp  string
	// mustErr bool
}

func TestEval(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{name: "nil", src: `()`, exp: "()"},
		{name: "zero", src: "0", exp: "0"},
		{name: "quote-sym", src: "quote", exp: "#<syntax:quote>"},
		{name: "quote-zero", src: "(quote 0)", exp: "0"},
		{name: "quote-nil", src: "(quote ())", exp: "()"},
		{name: "quote-list", src: "(quote (1 2 3))", exp: "(1 2 3)"},
		{name: "hello", src: "hello", exp: `"Hello, World"`},
		{name: "cat-empty", src: `(cat)`, exp: `""`},
		{name: "cat-123", src: "(cat 1 2 3)", exp: `"123"`},
		{name: "cat-hello-sx", src: `(cat hello ": sx")`, exp: `"Hello, World: sx"`},
		// {name: "err-binding", src: "moin", mustErr: true},
		// {name: "err-callable", src: "(hello)", mustErr: true},
	}
	sf := sx.MakeMappedFactory()
	root := createTestEnv(sf)
	root.Bind(sf.MustMake("quote"), sxeval.MakeSyntax("quote", func(_ *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
		return sxeval.ObjExpr{Obj: args.Car()}, nil
	}))
	engine := sxeval.MakeEngine(sf, root)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src), sxreader.WithSymbolFactory(sf))
			val, err := rd.Read()
			if err != nil {
				t.Errorf("Error %v while reading %s", err, tc.src)
				return
			}
			env := sxeval.MakeChildEnvironment(root, tc.name, 0)
			res, err := engine.Eval(env, val)
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
