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
	"io"
	"strconv"
	"strings"

	"t73f.de/r/sx"
)

// Compilable is an interface, Expr should implement if they support compilation.
type Compilable interface {
	Compile(*Compiler) error
}

// Compiler is the data to be used at compilation time.
type Compiler struct {
	env      *Environment
	observer CompileObserver
	program  []Instruction
	curStack int
	maxStack int
}

// Instruction is a compiled command to execute one aspect of an expression.
type Instruction func(*Environment) error

// CompileObserver monitors the inner workings of the compilation.
type CompileObserver interface {
	LogCompile(*Compiler, string, ...string)
}

// Stats returns some basic statistics of the Compiler: length of program,
// current stack position, maximum stack position
func (sxc *Compiler) Stats() (int, int, int) { return len(sxc.program), sxc.curStack, sxc.maxStack }

// Compile the given expression. Do not call `expr.Compile()` directly.
func (sxc *Compiler) Compile(expr Expr) error {
	if iexpr, ok := expr.(Compilable); ok {
		return iexpr.Compile(sxc)
	}
	return MissingCompileError{Expr: expr}
}

// AdjustStack to track the current (and maximum) position of the abstract stack pointer.
func (sxc *Compiler) AdjustStack(offset int) {
	sxc.curStack += offset
	if sxc.curStack < 0 {
		panic("negative stack position")
	}
	sxc.maxStack = max(sxc.maxStack, sxc.curStack)
	if ob := sxc.observer; ob != nil {
		ob.LogCompile(sxc, "ADJUST", strconv.Itoa(offset))
	}
}

// Emit a threaded code.
func (sxc *Compiler) Emit(fn Instruction, s string, vals ...string) {
	if ob := sxc.observer; ob != nil {
		ob.LogCompile(sxc, s, vals...)
	}
	sxc.program = append(sxc.program, fn)
}

// TODO: EmitNOP that returns a patch function, to be used for jumps

// EmitBCall emits an instruction to call a builtin with more than two args
func (sxc *Compiler) EmitBCall(b *Builtin, numargs int) {
	sxc.AdjustStack(-numargs + 1)
	// TODO: cache fn
	fn := func(env *Environment) error {
		obj, err := b.Fn(env, env.Args(numargs))
		env.Kill(numargs - 1)
		env.Set(obj)
		return b.handleCallError(err)
	}
	sxc.Emit(fn, "BCALL", strconv.Itoa(numargs), b.Name)
}

// EmitBCall0 emits an instruction to call a builtin with no args
func (sxc *Compiler) EmitBCall0(b *Builtin) {
	sxc.AdjustStack(1)
	// TODO: cache fn
	fn := func(env *Environment) error {
		obj, err := b.Fn0(env)
		env.Push(obj)
		return b.handleCallError(err)
	}
	sxc.Emit(fn, "BCALL-0", b.Name)
}

// EmitBCall1 emits an instruction to call a builtin with one arg
func (sxc *Compiler) EmitBCall1(b *Builtin) {
	// TODO: cache fn
	fn := func(env *Environment) error {
		obj, err := b.Fn1(env, env.Top())
		env.Set(obj)
		return b.handleCallError(err)
	}
	sxc.Emit(fn, "BCALL-1", b.Name)
}

// EmitBCall2 emits an instruction to call a builtin with two args
func (sxc *Compiler) EmitBCall2(b *Builtin) {
	sxc.AdjustStack(-1)
	// TODO: cache fn
	fn := func(env *Environment) error {
		val1, val0 := env.Pop(), env.Top()
		obj, err := b.Fn2(env, val0, val1)
		env.Set(obj)
		return b.handleCallError(err)
	}
	sxc.Emit(fn, "BCALL-2", b.Name)
}

// MissingCompileError is signaled if an expression cannot be compiled
type MissingCompileError struct {
	Expr Expr // Expression unable to compile
}

func (mc MissingCompileError) Error() string {
	var sb strings.Builder
	_, _ = sb.WriteString("unable to compile: ")
	_, _ = mc.Expr.Print(&sb)
	return sb.String()
}

// ----- CompiledExpr: the result of a compilation

var _ Expr = (*CompiledExpr)(nil)

// CompiledExpr is an expression that contains a program.
type CompiledExpr struct {
	program   []Instruction
	stacksize int
	source    Expr // Source of program
}

// Unparse the expression as an sx.Object
func (comp *CompiledExpr) Unparse() sx.Object { return &ExprObj{expr: comp} }

// Compute the expression in an environment and return the result.
// It may have side-effects, on the given environment, or on the
// general environment of the system.
func (comp *CompiledExpr) Compute(env *Environment) (sx.Object, error) {
	program := comp.program
	ip := 0

	for ip < len(program) {
		err := program[ip](env)
		if err != nil {
			// TODO: ip recalc
			return nil, err
		}
		ip++
	}
	return env.Pop(), nil
}

// Print the expression on the given writer.
func (comp *CompiledExpr) Print(w io.Writer) (int, error) {
	length, err := fmt.Fprintf(w, "{COMPILED %d %d ", comp.stacksize, len(comp.program))
	if err != nil {
		return length, err
	}
	l, err := comp.source.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
