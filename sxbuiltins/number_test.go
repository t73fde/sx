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

import "testing"

func TestNumber(t *testing.T) {
	t.Parallel()
	tcsNumber.Run(t)
}

var tcsNumber = tTestCases{
	{name: "err-number?-0",
		src:     "(number?)",
		exp:     "{[{number?: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "number?-1", src: "(number? 1)", exp: "T"},
	{name: "number?-nil", src: "(number? ())", exp: "()"},
	{name: "number?-sym", src: "(number? 'number?)", exp: "()"},

	{name: "add-0", src: "(+)", exp: "0"},
	{name: "add-1", src: "(+ 1)", exp: "1"},
	{name: "add-2", src: "(+ 3 4)", exp: "7"},
	{name: "add-5", src: "(+ 3 4 5 10 21)", exp: "43"},
	{name: "err-add-3",
		src:     "(+ 1 () 3)",
		exp:     "{[{+: argument 2 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},

	{name: "err-sub-0",
		src:     "(-)",
		exp:     "{[{-: at least 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-sub-nonum",
		src:     "(- 'a)",
		exp:     "{[{-: argument 1 is not a number, but *sx.Symbol/a}]}",
		withErr: true,
	},
	{name: "sub-1", src: "(- 1)", exp: "-1"},
	{name: "sub-2", src: "(- 3 4)", exp: "-1"},
	{name: "sub-5", src: "(- 3 4 5 10 21)", exp: "-37"},
	{name: "err-sub-2",
		src:     "(- () 3)",
		exp:     "{[{-: argument 1 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "err-sub-3",
		src:     "(- 1 () 3)",
		exp:     "{[{-: argument 2 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},

	{name: "mul-0", src: "(*)", exp: "1"},
	{name: "mul-1", src: "(* 3)", exp: "3"},
	{name: "mul-2", src: "(* 3 4)", exp: "12"},
	{name: "mul-5", src: "(* 3 4 5 10 21)", exp: "12600"},
	{name: "err-mul-3",
		src:     "(* 1 () 3)",
		exp:     "{[{*: argument 2 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},

	{name: "err-div-0",
		src:     "(div)",
		exp:     "{[{div: exactly 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-div-nonum-1",
		src:     "(div () 45)",
		exp:     "{[{div: argument 1 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "err-div-nonum-2",
		src:     "(div 45 ())",
		exp:     "{[{div: argument 2 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "err-div-zero",
		src:     "(div 45 0)",
		exp:     "{[{div: number zero not allowed}]}",
		withErr: true,
	},
	{name: "div-full", src: "(div 35 7)", exp: "5"},
	{name: "div-rest", src: "(div 34 7)", exp: "4"},

	{name: "err-mod-0",
		src:     "(mod)",
		exp:     "{[{mod: exactly 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-mod-nonum-1",
		src:     "(mod () 45)",
		exp:     "{[{mod: argument 1 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "err-mod-nonum-2",
		src:     "(mod 45 ())",
		exp:     "{[{mod: argument 2 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "err-mod-zero",
		src:     "(mod 45 0)",
		exp:     "{[{mod: number zero not allowed}]}",
		withErr: true,
	},
	{name: "mod-full", src: "(mod 35 7)", exp: "0"},
	{name: "mod-rest", src: "(mod 34 7)", exp: "6"},
}
