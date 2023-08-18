//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package cond_test

import (
	"strings"
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins/cond"
	"zettelstore.de/sx.fossil/sxeval"
	"zettelstore.de/sx.fossil/sxreader"
)

func TestQuote(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		src     string
		exp     string
		withErr bool
	}{
		{name: "if-symbol", src: "if", exp: "#<syntax:if>"},
		{name: "if-0", src: "(if)", exp: "if: requires 2 or 3 arguments, got none", withErr: true},
		{name: "if-1", src: "(if 1)", exp: "if: requires 2 or 3 arguments, got one", withErr: true},
		{name: "if-4", src: "(if 1 2 3 4)", exp: "if: requires 2 or 3 arguments, got more", withErr: true},
		{name: "if-2-one", src: "(if 1 2)", exp: "2"},
		{name: "if-2-true", src: "(if True 2)", exp: "2"},
		{name: "if-2-a", src: "(if \"a\" 2)", exp: "2"},
		{name: "if-2-nil", src: "(if () 2)", exp: "()"},
		{name: "if-2-empty", src: "(if \"\" 2)", exp: "()"},
		{name: "if-2-false", src: "(if False 2)", exp: "()"},
		{name: "if-3-one", src: "(if 1 2 3)", exp: "2"},
		{name: "if-3-true", src: "(if True 2 3)", exp: "2"},
		{name: "if-3-string", src: "(if \"string\" 2 3)", exp: "2"},
		{name: "if-3-nil", src: "(if () 2 3)", exp: "3"},
		{name: "if-3-empty", src: "(if \"\" 2 3)", exp: "3"},
		{name: "if-3-false", src: "(if False 2 3)", exp: "3"},
		{name: "if-3-err-cond", src: "(if (if) 2 3)", exp: "if: requires 2 or 3 arguments, got none", withErr: true},
		{name: "if-3-err-true", src: "(if 1 (if) 3)", exp: "if: requires 2 or 3 arguments, got none", withErr: true},
		{name: "if-3-err-false", src: "(if 1 2 (if))", exp: "if: requires 2 or 3 arguments, got none", withErr: true},
		{name: "if-3-err-sym", src: "(if x 2 3)", exp: "symbol \"x\" not bound in environment \"if-3-err-sym\"", withErr: true},
	}

	root := sxeval.MakeRootEnvironment()
	engine := sxeval.MakeEngine(sx.MakeMappedFactory(), root)
	engine.BindSyntax("if", cond.IfS)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src), sxreader.WithSymbolFactory(engine.SymbolFactory()))

			val, err := rd.Read()
			if err != nil {
				t.Errorf("Error %v while reading %s", err, tc.src)
				return
			}
			env := sxeval.MakeChildEnvironment(root, tc.name, 0)
			res, err := engine.Eval(env, val)
			if err != nil {
				if tc.withErr {
					if got := err.Error(); got != tc.exp {
						t.Errorf("Error %q expected, but got %q", got, tc.exp)
						return
					}
					return
				}
				t.Errorf("unexpected error %v", err)
				return
			}
			if got := res.Repr(); got != tc.exp {
				t.Errorf("%s should result in %q, but got %q", tc.src, tc.exp, got)
			}
		})
	}
}
