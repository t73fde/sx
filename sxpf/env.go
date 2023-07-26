//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxpf

import (
	"fmt"
	"io"
	"strconv"
	"sync"
)

// Environment maintains a mapping between symbols and values.
// Form are evaluated within environments.
type Environment interface {
	// An environment is an object by itself
	Object

	// String returns the local name of this environment.
	String() string

	// Parent allows to retrieve the parent environment. Environment is the root
	// environment, nil is returned. Lookups that cannot be satisfied in an
	// environment are delegated to the parent envrionment.
	Parent() Environment

	// Bind creates a local mapping with a given symbol and object.
	// A previous mapping will be overwritten.
	Bind(*Symbol, Object) error

	// Lookup will search for a local binding of the given symbol. If not
	// found, the search will *not* be continued in the parent environment.
	// Use the global `Resolve` function, if you want a search up to the parent.
	Lookup(*Symbol) (Object, bool)

	// Bindings returns all bindings as an a-list in some random order.
	Bindings() *Pair

	// Unbind removes the mapping of the given symbol to an object.
	Unbind(*Symbol) error

	// Freeze sets the environment in a read-only state.
	Freeze()
}

// ErrEnvFrozen is returned when trying to update a frozen environment.
type ErrEnvFrozen struct{ Env Environment }

func (err ErrEnvFrozen) Error() string { return fmt.Sprintf("enviroment is frozen: %v", err.Env) }

type mapSymObj map[*Symbol]Object

func (mso mapSymObj) isEqual(other mapSymObj) bool {
	if len(mso) != len(other) {
		return false
	}
	for k, v := range mso {
		ov, found := other[k]
		if !found || !v.IsEqual(ov) {
			return false
		}
	}
	return true
}

// GetEnvironment returns the object as an environment, if possible.
func GetEnvironment(obj Object) (Environment, bool) {
	if IsNil(obj) {
		return nil, false
	}
	env, ok := obj.(Environment)
	return env, ok
}

func (mso mapSymObj) asAlist() *Pair {
	result := Nil()
	for k, v := range mso {
		result = result.Cons(Cons(k, v))
	}
	return result
}

// MakeRootEnvironment creates a new root environment.
func MakeRootEnvironment() Environment {
	return &rootEnvironment{
		vars:   make(mapSymObj, RootEnvironmentSize),
		frozen: false,
	}
}

// RootEnvironmentSize is the base size of the root environment.
// If more bindings are entered, it must be re-sized, which may consume some time.
const RootEnvironmentSize = 128

type rootEnvironment struct {
	mu     sync.RWMutex
	vars   mapSymObj
	frozen bool
}

func (re *rootEnvironment) IsNil() bool             { return re == nil }
func (re *rootEnvironment) IsAtom() bool            { return re == nil }
func (re *rootEnvironment) IsEql(other Object) bool { return re == other }
func (re *rootEnvironment) IsEqual(other Object) bool {
	if re == other {
		return true
	}
	if re.IsNil() {
		return IsNil(other)
	}
	ore, ok := other.(*rootEnvironment)
	if !ok {
		return false
	}
	re.mu.RLock()
	ore.mu.RLock()
	result := re.vars.isEqual(ore.vars)
	ore.mu.RUnlock()
	re.mu.RUnlock()
	return result
}
func (re *rootEnvironment) Repr() string   { return Repr(re) }
func (re *rootEnvironment) String() string { return "<root>" }
func (re *rootEnvironment) Print(w io.Writer) (int, error) {
	re.mu.RLock()
	length, err := WriteStrings(w, "#<env:", re.String(), "/", strconv.Itoa(len(re.vars)), ">")
	re.mu.RUnlock()
	return length, err
}
func (re *rootEnvironment) Parent() Environment { return nil }
func (re *rootEnvironment) Bind(sym *Symbol, obj Object) error {
	re.mu.Lock()
	var err error
	if re.frozen {
		err = ErrEnvFrozen{Env: re}
	} else {
		re.vars[sym] = obj
	}
	re.mu.Unlock()
	return err
}
func (re *rootEnvironment) Lookup(sym *Symbol) (Object, bool) {
	re.mu.RLock()
	obj, found := re.vars[sym]
	re.mu.RUnlock()
	return obj, found
}
func (re *rootEnvironment) Bindings() *Pair {
	re.mu.RLock()
	al := re.vars.asAlist()
	re.mu.RUnlock()
	return al
}
func (re *rootEnvironment) Unbind(sym *Symbol) error {
	re.mu.Lock()
	var err error
	if re.frozen {
		err = ErrEnvFrozen{Env: re}
	} else {
		delete(re.vars, sym)
	}
	re.mu.Unlock()
	return err
}
func (re *rootEnvironment) Freeze() { re.frozen = true }

