//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxpf.
//
// sxpf is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package quote

import (
	"fmt"
	"io"

	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/builtins/list"
	"zettelstore.de/sx.fossil/sxpf/eval"
	"zettelstore.de/sx.fossil/sxpf/reader"
)

// InstallQuasiQuoteReader sets the reader macros to support quasi quotation.
func InstallQuasiQuoteReader(rd *reader.Reader, symQQ *sxpf.Symbol, chQQ rune, symUQ *sxpf.Symbol, chUQ rune, symUQS *sxpf.Symbol, chUQS rune) {
	if sf := rd.SymbolFactory(); sf != symQQ.Factory() || sf != symUQ.Factory() || sf != symUQ.Factory() {
		panic("reader symbol factory differ from factory of symbols")
	}
	rd.SetMacro(chQQ, makeQuotationMacro(symQQ))
	rd.SetMacro(chUQ, func(rd *reader.Reader, _ rune) (sxpf.Object, error) {
		ch, err := rd.NextRune()
		if err != nil {
			return nil, err
		}
		sym := symUQS
		if ch != chUQS {
			rd.Unread(ch)
			sym = symUQ
		}
		obj, err := rd.Read()
		if err == nil {
			return sxpf.Nil().Cons(obj).Cons(sym), nil
		}
		return obj, err
	})

}

func InstallQuasiQuoteSyntax(env sxpf.Environment, symQQ, symUQ, symUQS *sxpf.Symbol) error {
	err := env.Bind(symQQ, eval.MakeSyntax(symQQ.Name(), makeQuasiQuoteSyntax(symQQ, symUQ, symUQS)))
	if err != nil {
		return err
	}
	err = env.Bind(symUQ, eval.MakeSyntax(symUQ.Name(), makeUnquoteSyntax(symQQ)))
	if err != nil {
		return err
	}
	err = env.Bind(symUQS, eval.MakeSyntax(symUQS.Name(), makeUnquoteSyntax(symQQ)))
	return err
}

func makeUnquoteSyntax(symQQ *sxpf.Symbol) eval.SyntaxFn {
	return func(*eval.Engine, sxpf.Environment, *sxpf.Pair) (eval.Expr, error) {
		return nil, fmt.Errorf("not allowed outside %s", symQQ.Name())
	}
}

func makeQuasiQuoteSyntax(symQQ, symUQ, symUQS *sxpf.Symbol) eval.SyntaxFn {
	return func(eng *eval.Engine, env sxpf.Environment, args *sxpf.Pair) (eval.Expr, error) {
		if sxpf.IsNil(args) {
			return nil, eval.ErrNoArgs
		}
		if !sxpf.IsNil(args.Cdr()) {
			return nil, fmt.Errorf("more than one argument: %v", args)
		}
		qqp := qqParser{
			engine:             eng,
			env:                env,
			symQuasiQuote:      symQQ,
			symUnquote:         symUQ,
			symUnquoteSplicing: symUQS,
		}
		return qqp.parseQQ(args.Car())
	}
}

type qqParser struct {
	engine             *eval.Engine
	env                sxpf.Environment
	symQuasiQuote      *sxpf.Symbol
	symUnquote         *sxpf.Symbol
	symUnquoteSplicing *sxpf.Symbol
}

func (qqp *qqParser) parse(obj sxpf.Object) (eval.Expr, error) {
	return qqp.engine.Parse(qqp.env, obj)
}

func (qqp *qqParser) parseQQ(obj sxpf.Object) (eval.Expr, error) {
	pair, isPair := sxpf.GetPair(obj)
	if !isPair || pair == nil {
		// `basic is the same as (quote basic), for any form basic that is not a list.
		return eval.ObjExpr{Obj: obj}, nil
	}
	first := pair.Car()
	if sym, isSymbol := sxpf.GetSymbol(first); isSymbol {
		if qqp.symUnquote.IsEqual(sym) {
			form, err := getUnquoteObj(sym, pair)
			if err != nil {
				return nil, err
			}
			// `,form is the same as form, for any form.
			return qqp.parse(form)
		}
		if qqp.symQuasiQuote.IsEqual(sym) {
			form, err := getUnquoteObj(sym, pair)
			if err != nil {
				return nil, err
			}
			// If the backquote syntax is nested, the innermost backquoted form should be expanded first.
			// This means that if several commas occur in a row, the leftmost one belongs to the innermost backquote.
			expr, err := qqp.parseQQ(form)
			if err != nil {
				return nil, err
			}
			return listArgs([]eval.Expr{eval.ObjExpr{Obj: sym}, expr}), err
		}
		if qqp.symUnquoteSplicing.IsEqual(sym) {
			// `,@form has undefined consequences.
			return nil, fmt.Errorf("(%v %v) is not allowed", qqp.symQuasiQuote.Name(), obj)
		}
	}
	args, err := qqp.parseList(pair)
	if err != nil {
		return nil, err
	}
	return combineArgs(args), nil
}

