//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins_test

import "testing"

func TestVector(t *testing.T) {
	t.Parallel()
	tcsVector.Run(t)
}

var tcsVector = tTestCases{
	{name: "vector-0", src: "(vector)", exp: "()"},
	{name: "vector-1", src: "(vector 4)", exp: "(vector 4)"},
	{name: "vector-2", src: "(vector 4 7)", exp: "(vector 4 7)"},
	{name: "vector-alias-args",
		src: "(let ((a (vector b 7 9)) (b (vector 7 7 b))) (+ (apply length `(,b)) 4) a)",
		exp: "(vector 11 7 9)"},

	{name: "err-vector?-0",
		src:     "(vector?)",
		exp:     "{[{vector?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-vector?-2",
		src:     "(vector? 1 2)",
		exp:     "{[{vector?: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "vector?-nil", src: "(vector? ())", exp: "T"},
	{name: "vector?-1", src: "(vector? 1)", exp: "()"},
	{name: "vector?-cons", src: "(vector? (cons 1 2))", exp: "()"},
	{name: "vector?-vector-0", src: "(vector? (vector))", exp: "T"},
	{name: "vector?-vector-1", src: "(vector? (vector 1))", exp: "T"},

	{name: "err-vset!-0",
		src:     "(vset!)",
		exp:     "{[{vset!: exactly 3 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-vset!-1",
		src:     "(vset! 1)",
		exp:     "{[{vset!: exactly 3 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},
	{name: "err-vset!-2",
		src:     "(vset! 1 2)",
		exp:     "{[{vset!: exactly 3 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "err-vset!-3",
		src:     "(vset! 1 2 3 4)",
		exp:     "{[{vset!: exactly 3 arguments required, but 4 given: [1 2 3 4]}]}",
		withErr: true,
	},
	{name: "err-vset!-num.vector",
		src:     "(vset! 1 2 3)",
		exp:     "{[{vset!: argument 1 is not a vector, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-vset!-nil.vector",
		src:     "(vset! (list 1) 2 3)",
		exp:     "{[{vset!: argument 1 is not a vector, but *sx.Pair/(1)}]}",
		withErr: true,
	},
	{name: "err-vset!-nonum",
		src:     "(vset! (vector 1) 'a 'b)",
		exp:     "{[{vset!: argument 2 is not a number, but *sx.Symbol/a}]}",
		withErr: true,
	},
	{name: "err-vset!-pos-min",
		src:     "(vset! (vector) -1 7)",
		exp:     "{[{vset!: negative vector index not allowed: -1}]}",
		withErr: true,
	},
	{name: "err-vset!-empty-pos-max",
		src:     "(vset! (vector) 0 7)",
		exp:     "{[{vset!: vector index out of range: 0}]}",
		withErr: true,
	},
	{name: "err-vset!-one-pos-max",
		src:     "(vset! (vector 1) 1 7)",
		exp:     "{[{vset!: vector index out of range: 1}]}",
		withErr: true,
	},
	{name: "vset!-one", src: "(vset! (vector 1) 0 3)", exp: "(vector 3)"},
	{name: "vset!-two-one", src: "(vset! (vector 1 2) 0 3)", exp: "(vector 3 2)"},
	{name: "vset!-two-zwo", src: "(vset! (vector 1 2) 1 3)", exp: "(vector 1 3)"},
}