// MakeChildEnvironment creates a new environment with a given parent.
func MakeChildEnvironment(parent Environment, name string, baseSize int) Environment {
	if baseSize <= 0 {
		baseSize = 8
	}
	return &childEnvironment{
		name:   name,
		parent: parent,
		vars:   make(mapSymObj, baseSize),
		frozen: false,
	}
}

type childEnvironment struct {
	name   string
	parent Environment
	vars   mapSymObj
	frozen bool
}

func (ce *childEnvironment) IsNil() bool             { return ce == nil }
func (ce *childEnvironment) IsAtom() bool            { return ce == nil }
func (ce *childEnvironment) IsEql(other Object) bool { return ce == other }
func (ce *childEnvironment) IsEqual(other Object) bool {
	if ce == other {
		return true
	}
	if ce.IsNil() {
		return IsNil(other)
	}
	oce, ok := other.(*childEnvironment)
	if !ok {
		return false
	}
	return ce.vars.isEqual(oce.vars)
}
func (ce *childEnvironment) Repr() string { return Repr(ce) }
func (ce *childEnvironment) Print(w io.Writer) (int, error) {
	return WriteStrings(w, "#<env:", ce.name, "/", strconv.Itoa(len(ce.vars)), ">")
}
func (ce *childEnvironment) String() string      { return ce.name }
func (ce *childEnvironment) Parent() Environment { return ce.parent }
func (ce *childEnvironment) Bind(sym *Symbol, val Object) error {
	if ce.frozen {
		return ErrEnvFrozen{Env: ce}
	}
	ce.vars[sym] = val
	return nil
}
func (ce *childEnvironment) Lookup(sym *Symbol) (Object, bool) {
	obj, found := ce.vars[sym]
	return obj, found
}
func (ce *childEnvironment) Bindings() *Pair { return ce.vars.asAlist() }
func (ce *childEnvironment) Unbind(sym *Symbol) error {
	if ce.frozen {
		return ErrEnvFrozen{Env: ce}
	}
	delete(ce.vars, sym)
	return nil
}
func (ce *childEnvironment) Freeze() { ce.frozen = true }

// RootEnv returns the root environment of the given environment.
func RootEnv(env Environment) Environment {
	currEnv := env
	for {
		if _, found := (currEnv).(*rootEnvironment); found {
			return currEnv
		}
	}
}

// Resolve a symbol is an environment and all of its parent environment.
func Resolve(env Environment, sym *Symbol) (Object, bool) {
	currEnv := env
	for {
		obj, found := currEnv.Lookup(sym)
		if found {
			return obj, true
		}
		if _, ok := currEnv.(*rootEnvironment); ok {
			return Nil(), false
		}
		currEnv = currEnv.Parent()
	}
}

// AllBindings returns an a-list of all bindings in the given environment and its parent environments.
func AllBindings(env Environment) *Pair {
	currEnv := env
	result := currEnv.Bindings()
	currResult := result
	if currResult != nil {
		for currResult.Tail() != nil {
			currResult = currResult.Tail()
		}
	}
	for {
		if _, ok := currEnv.(*rootEnvironment); ok {
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
