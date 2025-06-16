//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

// Provides some special/builtin functions to work with frames.

import (
	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// CurrentFrame returns the current frame.
var CurrentFrame = sxeval.Builtin{
	Name:     "current-frame",
	MinArity: 0,
	MaxArity: 0,
	TestPure: nil,
	Fn0: func(_ *sxeval.Environment, frame *sxeval.Frame) (sx.Object, error) {
		return frame, nil
	},
}

// ParentFrame returns the parent frame of the given frame. For the
// top-most frame, NIL is returned.
var ParentFrame = sxeval.Builtin{
	Name:     "parent-frame",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		frame, err := GetFrame(arg, 0)
		if err != nil {
			return nil, err
		}
		return frame.Parent(), nil
	},
}

// Bindings returns the given frame as an association list.
var Bindings = sxeval.Builtin{
	Name:     "bindings",
	MinArity: 1,
	MaxArity: 1,
	TestPure: sxeval.AssertPure,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		frame, err := GetFrame(arg, 0)
		if err != nil {
			return nil, err
		}
		return frame.Bindings(), nil
	},
}

// FrameLookup returns the symbol's bound value in the given
// frame, or Undefined if there is no such value.
var FrameLookup = sxeval.Builtin{
	Name:     "frame-lookup",
	MinArity: 1,
	MaxArity: 2,
	TestPure: nil,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, frame *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(arg, 0)
		if err != nil {
			return nil, err
		}
		if obj, found := frame.Lookup(sym); found {
			return obj, nil
		}
		return sx.MakeUndefined(), nil
	},
	Fn: func(_ *sxeval.Environment, args sx.Vector, _ *sxeval.Frame) (sx.Object, error) {
		sym, err := GetSymbol(args[0], 0)
		if err != nil {
			return nil, err
		}
		if arg1 := args[1]; !sx.IsNil(arg1) {
			frame, err2 := GetFrame(arg1, 1)
			if err2 != nil {
				return nil, err2
			}
			if obj, found := frame.Lookup(sym); found {
				return obj, nil
			}
		}
		return sx.MakeUndefined(), nil
	},
}