// combineArgs optimizes some cases for (append ...).
//
// (append) --> ()
// (append x) --> x
// (append (x) (y) ...) --> (list x y ...) OR (list* x y ...), if all but the last element
// are not spliced. (list* ...) will be used, if the last element was spliced, (list ...) if not.
//
// In addition, for some sequences in (append ... (x) (y) ...), these will be simplified
// into (append ... (list x y) ...).
func combineArgs(args []eval.Expr) eval.Expr {
	if len(args) == 0 {
		return eval.NilExpr
	}
	if len(args) == 1 {
		return args[0]
	}
	mleCount := countMLE(args)
	if mleCount < len(args)-1 {
		newArgs := collectMakeList(args)
		return &eval.BuiltinCallExpr{Proc: eval.BuiltinA(list.Append), Args: newArgs}
	}
	newArgs := make([]eval.Expr, len(args))
	for i := 0; i < mleCount; i++ {
		newArgs[i] = args[i].(MakeListExpr).Elem
	}
	if mleCount < len(args) {
		newArgs[mleCount] = args[mleCount]
		return listStarArgs(newArgs)
	}
	return listArgs(newArgs)
}
func countMLE(args []eval.Expr) int {
	for i := 0; i < len(args); i++ {
		if _, isMLE := args[i].(MakeListExpr); !isMLE {
			return i
		}
	}
	return len(args)
}
func collectMakeList(args []eval.Expr) []eval.Expr {
	result := make([]eval.Expr, 0, len(args))
	var makeLists []eval.Expr
	for _, arg := range args {
		if mle, isMLE := arg.(MakeListExpr); isMLE {
			makeLists = append(makeLists, mle.Elem)
		} else {
			if len(makeLists) > 0 {
				result = append(result, listArgs(makeLists))
			}
			makeLists = nil
			result = append(result, arg)
		}
	}
	if len(makeLists) > 0 {
		result = append(result, listArgs(makeLists))
	}
	return result
}

// listArgs optimizes some cases for (list ...).
//
// (list)           --> ()
// (list 'x 'y ...) --> '(x y ...)
//
// In addition, arguments of the form (list 'x) are transformed into '(x) before optimization.
func listArgs(args []eval.Expr) eval.Expr {
	if len(args) == 0 {
		return eval.NilExpr
	}
	if countQuote(args) < len(args) {
		lstArgs := resolveMakeListQuoted(args)
		return &eval.BuiltinCallExpr{Proc: eval.BuiltinA(list.List), Args: lstArgs}
	}
	lstArgs := make([]sxpf.Object, len(args))
	for i, arg := range args {
		if oe, isObj := arg.(eval.ObjExpr); isObj {
			lstArgs[i] = oe.Obj
		} else {
			lstArgs[i] = sxpf.MakeList(arg.(MakeListExpr).Elem.(eval.ObjExpr).Obj)
		}
	}
	return eval.ObjExpr{Obj: sxpf.MakeList(lstArgs...)}
}
func countQuote(args []eval.Expr) int {
	for i, arg := range args {
		if _, isObj := arg.(eval.ObjExpr); isObj {
			continue
		}
		if mle, isMLE := arg.(MakeListExpr); isMLE {
			if _, isObj := mle.Elem.(eval.ObjExpr); isObj {
				continue
			}
		}
		return i
	}
	return len(args)
}

// resolveMakeListQuoted changes arguments.
//
// It basically transforms a (list 'x) into '(x).
// It does not work on arbitrary (list ...)-calls, but only those with one arg.
func resolveMakeListQuoted(args []eval.Expr) []eval.Expr {
	result := make([]eval.Expr, len(args))
	for i, arg := range args {
		if mle, isMLE := arg.(MakeListExpr); isMLE {
			if oe, isObj := mle.Elem.(eval.ObjExpr); isObj {
				result[i] = eval.ObjExpr{Obj: sxpf.MakeList(oe.Obj)}
				continue
			}
		}
		result[i] = arg
	}
	return result
}

// listStarArgs optimizes some cases for (list* ...).
//
// (list*)     --> ()
// (list* x)   --> x
// (list* x y) --> (cons x y)
func listStarArgs(args []eval.Expr) eval.Expr {
	switch len(args) {
	case 0:
		return eval.NilExpr
	case 1:
		lstArgs := resolveMakeListQuoted(args)
		return lstArgs[0]
	case 2:
		lstArgs := resolveMakeListQuoted(args)
		return &eval.BuiltinCallExpr{Proc: eval.BuiltinA(list.Cons), Args: lstArgs}
	default:
		lstArgs := resolveMakeListQuoted(args)
		return &eval.BuiltinCallExpr{Proc: eval.BuiltinA(list.ListStar), Args: lstArgs}
	}
}

