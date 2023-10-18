//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxbuiltins

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxeval"
)

// Contains all syntaxes for generic condition handling.

// CondS parses a cond form: `(cond CLAUSE ...)`, where CLAUSE is
// `(EXPR EXPR ...)`.
var CondS = sxeval.Syntax{
	Name: "cond",
	Fn: func(pf *sxeval.ParseFrame, args *sx.Pair) (sxeval.Expr, error) {
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
			seq, err := ParseExprSeq(pf, clause)
			if err != nil {
				return nil, err
			}
			if l := len(seq); l > 1 {
				var front []sxeval.Expr
				if l == 2 {
					front = nil
				} else {
					front = make([]sxeval.Expr, l-2)
					copy(front, seq[1:l-1])
				}
				caseList = append(caseList, CondCase{Test: seq[0], Front: front, Last: seq[l-1]})
			}

			obj = node.Cdr()
			if sx.IsNil(obj) {
				break
			}
			rest, isPair := sx.GetPair(obj)
			if !isPair {
				return nil, fmt.Errorf("improper clause list: %v", args)
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
	Test  sxeval.Expr
	Front []sxeval.Expr
	Last  sxeval.Expr
}

func (ce *CondExpr) Rework(rf *sxeval.ReworkFrame) sxeval.Expr {
	missing := 0
	for i, cas := range ce.Cases {
		for j, expr := range cas.Front {
			cas.Front[j] = expr.Rework(rf)
		}
		ce.Cases[i].Last = cas.Last.Rework(rf)
		test := cas.Test.Rework(rf)
		if objExpr, isObjExpr := test.(sxeval.ObjectExpr); isObjExpr {
			if sx.IsTrue(objExpr.Object()) {
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
			if !sxeval.NilExpr.IsEqual(cas.Test) {
				newCases[j] = cas
				j++
			}
		}
		ce.Cases = newCases
	}
	if len(ce.Cases) == 0 {
		return sxeval.NilExpr.Rework(rf)
	}
	return ce
}

func (ce *CondExpr) Compute(frame *sxeval.Frame) (sx.Object, error) {
	subFrame := frame.MakeCalleeFrame()
	for _, cas := range ce.Cases {
		found := cas.Test == nil
		if !found {
			test, err := subFrame.Execute(cas.Test)
			if err != nil {
				return nil, err
			}
			found = sx.IsTrue(test)
		}
		if found {
			for _, expr := range cas.Front {
				_, err := subFrame.Execute(expr)
				if err != nil {
					return nil, err
				}
			}
			return frame.ExecuteTCO(cas.Last)
		}
	}
	return sx.Nil(), nil
}
func (ce *CondExpr) IsEqual(other sxeval.Expr) bool {
	if ce == other {
		return true
	}
	if otherC, ok := other.(*CondExpr); ok && otherC != nil {
		if c1, c2 := ce.Cases, otherC.Cases; len(c1) == len(c2) {
			for i, cas1 := range c1 {
				cas2 := c2[i]
				if !cas1.Test.IsEqual(cas2.Test) || !cas1.Last.IsEqual(cas2.Last) ||
					sxeval.EqualExprSlice(cas1.Front, cas2.Front) {
					return false
				}
			}
			return true
		}
	}
	return false
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
		l, err = sxeval.PrintExprs(w, cas.Front)
		length += l
		if err != nil {
			return length, err
		}
		l, err = io.WriteString(w, " ")
		length += l
		if err != nil {
			return length, err
		}
		l, err = cas.Last.Print(w)
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
