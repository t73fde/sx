//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxeval

import (
	"fmt"
	"io"
	"strconv"

	"zettelstore.de/sx.fossil"
)

// Environment maintains a mapping between symbols and values.
// Forms are evaluated within environments.
type Environment interface {
	// An environment is an object by itself
	sx.Object

	// String returns the local name of this environment.
	String() string

	// Parent allows to retrieve the parent environment. Environment is the root
	// environment, nil is returned. Lookups that cannot be satisfied in an
	// environment are delegated to the parent envrionment.
	Parent() Environment

	// IsRoot returns true for the root environment.
	IsRoot() bool

	// Bind creates a local mapping with a given symbol and object.
	//
	// A previous, non-const mapping will be overwritten.
	Bind(*sx.Symbol, sx.Object) error

	// BindConst creates a local mapping of the symbol to the object, which
	// cannot be changed afterwards.
	BindConst(*sx.Symbol, sx.Object) error

	// Lookup will search for a local binding of the given symbol. If not
	// found, the search will *not* be continued in the parent environment.
	// Use the global `Resolve` function, if you want a search up to the parent.
	Lookup(*sx.Symbol) (sx.Object, bool)

	// IsConst returns true if the binding of the symbol is a constant binding.
	IsConst(*sx.Symbol) bool

	// Bindings returns all bindings as an a-list in some random order.
	Bindings() *sx.Pair

	// Unbind removes the mapping of the given symbol to an object.
	Unbind(*sx.Symbol) error

	// Freeze sets the environment in a read-only state.
	Freeze()
}

// ErrEnvFrozen is returned when trying to update a frozen environment.
type ErrEnvFrozen struct{ Env Environment }

func (err ErrEnvFrozen) Error() string { return fmt.Sprintf("enviroment is frozen: %v", err.Env) }

// ErrConstBinding is returned when a constant binding should be changed.
type ErrConstBinding struct{ Sym *sx.Symbol }

func (err ErrConstBinding) Error() string {
	return fmt.Sprintf("constant bindung for symbol %v", err.Sym.Repr())
}

type mapSymObj = map[*sx.Symbol]sx.Object

// MakeRootEnvironment creates a new root environment.
func MakeRootEnvironment(sizeHint int) Environment {
	return &mappedEnvironment{
		name:   "root",
		parent: nil,
		vars:   make(mapSymObj, sizeHint),
		isRoot: true,
		frozen: false,
	}
}

// MakeChildEnvironment creates a new environment with a given parent.
func MakeChildEnvironment(parent Environment, name string, sizeHint int) Environment {
	if sizeHint <= 0 {
		sizeHint = 3
	}
	return &mappedEnvironment{
		name:   name,
		parent: parent,
		vars:   make(mapSymObj, sizeHint),
		isRoot: false,
		frozen: false,
	}
}

type mappedEnvironment struct {
	name   string
	parent Environment
	vars   mapSymObj
	consts map[*sx.Symbol]struct{}
	isRoot bool
	frozen bool
}

