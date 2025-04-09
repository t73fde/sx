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
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"t73f.de/r/sx"
)

// Compiler is the data to be used at compilation time.
type Compiler struct {
	level    int
	env      *Environment
	observer CompileObserver
	program  []Instruction
	curStack int
	maxStack int

	bcallInstrCache  map[*Builtin]Instruction
	bcall0InstrCache map[*Builtin]Instruction
	bcall1InstrCache map[*Builtin]Instruction
}

// Instruction is a compiled command to execute one aspect of an expression.
type Instruction func(*Environment, *Binding) error

// CompileObserver monitors the inner workings of the compilation.
type CompileObserver interface {
	LogCompile(*Compiler, string, ...string)
}

// MakeChildCompiler builds a new compiler.
func (sxc *Compiler) MakeChildCompiler() *Compiler {
	result := *sxc
	result.level++
	result.resetState()
	return &result
}

func (sxc *Compiler) resetState() {
	sxc.program = nil
	sxc.curStack = 0
	sxc.maxStack = 0

	if sxc.bcallInstrCache == nil {
		sxc.bcallInstrCache = make(map[*Builtin]Instruction)
	}
	if sxc.bcall0InstrCache == nil {
		sxc.bcall0InstrCache = make(map[*Builtin]Instruction)
	}
	if sxc.bcall1InstrCache == nil {
		sxc.bcall1InstrCache = make(map[*Builtin]Instruction)
	}
}

// Stats returns some basic statistics of the Compiler: level,
// length of program, current stack position, maximum stack position
func (sxc *Compiler) Stats() (int, int, int, int) {
	return sxc.level, len(sxc.program), sxc.curStack, sxc.maxStack
}

// CompileProgram builds a ProgramExpr by compiling the Expr.
func (sxc *Compiler) CompileProgram(expr Expr) (*ProgramExpr, error) {
	sxc.resetState()

	if err := sxc.Compile(expr, true); err != nil {
		return nil, err
	}
	if sxc.curStack != 1 {
		return nil, fmt.Errorf("wrong stack position: %d", sxc.curStack)
	}
	return &ProgramExpr{
		program:   slices.Clip(sxc.program),
		stacksize: sxc.maxStack,
		level:     sxc.level,
		source:    expr,
	}, nil
}

// Compile the given expression. Do not call `expr.Compile()` directly.
func (sxc *Compiler) Compile(expr Expr, tailPos bool) error {
	return expr.Compile(sxc, tailPos)
}

// AdjustStack to track the current (and maximum) position of the abstract stack pointer.
func (sxc *Compiler) AdjustStack(offset int) {
	sxc.curStack += offset
	if sxc.curStack < 0 {
		panic("negative stack position")
	}
	sxc.maxStack = max(sxc.maxStack, sxc.curStack)
	if ob := sxc.observer; ob != nil {
		ob.LogCompile(sxc, "adjust", strconv.Itoa(offset))
	}
}

// Emit a threaded code.
func (sxc *Compiler) Emit(instr Instruction, s string, vals ...string) {
	if ob := sxc.observer; ob != nil {
		ob.LogCompile(sxc, s, vals...)
	}
	sxc.program = append(sxc.program, instr)
}

// EmitPush emits code to push a value on the stack.
func (sxc *Compiler) EmitPush(val sx.Object) {
	sxc.AdjustStack(1)
	if sx.IsNil(val) {
		sxc.Emit(func(env *Environment, _ *Binding) error {
			env.Push(sx.Nil())
			return nil
		}, "PUSH-NIL")
	} else {
		sxc.Emit(func(env *Environment, _ *Binding) error {
			env.Push(val)
			return nil
		}, "PUSH", val.String())
	}
}

// EmitKill emits code to remove some TOS elements.
func (sxc *Compiler) EmitKill(num int) {
	if num > 0 {
		sxc.AdjustStack(-num)
		sxc.Emit(func(env *Environment, _ *Binding) error { env.Kill(num); return nil }, "KILL", strconv.Itoa(num))
	}
}

// NopInstruction is an empty intruction. It does nothing.
func NopInstruction(*Environment, *Binding) error { return nil }

// EmitJumpNIL emits some preliminary code to jump if TOS is nil.
// It returns a patch function to update the jump target.
func (sxc *Compiler) EmitJumpNIL() func() {
	sxc.AdjustStack(-1)
	pc := len(sxc.program)
	sxc.Emit(NopInstruction, "JUMP-NIL", strconv.Itoa(pc), "<- to be patched")
	return func() {
		pos := len(sxc.program)
		if ob := sxc.observer; ob != nil {
			ob.LogCompile(sxc, "patch", strconv.Itoa(pc), "JUMP-NIL", strconv.Itoa(pos))
		}
		sxc.program[pc] = func(env *Environment, _ *Binding) error {
			val := env.Pop()
			if sx.IsNil(val) {
				return jumpToError{pos: pos}
			}
			return nil
		}
	}
}

