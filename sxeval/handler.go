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

package sxeval

import (
	"fmt"
	"math"

	"t73f.de/r/sx"
)

// DefaultComputeHandler is a ComputeHandler that just compute the expression.
//
// It is used to terminate a handler chain.
type DefaultComputeHandler struct{}

// Compute the expression in the environment, within a frame.
func (DefaultComputeHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
	return expr.Compute(env, frame)
}

// Reset the handler.
func (DefaultComputeHandler) Reset() {}

// LimitNestingHandler limits nested compute call to a given maximum.
type LimitNestingHandler struct {
	next         ComputeHandler
	currNesting  int
	maxNesting   int
	limitNesting int
}

// MakeLimitNestingHandler builds a new LimitNestingHandler.
func MakeLimitNestingHandler(limitNesting int, next ComputeHandler) *LimitNestingHandler {
	if limitNesting <= 0 {
		limitNesting = math.MaxInt
	}
	return &LimitNestingHandler{
		next:         next,
		limitNesting: limitNesting,
	}
}

// Compute the expression in the environment, within a frame.
func (h *LimitNestingHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
	h.currNesting++
	level := h.currNesting
	if level > h.maxNesting {
		h.maxNesting = level
	}
	if level > h.limitNesting {
		return nil, ErrNestingLimit{h.limitNesting}
	}

	obj, err := h.next.Compute(env, expr, frame)
	h.currNesting--
	return obj, err
}

// Reset the handler.
func (h *LimitNestingHandler) Reset() { h.currNesting = 0; h.next.Reset() }

// MaxNesting returns the maximum nesting value occured.
func (h *LimitNestingHandler) MaxNesting() int { return h.maxNesting }

// ErrNestingLimit is an error to be returned if the nesting limit is exceeded.
type ErrNestingLimit struct{ level int }

func (e ErrNestingLimit) Error() string { return fmt.Sprintf("nesting level exceeded: %d", e.level) }

// StepsLimitHandler limits the number of compute steps.
type StepsLimitHandler struct {
	next       ComputeHandler
	currSteps  int
	limitSteps int
}

// MakeStepsLimitHandler builds a new StepsLimitHandler.
func MakeStepsLimitHandler(limitSteps int, next ComputeHandler) *StepsLimitHandler {
	if limitSteps <= 0 {
		limitSteps = math.MaxInt
	}
	return &StepsLimitHandler{
		next:       next,
		limitSteps: limitSteps,
	}
}

// Compute the expression in the environment, within a frame.
func (h *StepsLimitHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
	h.currSteps++
	if h.currSteps > h.limitSteps {
		return nil, ErrStepsLimit{h.limitSteps}
	}

	return h.next.Compute(env, expr, frame)
}

// Reset the handler.
func (h *StepsLimitHandler) Reset() { h.currSteps = 0; h.next.Reset() }

// Steps returns the observed number of computation steps.
func (h *StepsLimitHandler) Steps() int { return h.currSteps }

// ErrStepsLimit is an error to signal that the number of computation steps is exceeded.
type ErrStepsLimit struct{ steps int }

func (e ErrStepsLimit) Error() string { return fmt.Sprintf("compute steps exceeded: %d", e.steps) }
