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

func TestFrame(t *testing.T) {
	t.Parallel()
	tcsFrame.Run(t)
}

var tcsFrame = tTestCases{
	{name: "err-current-frame-1",
		src:     "(current-frame 1)",
		exp:     "{[{current-frame: exactly 0 arguments required, but 1 given: [1]}]}",
		withErr: true,
	},

	{name: "err-parent-frame-0",
		src:     "(parent-frame)",
		exp:     "{[{parent-frame: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-parent-frame-1-noenv",
		src:     "(parent-frame 1)",
		exp:     "{[{parent-frame: argument 1 is not a frame, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "parent-frame-current", src: "(parent-frame (current-frame))", exp: "()"},
	// {name: "parent-frame-1-root", src: "(parent-frame ROOT)", exp: "#<undefined>"},

	{name: "err-bindings-0",
		src:     "(bindings)",
		exp:     "{[{bindings: exactly 1 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-bindings-1-noenv",
		src:     "(bindings 1)",
		exp:     "{[{bindings: argument 1 is not a frame, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "current-bindings", src: "(bindings (current-frame))", exp: "()"},
	{name: "let-bindings",
		src: "(let ((a 3)) (bindings (current-frame)))",
		exp: "((a . 3))",
	},

	{name: "err-lookup-0",
		src:     "(frame-lookup)",
		exp:     "{[{frame-lookup: between 1 and 2 arguments required, but none given}]}",
		withErr: true,
	},
	{name: "err-lookup-1-nosym",
		src:     "(frame-lookup 1)",
		exp:     "{[{frame-lookup: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-lookup-1-nosym-2",
		src:     "(frame-lookup 1 2)",
		exp:     "{[{frame-lookup: argument 1 is not a symbol, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-lookup-2-noenv",
		src:     "(frame-lookup 'a 1)",
		exp:     "{[{frame-lookup: argument 2 is not a frame, but sx.Int64/1}]}",
		withErr: true,
	},
	{name: "err-lookup-3-noenv",
		src:     "(frame-lookup 'a (current-frame) 1)",
		exp:     "{[{frame-lookup: between 1 and 2 arguments required, but 3 given: [a <nil> 1]}]}",
		withErr: true,
	},
	// {name: "lookup-a", src: "(defvar a 3)(binding-lookup 'a)", exp: "3 3"},
	{name: "lookup-a", src: "(let ((a 3)) (frame-lookup 'a))", exp: "3"},
	{name: "lookup-a-2", src: "(let ((a 3)) (frame-lookup 'a (current-frame)))", exp: "3"},
	{name: "lookup-b", src: "(frame-lookup 'b)", exp: "#<undefined>"},
	{name: "lookup-b-2", src: "(frame-lookup 'b (current-frame))", exp: "#<undefined>"},
}
