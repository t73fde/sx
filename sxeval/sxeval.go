//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package sxeval allows to evaluate s-expressions. Evaluation is splitted into
// parsing that s-expression and executing the result of the parsed expression.
// This is done to reduce syntax checks.
package sxeval

import (
	"fmt"
	"strings"

	"zettelstore.de/sx.fossil"
)

// Callable is a value that can be called for evaluation.
type Callable interface {
	// Call the value with the given args and frame.
	Call(*Frame, []sx.Object) (sx.Object, error)
}

// GetCallable returns the object as a Callable, if possible.
func GetCallable(obj sx.Object) (Callable, bool) {
	res, ok := obj.(Callable)
	return res, ok
}

// CallError encapsulate an error that occured during a call.
type CallError struct {
	Name string
	Err  error
}

func (e CallError) Unwrap() error { return e.Err }
func (e CallError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Name)
	sb.WriteString(": ")
	sb.WriteString(e.Err.Error())
	return sb.String()
}

// Executor is about controlling the execution of expressions.
// Do not call the method `Expr.Execute` directly, call the executor to do this.
type Executor interface {
	// Execute the expression in a frame and return the result.
	// It may have side-effects, on the given frame, or on the
	// general environment of the system.
	Execute(*Frame, Expr) (sx.Object, error)
}

type simpleExecutor struct{}

var mySimpleExecutor simpleExecutor

// Execute the given expression in the given environment of the given engine.
func (*simpleExecutor) Execute(frame *Frame, expr Expr) (sx.Object, error) {
	return expr.Compute(frame)
}

// Engine is the collection of all relevant data element to execute / evaluate an object.
type Engine struct {
	sf       sx.SymbolFactory
	root     Environment
	toplevel Environment
	pars     Parser
	exec     Executor
	bNames   map[uintptr]string
}

// MakeEngine creates a new engine.
func MakeEngine(sf sx.SymbolFactory, root Environment) *Engine {
	return &Engine{
		sf:       sf,
		root:     root,
		toplevel: root,
		pars:     &myDefaultParser,
		exec:     nil,
		bNames:   make(map[uintptr]string, 128),
	}
}

// SymbolFactory returns the symbol factory of the engine.
func (eng *Engine) SymbolFactory() sx.SymbolFactory { return eng.sf }

// RootEnvironment returns the root environment of the engine.
func (eng *Engine) RootEnvironment() Environment { return eng.root }

// SetToplevelEnv sets the given environment as the top-level environment.
// It must be the root environment or a child of it.
func (eng *Engine) SetToplevelEnv(env Environment) error {
	root := RootEnv(env)
	if root != eng.root {
		return fmt.Errorf("root of %v is not root of engine %v: %v", env, eng.root, root)
	}
	eng.toplevel = env
	return nil
}

// GetToplevelEnv returns the current top-level environment.
func (eng *Engine) GetToplevelEnv() Environment { return eng.toplevel }

// SetParser updates the current s-expression parser of the engine.
func (eng *Engine) SetParser(p Parser) Parser {
	orig := eng.pars
	if p == nil {
		p = &myDefaultParser
	}
	eng.pars = p
	return orig
}

// SetExecutor updates the executor pf parsed s-expressions.
func (eng *Engine) SetExecutor(e Executor) Executor {
	orig := eng.exec
	eng.exec = e
	if orig != nil {
		return orig
	}
	return &mySimpleExecutor
}

// SetQuote sets the quote symbol. It must be the same as for the reader.
func (eng *Engine) SetQuote(sym *sx.Symbol) error {
	if sym == nil {
		sym = eng.sf.MustMake("quote")
	}
	return eng.BindSyntax(sym.Name(), func(_ *ParseFrame, args *sx.Pair) (Expr, error) {
		if sx.IsNil(args) {
			return nil, ErrNoArgs
		}
		if args.Tail() != nil {
			return nil, fmt.Errorf("more than one argument: %v", args)
		}
		return ObjExpr{Obj: args.Car()}, nil
	})
}

// Eval parses the given object and executes it in the environment.
func (eng *Engine) Eval(env Environment, obj sx.Object) (sx.Object, error) {
	expr, err := eng.Parse(env, obj)
	if err != nil {
		return nil, err
	}
	expr = eng.Rework(env, expr)
	return eng.Execute(env, expr)
}

// Parse the given object in the given environment.
func (eng *Engine) Parse(env Environment, obj sx.Object) (Expr, error) {
	pf := ParseFrame{engine: eng, env: env, parser: eng.pars}
	return pf.Parse(obj)
}

// Rework the given expression with the options stored in the engine.
func (eng *Engine) Rework(env Environment, expr Expr) Expr {
	rf := ReworkFrame{env: env, engine: eng}
	return expr.Rework(&rf)
}

// Execute the given expression in the given environment.
func (eng *Engine) Execute(env Environment, expr Expr) (sx.Object, error) {
	frame := Frame{
		engine:   eng,
		executor: eng.exec,
		env:      env,
		caller:   nil,
	}
	return frame.Execute(expr)
}

// BindBuiltin binds the given builtin with its given name in the engine's
// root environment.
func (eng *Engine) BindBuiltin(b *Builtin) error {
	return eng.BindConst(b.Name, b)
}

// BindSyntax binds a syntax parser to the given name in the engine's root environment.
// It also binds the parser to the symbol directly.
func (eng *Engine) BindSyntax(name string, fn SyntaxFn) error {
	return eng.BindConst(name, MakeSyntax(name, fn))
}

// BindConst a given object to a symbol of the given name as a constant in the
// engine's root environment.
func (eng *Engine) BindConst(name string, obj sx.Object) error {
	return eng.root.BindConst(eng.sf.MustMake(name), obj)
}

// Bind a given object to a symbol of the given name in the engine's root environment.
func (eng *Engine) Bind(name string, obj sx.Object) error {
	return eng.root.Bind(eng.sf.MustMake(name), obj)
}
