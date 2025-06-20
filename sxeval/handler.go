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

	"t73f.de/r/sx"
)

// DefaultHandler is a ComputeHandler that just compute the expression.
//
// It is used to terminate a handler chain.
type DefaultHandler struct{}

// Compute the expression in the environment, within a frame.
func (DefaultHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
	return expr.Compute(env, frame)
}

// Reset the handler.
func (DefaultHandler) Reset() {}

// ----- NestingHandler -----------------------------------------------------

// NestingHandler counts nested compute calls.
type NestingHandler struct {
	next        ComputeHandler
	currNesting int
	maxNesting  int
}

// MakeNestingHandler builds a new NestingHandler.
func MakeNestingHandler(next ComputeHandler) *NestingHandler {
	return &NestingHandler{next: next}
}

// Compute the expression in the environment, within a frame.
func (h *NestingHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
	h.currNesting++
	level := h.currNesting
	if level > h.maxNesting {
		h.maxNesting = level
	}

	obj, err := h.next.Compute(env, expr, frame)
	h.currNesting--
	return obj, err
}

// Reset the handler.
func (h *NestingHandler) Reset() { h.currNesting, h.maxNesting = 0, 0; h.next.Reset() }

// Nesting returns the current and the maximum nesting value occured.
func (h *NestingHandler) Nesting() (int, int) { return h.currNesting, h.maxNesting }

// ----- NestingLimitHandler ------------------------------------------------

// NestingLimitHandler limits nested compute call to a given maximum.
type NestingLimitHandler struct {
	next         ComputeHandler
	currNesting  int
	maxNesting   int
	limitNesting int
}

// MakeNestingLimitHandler builds a new NestingLimitHandler.
func MakeNestingLimitHandler(limitNesting int, next ComputeHandler) *NestingLimitHandler {
	return &NestingLimitHandler{
		next:         next,
		limitNesting: limitNesting,
	}
}

// Compute the expression in the environment, within a frame.
func (h *NestingLimitHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
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
func (h *NestingLimitHandler) Reset() { h.currNesting = 0; h.next.Reset() }

// Nesting returns the current and the maximum nesting value occured.
func (h *NestingLimitHandler) Nesting() (int, int) { return h.currNesting, h.maxNesting }

// ErrNestingLimit is an error to be returned if the nesting limit is exceeded.
type ErrNestingLimit struct{ level int }

func (e ErrNestingLimit) Error() string {
	return fmt.Sprintf("allowed nesting level exceeded: %d", e.level)
}

// ----- StepsHandler -------------------------------------------------------

// StepsHandler just counts the number of execution steps.
type StepsHandler struct {
	next  ComputeHandler
	Steps int
}

// MakeStepsHandler builds a new StepsHandler.
func MakeStepsHandler(next ComputeHandler) *StepsHandler {
	return &StepsHandler{next: next}
}

// Compute the expression in the environment, within a frame.
func (h *StepsHandler) Compute(env *Environment, expr Expr, frame *Frame) (sx.Object, error) {
	h.Steps++
	return h.next.Compute(env, expr, frame)
}

// Reset the handler.
func (h *StepsHandler) Reset() { h.Steps = 0; h.next.Reset() }

// ----- StepsLimitHandler --------------------------------------------------

// StepsLimitHandler limits the number of compute steps.
type StepsLimitHandler struct {
	next       ComputeHandler
	currSteps  int
	limitSteps int
}

// MakeStepsLimitHandler builds a new StepsLimitHandler.
func MakeStepsLimitHandler(limitSteps int, next ComputeHandler) *StepsLimitHandler {
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

func (e ErrStepsLimit) Error() string {
	return fmt.Sprintf("allowed compute steps exceeded: %d", e.steps)
}