func (qqp *qqParser) parseList(lst *sxpf.Pair) ([]eval.Expr, error) {
	length, prevPair, lastPair := analyseList(lst)
	if length == 0 {
		return nil, nil
	}
	numArgs, realArgs := length, length
	var form eval.Expr
	if prevPair != nil {
		if sym, isSymbol := sxpf.GetSymbol(prevPair.Car()); isSymbol {
			if qqp.symUnquote.IsEqual(sym) {
				obj, err := getUnquoteObj(sym, prevPair)
				if err != nil {
					return nil, err
				}
				// `(x1 x2 x3 ... xn . ,form) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] form)
				expr, err := qqp.parse(obj)
				if err != nil {
					return nil, err
				}
				numArgs--
				realArgs -= 2
				form = expr
			} else if qqp.symUnquoteSplicing.IsEqual(sym) {
				// `(x1 x2 x3 ... xn . ,@form) has undefined consequences.
				return nil, fmt.Errorf("%v not allowed", lst)
			}
		}
	}
	if form == nil {
		last := lastPair.Cdr()
		if !sxpf.IsNil(last) {
			// `(x1 x2 x3 ... xn . atom) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] (quote atom))
			form = eval.ObjExpr{Obj: last}
			numArgs++
		}
	}

	args := make([]eval.Expr, numArgs)
	for node, i := lst, 0; i < realArgs; i++ {
		elem := node.Car()
		node = node.Tail()
		if elemList, isPair := sxpf.GetPair(elem); isPair && elemList != nil {
			if sym, isSymbol := sxpf.GetSymbol(elemList.Car()); isSymbol {
				if qqp.symUnquote.IsEqual(sym) {
					// -- [,form] is interpreted as (list form)
					obj, err := getUnquoteObj(sym, elemList)
					if err != nil {
						return nil, err
					}
					expr, err := qqp.parse(obj)
					if err != nil {
						return nil, err
					}
					args[i] = MakeListExpr{expr}
					continue
				}
				if qqp.symUnquoteSplicing.IsEqual(sym) {
					// -- [,@form] is interpreted as form.
					obj, err := getUnquoteObj(sym, elemList)
					if err != nil {
						return nil, err
					}
					expr, err := qqp.parse(obj)
					if err != nil {
						return nil, err
					}
					args[i] = expr
					continue
				}
			}
		}
		// -- [form] is interpreted as (list `form), which contains a backquoted form that must then be further interpreted.
		expr, err := qqp.parseQQ(elem)
		if err != nil {
			return nil, err
		}
		args[i] = MakeListExpr{expr}
	}

	if form != nil {
		// `(x1 x2 x3 ... xn . ,form) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] form)
		// or
		// `(x1 x2 x3 ... xn . atom) may be interpreted to mean (append [ x1] [ x2] [ x3] ... [ xn] (quote atom))
		args[realArgs] = form
	}
	return args, nil
}

func analyseList(lst *sxpf.Pair) (int, *sxpf.Pair, *sxpf.Pair) {
	length := 0
	prevObj, lastPair := sxpf.Nil(), sxpf.Nil()
	for node := lst; node != nil; {
		length++
		prevObj = lastPair
		lastPair = node
		next, isPair := sxpf.GetPair(node.Cdr())
		if !isPair {
			break
		}
		node = next
	}
	return length, prevObj, lastPair
}

func getUnquoteObj(sym *sxpf.Symbol, lst *sxpf.Pair) (sxpf.Object, error) {
	args, isPair := sxpf.GetPair(lst.Cdr())
	if !isPair {
		return nil, sxpf.ErrImproper{Pair: lst}
	}
	if args == nil {
		return nil, fmt.Errorf("missing argument for %s", sym.Name())
	}
	obj := args.Car()
	rest := args.Cdr()
	if sxpf.IsNil(rest) {
		return obj, nil
	}
	return nil, fmt.Errorf("additional arguments %v for %s", rest.Repr(), sym.Name())
}

type MakeListExpr struct{ Elem eval.Expr }

func (mle MakeListExpr) Compute(eng *eval.Engine, env sxpf.Environment) (sxpf.Object, error) {
	elem, err := eng.Execute(env, mle.Elem)
	if err != nil {
		return nil, err
	}
	return sxpf.Cons(elem, nil), nil
}
func (mle MakeListExpr) Print(w io.Writer) (int, error) {
	length, err := io.WriteString(w, "{MAKELIST ")
	if err != nil {
		return length, err
	}
	l, err := mle.Elem.Print(w)
	length += l
	if err != nil {
		return length, err
	}
	l, err = io.WriteString(w, "}")
	length += l
	return length, err

}
func (mle MakeListExpr) Rework(ro *eval.ReworkOptions, env sxpf.Environment) eval.Expr {
	mle.Elem = mle.Elem.Rework(ro, env)
	return mle
}
