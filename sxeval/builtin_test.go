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
	"t73f.de/r/sx/sxeval"
)

func TestBuiltinSimple(t *testing.T) {
	b := &sxeval.Builtin{
		Name:     "test-builtin",
		MinArity: 0,
		MaxArity: -1,
		TestPure: sxeval.AssertPure,
		Fn0:      func(env *sxeval.Environment, _ *sxeval.Binding) error { env.Push(nil); return nil },
		Fn1: func(env *sxeval.Environment, bind *sxeval.Binding) error {
			env.Set(sx.MakeList())
			return nil
		},
		Fn: func(env *sxeval.Environment, numargs int, _ *sxeval.Binding) error {
			obj := sx.MakeList(env.Args(numargs)[1:]...)
			env.Kill(numargs - 1)
			env.Set(obj)
			return nil
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

	env := sxeval.MakeEnvironment()
	args := sx.Vector{}
	for i := range 10 {
		env.PushArgs(args)
		if err := b.ExecuteCall(env, len(args), nil); err != nil {
			if size := env.Size(); size > 0 {
				t.Error("stack not empty, size:", size, i, err)
			}
			t.Error(err)
			break
		}
		res := env.Pop()
		if size := env.Size(); size > 0 {
			t.Error("stack not empty, size:", size, i)
		}
		if res != nil {
			if !sx.IsList(res) {
				t.Errorf("%d: result should be a list, but is not: %v", i, res)
			}
			exp := len(args) - 1
			if got := res.(*sx.Pair).Length(); got != exp {
				t.Errorf("Result list %v/%d must be one element shorter than arg %v/%d", res, got, args, exp)
			}
		}
		args = append(args, sx.Nil())
	}
}
