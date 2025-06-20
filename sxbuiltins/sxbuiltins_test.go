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
	_ "embed"
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
			env := sxeval.MakeEnvironment(bind)
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
				res, err := env.Eval(obj, nil)
				if size := env.Size(); size > 0 {
					t.Error("stack not empty, size:", size)
				}
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
	root := sxeval.MakeRootBinding(256)
	_ = sxbuiltins.BindAll(root)
	root.Freeze()
	vars := root.MakeChildBinding("vars", len(objects))
	_ = vars.Bind(sx.MakeSymbol("ROOT"), root)
	for _, obj := range objects {
		if err := vars.Bind(sx.MakeSymbol(obj.name), obj.obj); err != nil {
			panic(err)
		}
	}

	rd := sxreader.MakeReader(strings.NewReader(testprelude))
	env := sxeval.MakeEnvironment(vars)
	for {
		form, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				return vars
			}
			panic(err)
		}
		if _, err = env.Eval(form, nil); err != nil {
			panic(err)
		}
	}
}

//go:embed sxbuiltins_test.sxn
var testprelude string

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

func TestIsPure(t *testing.T) {
	args := make(sx.Vector, 128)
	for i := range cap(args) {
		args[i] = sx.MakeUndefined()
	}
	root := sxeval.MakeRootBinding(256)
	_ = sxbuiltins.BindAll(root)
	for p := root.Bindings(); p != nil; p = p.Tail() {
		val := p.Head().Cdr()
		if b, isBuiltin := val.(*sxeval.Builtin); isBuiltin {
			for i := range cap(args) {
				isPure := b.IsPure(args[0:i])
				if i < int(b.MinArity) {
					if isPure {
						t.Errorf("%v.IsPure(%d) should be false, but is true (min)", b, i)
					}
				} else if b.MaxArity >= b.MinArity && i > int(b.MaxArity) {
					if isPure {
						t.Errorf("%v.IsPure(%d) should be false, but is true (max)", b, i)
					}
				}
			}
		}
	}
}
