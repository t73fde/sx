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

// letMacro is the text to define a simple let macro.
const letMacro = "(defmacro let (bindings . body) `((lambda ,(map car bindings) ,@body) ,@(map cadr bindings)))"

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
	if err := sxbuiltins.InstallQuasiQuoteSyntax(root, sf.MustMake("quasiquote"), sf.MustMake("unquote"), sf.MustMake("unquote-splicing")); err != nil {
		panic(err)
	}

	engine := sxeval.MakeEngine(sf, root)
	engine.SetQuote(nil)
	for _, syntax := range syntaxes {
		engine.BindSyntax(syntax.name, syntax.fn)
	}
	for _, builtinA := range builtinsA {
		engine.BindBuiltinAold(builtinA.name, builtinA.fn)
	}
	for _, builtinFA := range builtinsFA {
		engine.BindBuiltinFAold(builtinFA.name, builtinFA.fn)
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
	{"define", sxbuiltins.DefineS},
	{"defvar", sxbuiltins.DefVarS},
	{"defconst", sxbuiltins.DefConstS},
	{"set!", sxbuiltins.SetXS},
	{"if", sxbuiltins.IfS},
	{"defun", sxbuiltins.DefunS}, {"lambda", sxbuiltins.LambdaS},
	{"defmacro", sxbuiltins.DefMacroS},
}

var builtinsA = []struct {
	name string
	fn   sxeval.BuiltinAold
}{
	{"==", sxbuiltins.IdenticalOld}, {"=", sxbuiltins.EqualOld},
	{"number?", sxbuiltins.NumberPold},
	{"+", sxbuiltins.AddOld}, {"-", sxbuiltins.SubOld}, {"*", sxbuiltins.MulOld},
	{"div", sxbuiltins.DivOld}, {"mod", sxbuiltins.ModOld},
	{"<", sxbuiltins.NumLessOld}, {"<=", sxbuiltins.NumLessEqualOld},
	{">=", sxbuiltins.NumGreaterEqualOld}, {">", sxbuiltins.NumGreaterOld},
	{"min", sxbuiltins.MinOld}, {"max", sxbuiltins.MaxOld},
	{"cons", sxbuiltins.ConsOld}, {"pair?", sxbuiltins.PairPold},
	{"null?", sxbuiltins.NullPold}, {"list?", sxbuiltins.ListPold},
	{"car", sxbuiltins.CarOld}, {"cdr", sxbuiltins.CdrOld},
	{"caar", sxbuiltins.CaarOld}, {"cadr", sxbuiltins.CadrOld}, {"cdar", sxbuiltins.CdarOld}, {"cddr", sxbuiltins.CddrOld},
	{"caaar", sxbuiltins.CaaarOld}, {"caadr", sxbuiltins.CaadrOld}, {"cadar", sxbuiltins.CadarOld}, {"caddr", sxbuiltins.CaddrOld},
	{"cdaar", sxbuiltins.CdaarOld}, {"cdadr", sxbuiltins.CdadrOld}, {"cddar", sxbuiltins.CddarOld}, {"cdddr", sxbuiltins.CdddrOld},
	{"caaaar", sxbuiltins.CaaaarOld}, {"caaadr", sxbuiltins.CaaadrOld}, {"caadar", sxbuiltins.CaadarOld}, {"caaddr", sxbuiltins.CaaddrOld},
	{"cadaar", sxbuiltins.CadaarOld}, {"cadadr", sxbuiltins.CadadrOld}, {"caddar", sxbuiltins.CaddarOld}, {"cadddr", sxbuiltins.CadddrOld},
	{"cdaaar", sxbuiltins.CdaaarOld}, {"cdaadr", sxbuiltins.CdaadrOld}, {"cdadar", sxbuiltins.CdadarOld}, {"cdaddr", sxbuiltins.CdaddrOld},
	{"cddaar", sxbuiltins.CddaarOld}, {"cddadr", sxbuiltins.CddadrOld}, {"cdddar", sxbuiltins.CdddarOld}, {"cddddr", sxbuiltins.CddddrOld},
	{"last", sxbuiltins.LastOld},
	{"list", sxbuiltins.ListOld}, {"list*", sxbuiltins.ListStarOld},
	{"append", sxbuiltins.AppendOld}, {"reverse", sxbuiltins.ReverseOld},
	{"length", sxbuiltins.LengthOld}, {"assoc", sxbuiltins.AssocOld},
	{"->string", sxbuiltins.ToStringOld}, {"string-append", sxbuiltins.StringAppendOld},
	{"callable?", sxbuiltins.CallablePold},
	{"parent-environment", sxbuiltins.ParentEnvOld},
	{"environment-bindings", sxbuiltins.EnvBindingsOld},
	{"undefined?", sxbuiltins.UndefinedPold}, {"defined?", sxbuiltins.DefinedPold},
}
var builtinsFA = []struct {
	name string
	fn   sxeval.BuiltinFAold
}{
	{"map", sxbuiltins.MapOld}, {"apply", sxbuiltins.ApplyOld},
	{"fold", sxbuiltins.FoldOld}, {"fold-reverse", sxbuiltins.FoldReverseOld},
	{"current-environment", sxbuiltins.CurrentEnvOld},
	{"bound?", sxbuiltins.BoundPold},
	{"environment-lookup", sxbuiltins.EnvLookupOld}, {"environment-resolve", sxbuiltins.EnvResolveOld},
	{"macroexpand-0", sxbuiltins.MacroExpand0old},
	{"pp", sxbuiltins.PrettyOld},
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
