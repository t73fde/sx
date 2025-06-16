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

func TestEval(t *testing.T) {
	t.Parallel()
	tcsEval.Run(t)
}

var tcsEval = tTestCases{
	{name: "err-parse-expression-0",
		src:     "(parse-expression)",
		exp:     "{[{parse-expression: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-parse-expression-3",
		src:     "(parse-expression 1 2 3)",
		exp:     "{[{parse-expression: between 1 and 2 arguments required, but 3 given: [1 2 3]}]}",
		withErr: true,
	},
	{name: "err-parse-expression-err-parse",
		src:     "(parse-expression '(set!))",
		exp:     "{[{set!: need at least two arguments}]}",
		withErr: true,
	},
	{name: "err-parse-expression-2",
		src:     "(parse-expression '(set!) 3)",
		exp:     "{[{parse-expression: argument 2 is not a frame, but sx.Int64/3}]}",
		withErr: true,
	},
	{name: "err-parse-expression-err-parse-2",
		src:     "(parse-expression '(set!) (current-frame))",
		exp:     "{[{set!: need at least two arguments}]}",
		withErr: true,
	},
	{name: "parse-expression-one", src: "(parse-expression 1)", exp: "#<{OBJ 1}>"},
	{name: "parse-expression-one-2", src: "(parse-expression 1 (current-frame))", exp: "#<{OBJ 1}>"},

	{name: "err-unparse-expression-0",
		src:     "(unparse-expression)",
		exp:     "{[{unparse-expression: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-unparse-expression-2",
		src:     "(unparse-expression 1 2)",
		exp:     "{[{unparse-expression: exactly 1 arguments required, but 2 given: [1 2]}]}",
		withErr: true,
	},
	{name: "err-unparse-expression-1",
		src:     "(unparse-expression 1)",
		exp:     "{[{unparse-expression: argument 1 is not an expression, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "unparse-expression-one", src: "(unparse-expression (parse-expression '(1)))", exp: "(1)"},

	{name: "err-execute-expression-0",
		src:     "(execute-expression)",
		exp:     "{[{execute-expression: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-execute-expression-3",
		src:     "(execute-expression 1 2 3)",
		exp:     "{[{execute-expression: between 1 and 2 arguments required, but 3 given: [1 2 3]}]}",
		withErr: true,
	},
	{name: "err-execute-expression-1",
		src:     "(execute-expression 1)",
		exp:     "{[{execute-expression: argument 1 is not an expression, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-execute-expression-2",
		src:     "(execute-expression 1 2)",
		exp:     "{[{execute-expression: argument 1 is not an expression, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-execute-expression-2-bind",
		src:     "(execute-expression (parse-expression '1) 2)",
		exp:     "{[{execute-expression: argument 2 is not a frame, but sx.Int64/2}]}",
		withErr: true,
	},
	{name: "execute-expression-one", src: "(execute-expression (parse-expression '1))", exp: "1"},
	{name: "execute-expression-two", src: "(execute-expression (parse-expression '1)  (current-frame))", exp: "1"},

	{name: "err-eval-0",
		src:     "(eval)",
		exp:     "{[{eval: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-eval-3",
		src:     "(eval 1 2 3)",
		exp:     "{[{eval: between 1 and 2 arguments required, but 3 given: [1 2 3]}]}",
		withErr: true,
	},
	{name: "err-eval-err-parse",
		src:     "(eval '(set!))",
		exp:     "{[{set!: need at least two arguments}]}",
		withErr: true,
	},
	{name: "err-eval-2",
		src:     "(eval '(set!) 2)",
		exp:     "{[{eval: argument 2 is not a frame, but sx.Int64/2}]}",
		withErr: true,
	},
	{name: "err-eval-err-parse",
		src:     "(eval '(set!) (current-frame))",
		exp:     "{[{set!: need at least two arguments}]}",
		withErr: true,
	},
	{name: "eval-one", src: "(eval 1)", exp: "1"},
	{name: "eval-one-parse", src: "(eval (parse-expression 1))", exp: "1"},
	{name: "eval-two", src: "(eval 1 (current-frame))", exp: "1"},
}
