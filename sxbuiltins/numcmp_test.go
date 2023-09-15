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

import "testing"

func TestNumCmp(t *testing.T) {
	t.Parallel()
	tcsNumCmp.Run(t)
}

var tcsNumCmp = tTestCases{
	{
		name:    "err-less-0",
		src:     "(<)",
		exp:     "{[{<: at least 2 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-less-nonum",
		src:     "(< 0 1 ())",
		exp:     "{[{<: argument 3 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "less-2", src: "(< 1 2)", exp: "1"},
	{name: "less-5", src: "(< 1 1 3 4 4)", exp: "()"},
	{name: "less-6", src: "(< 1 2 3 4 0 6)", exp: "()"},

	{
		name:    "err-less-equal-0",
		src:     "(<=)",
		exp:     "{[{<=: at least 2 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-less-equal-nonum",
		src:     "(<= 0 1 ())",
		exp:     "{[{<=: argument 3 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "less-equal-2", src: "(<= 1 2)", exp: "1"},
	{name: "less-equal-5", src: "(<= 1 1 3 4 4)", exp: "1"},
	{name: "less-equal-6", src: "(<= 1 2 3 4 0 6)", exp: "()"},

	{
		name:    "err-equal-0",
		src:     "(=)",
		exp:     "{[{=: at least 2 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{name: "equal-nonum", src: "(= 0 0 ())", exp: "()"},
	{name: "equal-2", src: "(= 3 3)", exp: "1"},
	{name: "equal-3", src: "(= 3 2 ())", exp: "()"},
	{name: "equal-5", src: "(= 4 4 4 4 4)", exp: "1"},
	{name: "equal-6", src: "(= 4 4 4 4 0 6)", exp: "()"},

	{
		name:    "err-greater-equal-0",
		src:     "(>=)",
		exp:     "{[{>=: at least 2 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-greater-equal-nonum",
		src:     "(>= 10 1 ())",
		exp:     "{[{>=: argument 3 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "greater-equal-2", src: "(>= 2 1)", exp: "1"},
	{name: "greater-equal-5", src: "(>= 4 4 3 1 1)", exp: "1"},
	{name: "greater-equal-6", src: "(>= 6 0 4 2 1)", exp: "()"},

	{
		name:    "err-greater-0",
		src:     "(>)",
		exp:     "{[{>: at least 2 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-greater-nonum",
		src:     "(> 10 1 ())",
		exp:     "{[{>: argument 3 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "greater-2", src: "(> 2 1)", exp: "1"},
	{name: "greater-5", src: "(> 6 4 3 1 1)", exp: "()"},
	{name: "greater-6", src: "(> 6 4 3 0 1 0)", exp: "()"},

	{
		name:    "err-min-0",
		src:     "(min)",
		exp:     "{[{min: at least 1 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-min-nonum",
		src:     "(min 0 1 ())",
		exp:     "{[{min: argument 3 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "min-1", src: "(min 1)", exp: "1"},
	{name: "min-2", src: "(min 1 2)", exp: "1"},
	{name: "min-5", src: "(min 1 1 3 4 4)", exp: "1"},
	{name: "min-6", src: "(min 1 2 3 4 0 6)", exp: "0"},

	{
		name:    "err-max-0",
		src:     "(max)",
		exp:     "{[{max: at least 1 arguments required, but only 0 given: []}]}",
		withErr: true,
	},
	{
		name:    "err-max-nonum",
		src:     "(max 0 1 ())",
		exp:     "{[{max: argument 3 is not a number, but *sx.Pair/()}]}",
		withErr: true,
	},
	{name: "max-1", src: "(max 1)", exp: "1"},
	{name: "max-2", src: "(max 1 2)", exp: "2"},
	{name: "max-5", src: "(max 1 1 3 4 4)", exp: "4"},
	{name: "max-6", src: "(max 1 2 3 4 0 6)", exp: "6"},
}
