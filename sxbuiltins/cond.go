//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Contains all syntaxes for generic condition handling.
const condName = "cond"

// CondS parses a cond form: `(cond CLAUSE ...)`, where CLAUSE is
// `(TEST EXPR ...)`.
var CondS = sxeval.Special{
	Name: condName,
	Fn: func(pf *sxeval.ParseEnvironment, args *sx.Pair) (sxeval.Expr, error) {
		if args == nil {
			return sxeval.NilExpr, nil
		}
		caseList := []CondCase{}
		for node := args; ; {
			obj := node.Car()
			clause, isPair := sx.GetPair(obj)
			if !isPair {
				return nil, fmt.Errorf("clause must be a list, but got %T/%v", obj, obj)
			}
			if !clause.IsNil() {
				test, err := pf.Parse(clause.Car())
				if err != nil {
					return nil, err
				}
				cdr := clause.Cdr()
				if cdr.IsNil() {
					// Only a test. If true returns ().
					caseList = append(caseList, CondCase{test, sxeval.NilExpr})
				} else {
					var expr sxeval.Expr
					if seq, isSeq := sx.GetPair(cdr); !isSeq {
						expr, err = pf.Parse(seq)
					} else {
						expr, err = ParseExprSeq(pf, seq)
					}
					if err != nil {
						return nil, err
					}
					caseList = append(caseList, CondCase{test, expr})
				}
			}

			obj = node.Cdr()
			if sx.IsNil(obj) {
				break
			}
			rest, isPair := sx.GetPair(obj)
			if !isPair {
				return nil, sx.ErrImproper{Pair: args}
			}
			node = rest
		}
		if len(caseList) == 0 {
			return sxeval.NilExpr, nil
		}
		return &CondExpr{Cases: caseList}, nil
	},
}

type CondExpr struct {
	Cases []CondCase
}
type CondCase struct {
	Test sxeval.Expr
	Expr sxeval.Expr
}

func (ce *CondExpr) Unparse() sx.Object {
	var obj sx.Object
	for i := len(ce.Cases) - 1; i >= 0; i-- {
		obj = sx.Cons(sx.MakeList(ce.Cases[i].Test.Unparse(), ce.Cases[i].Expr.Unparse()), obj)
	}
	return sx.Cons(sx.MakeSymbol(condName), obj)
}

func (ce *CondExpr) Rework(re *sxeval.ReworkEnvironment) sxeval.Expr {
	missing := 0
	for i, cas := range ce.Cases {
		ce.Cases[i].Expr = cas.Expr.Rework(re)
		test := cas.Test.Rework(re)
		if objExpr, isConstObject := test.(sxeval.ConstObjectExpr); isConstObject {
			if sx.IsTrue(objExpr.ConstObject()) {
				if i == missing {
					// Only False tests in front, this is definitely the first one that is True
					return ce.Cases[i].Expr
				}
				ce.Cases[i].Test = nil
				newCases := make([]CondCase, i+1)
				for j := 0; j <= i; j++ {
					newCases[j] = ce.Cases[j]
				}
				ce.Cases = newCases
				return ce
			}
			ce.Cases[i].Test = sxeval.NilExpr
			missing++
		}
	}
	if missing != 0 {
		newCases := make([]CondCase, len(ce.Cases)-missing)
		j := 0
		for _, cas := range ce.Cases {
			if cas.Test != sxeval.NilExpr {
				newCases[j] = cas
				j++
			}
		}
		ce.Cases = newCases
	}
	if len(ce.Cases) == 0 {
		return sxeval.NilExpr.Rework(re)
	}
	return ce
}

func (ce *CondExpr) Compute(env *sxeval.Environment) (sx.Object, error) {
	subEnv := env.NewDynamicEnvironment()
	for _, cas := range ce.Cases {
		found := cas.Test == nil
		if !found {
			test, err := subEnv.Execute(cas.Test)
			if err != nil {
				return nil, err
			}
			found = sx.IsTrue(test)
		}
		if found {
			return env.ExecuteTCO(cas.Expr)
		}
	}
	return sx.Nil(), nil
}
func (ce *CondExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{COND")
	if err != nil {
		return length, err
	}
	var l int
	for _, cas := range ce.Cases {
		l, err = io.WriteString(w, " [")
		length += l
		if err != nil {
			return length, err
		}
		if test := cas.Test; test == nil {
			l, err = io.WriteString(w, "T")
		} else {
			l, err = cas.Test.Print(w)
		}
		length += l
		if err != nil {
			return length, err
		}
		l, err = io.WriteString(w, " ")
		length += l
		if err != nil {
			return length, err
		}
		l, err = cas.Expr.Print(w)
		length += l
		if err != nil {
			return length, err
		}
		l, err = io.WriteString(w, "]")
		length += l
		if err != nil {
			return length, err
		}
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err
}
