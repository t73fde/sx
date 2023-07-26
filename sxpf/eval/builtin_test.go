//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package eval_test

import (
	"strings"
	"testing"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/eval"
)

func TestBuiltinSimple(t *testing.T) {
	b := eval.BuiltinA(func(args []sxpf.Object) (sxpf.Object, error) {
		if len(args) == 0 {
			return nil, nil
		}
		return sxpf.MakeList(args[1:]...), nil
	})
	if sxpf.IsNil(b) {
		t.Error("Builtin is wrongly treated as Nil()", b)
		return
	}
	expString := "#<builtin:"
	if got := b.String(); !strings.HasPrefix(got, expString) {
		t.Errorf("Builtin.String() should start with %q, but got %q", expString, got)
	}
	expLen := len(expString)
	var sb strings.Builder
	if got, err := sxpf.Print(&sb, b); err != nil || got <= expLen {
		if err != nil {
			t.Errorf("Builtin %v.Print() resulted in error %v", b, err)
		} else if got != expLen {
			t.Errorf("Builtin %v.Print() should deliver %d bytes, but got %d", b, expLen, got)
		}
	}

	args := []sxpf.Object{}
	for i := 0; i < 10; i++ {
		res, err := b.Call(nil, nil, args)
		if err != nil {
			t.Error(err)
			break
		}
		if res != nil {
			if !sxpf.IsList(res) {
				t.Errorf("%d: result should be a list, but is not: %v", i, res)
			}
			exp := len(args) - 1
			if got := res.(*sxpf.Pair).Length(); got != exp {
				t.Errorf("Result list %v/%d must be one element shorter than arg %v/%d", res, got, args, exp)
			}
		}
		args = append(args, sxpf.Nil())
	}
}
