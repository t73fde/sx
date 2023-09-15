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
	t.Helper()
	engine := createEngine()
	sf := engine.SymbolFactory()
	root := engine.GetToplevelEnv()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			rd := sxreader.MakeReader(strings.NewReader(tc.src), sxreader.WithSymbolFactory(sf))
			symQuote := sf.MustMake("quote")
			sxbuiltins.InstallQuoteReader(rd, symQuote, '\'')
			symQQ, symUQ, symUQS := sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")
			sxbuiltins.InstallQuasiQuoteReader(rd, symQQ, '`', symUQ, ',', symUQS, '@')

			var sb strings.Builder
			env := sxeval.MakeChildEnvironment(root, tc.name, 0)
			for {
				val, err := rd.Read()
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
				res, err := engine.Eval(env, val)
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
				sx.Print(&sb, res)
			}
			if got := sb.String(); got != tc.exp {
				t.Errorf("%s should result in %q, but got %q", tc.src, tc.exp, got)
			}
		})
	}

}

func createEngine() *sxeval.Engine {
	numBuiltins := len(syntaxes) + len(builtinsA) + len(builtinsFA) + len(objects)
	sf := sx.MakeMappedFactory(numBuiltins + 32)
	root := sxeval.MakeRootEnvironment(numBuiltins)
	if err := sxbuiltins.InstallQuoteSyntax(root, sf.MustMake("quote")); err != nil {
		panic(err)
	}
	if err := sxbuiltins.InstallQuasiQuoteSyntax(root, sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")); err != nil {
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
	root.Freeze()
	env := sxeval.MakeChildEnvironment(root, "vars", len(objects))
	for _, obj := range objects {
		if err := env.Bind(sf.MustMake(obj.name), obj.obj); err != nil {
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
	{"define", sxbuiltins.DefineS}, {"set!", sxbuiltins.SetXS},
	{"if", sxbuiltins.IfS},
	{"begin", sxbuiltins.BeginS},
	{"and", sxbuiltins.AndS}, {"or", sxbuiltins.OrS},
	{"lambda", sxbuiltins.LambdaS},
	{"let", sxbuiltins.LetS}, {"let*", sxbuiltins.LetStarS}, {"letrec", sxbuiltins.LetRecS},
	{"timeit", sxbuiltins.TimeitS},
	{"defmacro", sxbuiltins.DefMacroS},
}

var builtinsA = []struct {
	name string
	fn   sxeval.BuiltinA
}{
	{"eq?", sxbuiltins.EqP}, {"equal?", sxbuiltins.EqualP},
	{"boolean", sxbuiltins.Boolean}, {"not", sxbuiltins.Not},
	{"number?", sxbuiltins.NumberP},
	{"+", sxbuiltins.Add}, {"-", sxbuiltins.Sub}, {"*", sxbuiltins.Mul},
	{"div", sxbuiltins.Div}, {"mod", sxbuiltins.Mod},
	{"=", sxbuiltins.Equal},
	{"<", sxbuiltins.Less}, {"<=", sxbuiltins.LessEqual},
	{">=", sxbuiltins.GreaterEqual}, {">", sxbuiltins.Greater},
	{"min", sxbuiltins.Min}, {"max", sxbuiltins.Max},
	{"cons", sxbuiltins.Cons}, {"pair?", sxbuiltins.PairP},
	{"null?", sxbuiltins.NullP}, {"list?", sxbuiltins.ListP},
	{"car", sxbuiltins.Car}, {"cdr", sxbuiltins.Cdr}, {"last", sxbuiltins.Last},
	{"list", sxbuiltins.List}, {"list*", sxbuiltins.ListStar},
	{"append", sxbuiltins.Append}, {"reverse", sxbuiltins.Reverse},
	{"length", sxbuiltins.Length}, {"assoc", sxbuiltins.Assoc},
	{"->string", sxbuiltins.ToString}, {"string-append", sxbuiltins.StringAppend},
	{"callable?", sxbuiltins.CallableP},
	{"parent-environment", sxbuiltins.ParentEnv},
	{"environment-bindings", sxbuiltins.EnvBindings},
	{"undefined?", sxbuiltins.UndefinedP}, {"defined?", sxbuiltins.DefinedP},
}
var builtinsFA = []struct {
	name string
	fn   sxeval.BuiltinFA
}{
	{"map", sxbuiltins.Map}, {"apply", sxbuiltins.Apply},
	{"fold", sxbuiltins.Fold}, {"fold-reverse", sxbuiltins.FoldReverse},
	{"current-environment", sxbuiltins.CurrentEnv},
	{"bound?", sxbuiltins.BoundP},
	{"environment-lookup", sxbuiltins.EnvLookup}, {"environment-resolve", sxbuiltins.EnvResolve},
	{"macroexpand-0", sxbuiltins.MacroExpand0},
	{"pp", sxbuiltins.Pretty},
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
	{"lang0", sx.String("")}, {"lang1", sx.String("de-DE")},
}
