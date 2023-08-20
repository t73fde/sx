//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins"
	"zettelstore.de/sx.fossil/sxbuiltins/binding"
	"zettelstore.de/sx.fossil/sxbuiltins/callable"
	"zettelstore.de/sx.fossil/sxbuiltins/cond"
	"zettelstore.de/sx.fossil/sxbuiltins/define"
	"zettelstore.de/sx.fossil/sxbuiltins/env"
	"zettelstore.de/sx.fossil/sxbuiltins/equiv"
	"zettelstore.de/sx.fossil/sxbuiltins/list"
	"zettelstore.de/sx.fossil/sxbuiltins/macro"
	"zettelstore.de/sx.fossil/sxbuiltins/number"
	"zettelstore.de/sx.fossil/sxbuiltins/pprint"
	"zettelstore.de/sx.fossil/sxbuiltins/quote"
	"zettelstore.de/sx.fossil/sxbuiltins/timeit"
	"zettelstore.de/sx.fossil/sxeval"
	"zettelstore.de/sx.fossil/sxreader"
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
	engine := createEngine()
	sf := engine.SymbolFactory()
	root := engine.GetToplevelEnv()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src), sxreader.WithSymbolFactory(sf))
			symQuote := sf.MustMake("quote")
			quote.InstallQuoteReader(rd, symQuote, '\'')
			symQQ, symUQ, symUQS := sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")
			quote.InstallQuasiQuoteReader(rd, symQQ, '`', symUQ, ',', symUQS, '@')

			var sb strings.Builder
			for {
				val, err := rd.Read()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Errorf("Error %v while reading %s", err, tc.src)
					return
				}
				env := sxeval.MakeChildEnvironment(root, tc.name, 0)
				res, err := engine.Eval(env, val)
				if err != nil {
					if tc.withErr {
						sb.WriteString(fmt.Errorf("{[{%w}]}", err).Error())
						continue
					}
					t.Errorf("unexpected error: %v", fmt.Errorf("%w", err))
					continue
				} else if tc.withErr {
					t.Errorf("should fail, but got: %v", res)
					continue
				}
				sx.Print(&sb, res)
			}
			if got := sb.String(); got != tc.exp {
				t.Errorf("%s should result in %q, but got %q", tc.src, tc.exp, got)
			}
		})
	}

}

func createEngine() *sxeval.Engine {
	sf := sx.MakeMappedFactory()
	root := sxeval.MakeRootEnvironment()
	root, err := quote.InstallQuoteSyntax(root, sf.MustMake("quote"))
	if err != nil {
		panic(err)
	}
	root, err = quote.InstallQuasiQuoteSyntax(root, sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing"))
	if err != nil {
		panic(err)
	}

	engine := sxeval.MakeEngine(sf, root)
	for _, syntax := range syntaxes {
		engine.BindSyntax(syntax.name, syntax.fn)
	}
	for _, builtinA := range builtinsA {
		engine.BindBuiltinA(builtinA.name, builtinA.fn)
	}
	for _, builtinFA := range builtinsFA {
		engine.BindBuiltinFA(builtinFA.name, builtinFA.fn)
	}
	env := sxeval.MakeChildEnvironment(root, "vars", len(objects))
	for _, obj := range objects {
		env, err = env.Bind(sf.MustMake(obj.name), obj.obj)
		if err != nil {
			panic(err)
		}
	}
	engine.SetToplevelEnv(env)
	return engine
}

var syntaxes = []struct {
	name string
	fn   sxeval.SyntaxFn
}{
	{"define", define.DefineS}, {"set!", define.SetXS},
	{"if", cond.IfS},
	{"begin", cond.BeginS},
	{"and", sxbuiltins.AndS}, {"or", sxbuiltins.OrS},
	{"lambda", callable.LambdaS},
	{"let", binding.LetS},
	{"timeit", timeit.TimeitS},
	{"defmacro", macro.DefMacroS}, {"macro", macro.MacroS},
}

var builtinsA = []struct {
	name string
	fn   sxeval.BuiltinA
}{
	{"eq?", equiv.EqP}, {"eql?", equiv.EqlP}, {"equal?", equiv.EqualP},
	{"boolean?", sxbuiltins.BooleanP}, {"boolean", sxbuiltins.Boolean}, {"not", sxbuiltins.Not},
	{"number?", number.NumberP},
	{"+", number.Add}, {"-", number.Sub}, {"*", number.Mul},
	{"div", number.Div}, {"mod", number.Mod},
	{"=", number.Equal},
	{"<", number.Less}, {"<=", number.LessEqual},
	{">=", number.GreaterEqual}, {">", number.Greater},
	{"min", number.Min}, {"max", number.Max},
	{"cons", list.Cons}, {"pair?", list.PairP},
	{"null?", list.NullP}, {"list?", list.ListP},
	{"car", list.Car}, {"cdr", list.Cdr}, {"last", list.Last},
	{"list", list.List}, {"list*", list.ListStar}, {"append", list.Append}, {"reverse", list.Reverse},
	{"length", list.Length},
	{"callable?", callable.CallableP},
	{"parent-env", env.ParentEnv}, {"bindings", env.Bindings}, {"all-bindings", env.AllBindings},
}
var builtinsFA = []struct {
	name string
	fn   sxeval.BuiltinFA
}{
	{"map", callable.Map}, {"apply", callable.Apply},
	{"fold", callable.Fold}, {"fold-reverse", callable.FoldReverse},
	{"env", env.Env},
	{"bound?", env.BoundP},
	{"macroexpand-0", macro.MacroExpand0},
	{"pp", pprint.Pretty},
}

var objects = []struct {
	name string
	obj  sx.Object
}{
	{"NIL", sx.Nil()}, {"TRUE", sx.True}, {"FALSE", sx.False},
	{"ZERO", sx.Int64(0)}, {"ONE", sx.Int64(1)}, {"TWO", sx.Int64(2)},
}