// EmitJump emits some preliminary code to jump unconditionally.
// It returns a patch function to update the jump target.
func (sxc *Compiler) EmitJump() func() {
	pc := len(sxc.program)
	sxc.Emit(NopInstruction, "JUMP", strconv.Itoa(pc), "<- to be patched")
	return func() {
		pos := len(sxc.program)
		if ob := sxc.observer; ob != nil {
			ob.LogCompile(sxc, "patch", strconv.Itoa(pc), "JUMP", strconv.Itoa(pos))
		}
		sxc.program[pc] = func(*Environment, *Binding) error { return jumpToError{pos: pos} }
	}
}

// EmitBCall emits an instruction to call a builtin with more than two args
func (sxc *Compiler) EmitBCall(b *Builtin, numargs int) {
	sxc.AdjustStack(-numargs + 1)
	instr, found := sxc.bcallInstrCache[b]
	if !found {
		fn, handleCallError := b.Fn, b.handleCallError
		instr = func(env *Environment, bind *Binding) error {
			obj, err := fn(env, env.Args(numargs), bind)
			env.Kill(numargs - 1)
			env.Set(obj)
			return handleCallError(err)
		}
		sxc.bcallInstrCache[b] = instr
	}
	sxc.Emit(instr, "BCALL", strconv.Itoa(numargs), b.Name)
}

// EmitBCall0 emits an instruction to call a builtin with no args
func (sxc *Compiler) EmitBCall0(b *Builtin) {
	sxc.AdjustStack(1)
	instr, found := sxc.bcall0InstrCache[b]
	if !found {
		fn0, handleCallError := b.Fn0, b.handleCallError
		instr = func(env *Environment, bind *Binding) error {
			obj, err := fn0(env, bind)
			env.Push(obj)
			return handleCallError(err)
		}
		sxc.bcall0InstrCache[b] = instr
	}
	sxc.Emit(instr, "BCALL-0", b.Name)
}

// EmitBCall1 emits an instruction to call a builtin with one arg
func (sxc *Compiler) EmitBCall1(b *Builtin) {
	instr, found := sxc.bcall1InstrCache[b]
	if !found {
		fn1, handleCallError := b.Fn1, b.handleCallError
		instr = func(env *Environment, bind *Binding) error {
			obj, err := fn1(env, env.Top(), bind)
			env.Set(obj)
			return handleCallError(err)
		}
		sxc.bcall1InstrCache[b] = instr
	}
	sxc.Emit(instr, "BCALL-1", b.Name)
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

// ----- ProgramExpr: the result of a compilation

// ProgramExpr is an expression that contains a program.
type ProgramExpr struct {
	program   []Instruction
	stacksize int
	level     int
	source    Expr // Source of program
}

// Unparse the expression as an sx.Object
func (cpe *ProgramExpr) Unparse() sx.Object { return &ExprObj{expr: cpe} }

// Compile the expression: nothing to do since it is already compiled.
func (cpe *ProgramExpr) Compile(*Compiler, bool) error { return nil }

// Compute the expression in an environment and return the result.
// It may have side-effects, on the given environment, or on the
// general environment of the system.
func (cpe *ProgramExpr) Compute(env *Environment, bind *Binding) (sx.Object, error) {
	err := cpe.Interpret(env, bind)
	if err != nil {
		env.Reset()
		return sx.Nil(), err
	}
	return env.Pop(), nil
}

// Interpret the program in an environment.
func (cpe *ProgramExpr) Interpret(env *Environment, bind *Binding) error {
	currBind := bind
	program := cpe.program

	for ip := 0; ip < len(program); ip++ {
		if err := program[ip](env, currBind); err != nil {
			if err == errSwitchBinding {
				currBind = env.tco.binding
			} else if jerr, ok := err.(jumpToError); ok {
				ip = jerr.pos - 1
			} else if err == errExecuteAgain {
				currBind = env.tco.binding
				expr := env.tco.expr
				if newCpe, isProgram := expr.(*ProgramExpr); isProgram {
					program = newCpe.program
					ip = -1
					continue
				}
				obj, err := env.Execute(expr, currBind)
				if err == nil {
					env.Push(obj)
					continue
				}
				return err
			} else {
				return err
			}
		}
	}
	return nil
}

// Print the expression on the given writer.
func (cpe *ProgramExpr) Print(w io.Writer) (int, error) {
	length, err := fmt.Fprintf(w, "{COMPILED-%d %d %d ", cpe.level, cpe.stacksize, len(cpe.program))
	if err != nil {
		return length, err
	}
	l, err := cpe.source.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}

type jumpToError struct{ pos int }

func (jerr jumpToError) Error() string { return fmt.Sprintf("jump: %d", jerr.pos) }

var errSwitchBinding = errors.New("switch-binding")

// SwitchBinding make the interpreter to switch to a new binding.
func SwitchBinding(env *Environment, bind *Binding) error {
	env.tco.binding = bind
	return errSwitchBinding
}
