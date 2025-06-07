//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sxeval_test

import (
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/sx/sxeval"
)

func TestBuiltinSimple(t *testing.T) {
	b := &sxeval.Builtin{
		Name:     "test-builtin",
		MinArity: 0,
		MaxArity: -1,
		TestPure: sxeval.AssertPure,
		Fn0:      func(*sxeval.Environment, *sxeval.Binding) (sx.Object, error) { return sx.Nil(), nil },
		Fn1: func(*sxeval.Environment, sx.Object, *sxeval.Binding) (sx.Object, error) {
			return sx.MakeList(), nil
		},
		Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Binding) (sx.Object, error) {
			return sx.MakeList(args[1:]...), nil
		},
	}

	if sx.IsNil(b) {
		t.Error("Builtin is wrongly treated as Nil()", b)
		return
	}
	expString := "#<builtin:"
	if got := b.String(); !strings.HasPrefix(got, expString) {
		t.Errorf("Builtin.String() should start with %q, but got %q", expString, got)
	}
	expLen := len(expString)
	var sb strings.Builder
	if got, err := sx.Print(&sb, b); err != nil || got <= expLen {
		if err != nil {
			t.Errorf("Builtin %v.Print() resulted in error %v", b, err)
		} else if got != expLen {
			t.Errorf("Builtin %v.Print() should deliver %d bytes, but got %d", b, expLen, got)
		}
	}

	root := sxeval.MakeRootBinding(8)
	_ = sxeval.BindBuiltins(root, b, &sxbuiltins.Apply, &sxbuiltins.List)
	env := sxeval.MakeEnvironment()

	args := sx.Vector{}
	for i := range 10 {
		form := sx.MakeList(
			sx.MakeSymbol(sxbuiltins.Apply.Name),
			sx.MakeSymbol(b.Name),
			sx.MakeList(args...).Cons(sx.MakeSymbol(sxbuiltins.List.Name)))
		res0, err := env.Eval(form, root)
		if err != nil {
			t.Error(err)
			break
		}

		res, err := b.ExecuteCall(nil, args, nil)
		if err != nil {
			t.Error(err)
			break
		}

		if !sx.IsList(res) {
			t.Errorf("%d: result should be a list, but is not: %v", i, res)
		}
		if i > 0 {
			exp := len(args) - 1
			if got := res.(*sx.Pair).Length(); got != exp {
				t.Errorf("Result list %v/%d must be one element shorter than arg %v/%d", res, got, args, exp)
			}
		}

		if !res0.IsEqual(res) {
			t.Error("execution and eval differ. execution:", res, ", eval:", res0)
		}

		args = append(args, sx.Nil())
	}
}
