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
	"io"
	"strings"
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxbuiltins/callable"
	"zettelstore.de/sx.fossil/sxbuiltins/cond"
	"zettelstore.de/sx.fossil/sxbuiltins/define"
	"zettelstore.de/sx.fossil/sxbuiltins/number"
	"zettelstore.de/sx.fossil/sxbuiltins/quote"
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
type testCases []testcase

func (testcases testCases) Run(t *testing.T, engine *sxeval.Engine) {
	sf := engine.SymbolFactory()
	root := engine.GetToplevelEnv()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src), sxreader.WithSymbolFactory(sf))
			symQuote := sf.MustMake("quote")
			quote.InstallQuoteReader(rd, symQuote, '\'')
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

func TestEval(t *testing.T) {
	t.Parallel()

	testcases := testCases{
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
	testcases.Run(t, engine)
}

var sxEvenOdd = `;;; Indirekt recursive definition of even/odd
(define (even? n) (if (= n 0) True (odd? (- n 1))))
(define (odd? n) (if (= n 0) False (even? (- n 1))))
`

func createEngineForTCO() *sxeval.Engine {
	sf := sx.MakeMappedFactory()
	root := sxeval.MakeRootEnvironment()
	engine := sxeval.MakeEngine(sf, root)
	engine.BindSyntax("define", define.DefineS)
	engine.BindSyntax("if", cond.IfS)
	engine.BindBuiltinA("=", number.Equal)
	engine.BindBuiltinA("-", number.Sub)
	engine.BindBuiltinFA("map", callable.Map)
	rd := sxreader.MakeReader(strings.NewReader(sxEvenOdd), sxreader.WithSymbolFactory(sf))
	symQuote := sf.MustMake("quote")
	root, errQ := quote.InstallQuoteSyntax(root, symQuote)
	if errQ != nil {
		panic(errQ)
	}
	env := sxeval.MakeChildEnvironment(root, "TCO", 0)
	for {
		obj, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		_, err = engine.Eval(env, obj)
		if err != nil {
			panic(err)
		}
	}
	engine.SetToplevelEnv(env)
	return engine
}
func TestTailCallOptimization(t *testing.T) {
	t.Parallel()
	testcases := testCases{
		{name: "trivial-even", src: "(even? 0)", exp: "True"},
		{name: "trivial-odd", src: "(odd? 0)", exp: "False"},
		{name: "trivial-map-even", src: "(map even? '(0 1 2 3 4 5 6))", exp: "(True False True False True False True)"},
		{name: "trivial-map-odd", src: "(map odd? '(0 1 2 3 4 5 6))", exp: "(False True False True False True False)"},
		{name: "heavy-even", src: "(even? 1000000)", exp: "True"},
	}
	engine := createEngineForTCO()
	testcases.Run(t, engine)
}
func BenchmarkTCO10(b *testing.B)    { benchmarkTCO(b, 10) }
func BenchmarkTCO20(b *testing.B)    { benchmarkTCO(b, 20) }
func BenchmarkTCO40(b *testing.B)    { benchmarkTCO(b, 40) }
func BenchmarkTCO80(b *testing.B)    { benchmarkTCO(b, 80) }
func BenchmarkTCO160(b *testing.B)   { benchmarkTCO(b, 160) }
func BenchmarkTCO320(b *testing.B)   { benchmarkTCO(b, 320) }
func BenchmarkTCO640(b *testing.B)   { benchmarkTCO(b, 640) }
func BenchmarkTCO1280(b *testing.B)  { benchmarkTCO(b, 1680) }
func BenchmarkTCO2560(b *testing.B)  { benchmarkTCO(b, 2560) }
func BenchmarkTCO5120(b *testing.B)  { benchmarkTCO(b, 5120) }
func BenchmarkTCO10240(b *testing.B) { benchmarkTCO(b, 10240) }
func BenchmarkTCO20480(b *testing.B) { benchmarkTCO(b, 20480) }
func BenchmarkTCO40960(b *testing.B) { benchmarkTCO(b, 40960) }
func BenchmarkTCO81920(b *testing.B) { benchmarkTCO(b, 81920) }
func benchmarkTCO(b *testing.B, i int) {
	engine := createEngineForTCO()
	obj := sx.MakeList(
		engine.SymbolFactory().MustMake("even?"),
		sx.Int64(i),
	)
	root := engine.GetToplevelEnv()
	expr, err := engine.Parse(root, obj)
	if err != nil {
		panic(err)
	}
	expr = engine.Rework(root, expr)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		engine.Execute(root, expr)
	}
}