func (me *mappedEnvironment) IsNil() bool  { return me == nil }
func (me *mappedEnvironment) IsAtom() bool { return me == nil }
func (me *mappedEnvironment) IsEqual(other sx.Object) bool {
	if me == other {
		return true
	}
	if me.IsNil() {
		return sx.IsNil(other)
	}
	if ome, ok := other.(*mappedEnvironment); ok {
		mvars, ovars := me.vars, ome.vars
		if len(mvars) != len(ovars) {
			return false
		}
		for k, v := range mvars {
			ov, found := ovars[k]
			if !found || !v.IsEqual(ov) {
				return false
			}
		}
		return true
	}
	return false
}
func (me *mappedEnvironment) Repr() string { return sx.Repr(me) }
func (me *mappedEnvironment) Print(w io.Writer) (int, error) {
	return sx.WriteStrings(w, "#<env:", me.name, "/", strconv.Itoa(len(me.vars)), ">")
}
func (me *mappedEnvironment) String() string { return me.name }
func (me *mappedEnvironment) Parent() Environment {
	if me == nil {
		return nil
	}
	return me.parent
}
func (me *mappedEnvironment) IsRoot() bool { return me == nil || me.isRoot }
func (me *mappedEnvironment) Bind(sym *sx.Symbol, val sx.Object) error {
	if me.frozen {
		return ErrEnvFrozen{Env: me}
	}
	if me.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if _, found := me.vars[sym]; found {
		me.vars[sym] = val
		return nil
	}
	me.vars[sym] = val
	return nil
}
func (me *mappedEnvironment) BindConst(sym *sx.Symbol, val sx.Object) error {
	if me.frozen {
		return ErrEnvFrozen{Env: me}
	}
	if me.IsConst(sym) {
		return ErrConstBinding{Sym: sym}
	}
	if me.consts == nil {
		me.consts = map[*sx.Symbol]struct{}{sym: struct{}{}}
	} else {
		me.consts[sym] = struct{}{}
	}
	if _, found := me.vars[sym]; found {
		me.vars[sym] = val
		return nil
	}
	me.vars[sym] = val
	return nil
}
func (me *mappedEnvironment) Lookup(sym *sx.Symbol) (sx.Object, bool) {
	obj, found := me.vars[sym]
	return obj, found
}
func (me *mappedEnvironment) IsConst(sym *sx.Symbol) bool {
	if me == nil {
		return false
	}
	if me.frozen {
		if _, found := me.vars[sym]; found {
			return true
		}
	}
	if me.consts == nil {
		return false
	}
	_, found := me.consts[sym]
	return found
}
func (me *mappedEnvironment) Bindings() *sx.Pair {
	result := sx.Nil()
	for k, v := range me.vars {
		result = result.Cons(sx.Cons(k, v))
	}
	return result
}
func (me *mappedEnvironment) Unbind(sym *sx.Symbol) error {
	if me.frozen {
		return ErrEnvFrozen{Env: me}
	}
	delete(me.vars, sym)
	return nil
}
func (me *mappedEnvironment) Freeze() { me.frozen = true }

// GetEnvironment returns the object as an environment, if possible.
func GetEnvironment(obj sx.Object) (Environment, bool) {
	if sx.IsNil(obj) {
		return nil, false
	}
	env, ok := obj.(Environment)
	return env, ok
}

// RootEnv returns the root environment of the given environment.
func RootEnv(env Environment) Environment {
	currEnv := env
	for {
		if currEnv.IsRoot() {
			return currEnv
		}
		currEnv = currEnv.Parent()
	}
}

// Resolve a symbol is an environment and all of its parent environment.
func Resolve(env Environment, sym *sx.Symbol) (sx.Object, bool) {
	currEnv := env
	for {
		obj, found := currEnv.Lookup(sym)
		if found {
			return obj, true
		}
		if currEnv.IsRoot() {
			return sx.Nil(), false
		}
		currEnv = currEnv.Parent()
	}
}

// IsConstBinding returns true if the symbol is defined with a constant
// binding in the given environment or its parent environments.
func IsConstantBinding(env Environment, sym *sx.Symbol) bool {
	currEnv := env
	for !sx.IsNil(currEnv) {
		if currEnv.IsConst(sym) {
			return true
		}
		currEnv = currEnv.Parent()
	}
	return false
}

// AllBindings returns an a-list of all bindings in the given environment and its parent environments.
func AllBindings(env Environment) *sx.Pair {
	currEnv := env
	result := currEnv.Bindings()
	currResult := result
	if currResult != nil {
		for currResult.Tail() != nil {
			currResult = currResult.Tail()
		}
	}
	for {
		if currEnv.IsRoot() {
			return result
		}
		currEnv = currEnv.Parent()
		if currEnv == nil {
			return result
		}
		res := currEnv.Bindings()
		if result == nil {
			result = res
			currResult = result
			if currResult != nil {
				for currResult.Tail() != nil {
					currResult = currResult.Tail()
				}
			}
		} else {
			currResult = currResult.ExtendBang(res)
		}
	}
}
