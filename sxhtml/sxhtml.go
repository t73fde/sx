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

// Package sxhtml represents HTML as s-expressions.
package sxhtml

import (
	"io"
	"sort"
	"strings"

	"zettelstore.de/sx.fossil"
)

type attrType int

const (
	_         attrType = iota
	attrPlain          // No further escape needed
	attrURL            // Escape URL
	attrCSS            // Special CSS escaping
	attrJS             // Escape JavaScript
)

// Names for special symbols.
const (
	nameAttr          = "@"
	nameCDATA         = "@C"
	nameNoEscape      = "@H"
	nameListSplice    = "@L"
	nameInlineComment = "@@"
	nameBlockComment  = "@@@"
	nameDoctype       = "@@@@"
)

var (
	SymAttr          = sx.MakeSymbol(nameAttr)
	SymCDATA         = sx.MakeSymbol(nameCDATA)
	SymNoEscape      = sx.MakeSymbol(nameNoEscape)
	SymListSplice    = sx.MakeSymbol(nameListSplice)
	SymInlineComment = sx.MakeSymbol(nameInlineComment)
	SymBlockComment  = sx.MakeSymbol(nameBlockComment)
	SymDoctype       = sx.MakeSymbol(nameDoctype)
)

// Generator is the object that allows to generate HTML.
type Generator struct {
	withNewline bool
}

// Option allows to customize the generator.
type Option func(*Generator)

// WithNewline will add new-line characters before certain tags.
func WithNewline(gen *Generator) { gen.withNewline = true }

// NewGenerator creates a new generator.
func NewGenerator(opts ...Option) *Generator {
	gen := Generator{}
	for _, opt := range opts {
		opt(&gen)
	}
	return &gen
}

// Special elements / attributes
var (
	// Void elements: https://html.spec.whatwg.org/multipage/syntax.html#void-elements
	voidTags = map[string]struct{}{
		"area": {}, "base": {}, "br": {}, "col": {}, "embed": {}, "hr": {},
		"img": {}, "input": {}, "link": {}, "meta": {}, "source": {},
		"track": {}, "wbr": {},
	}
	// Attributes with URL values: https://html.spec.whatwg.org/multipage/indices.html#attributes-1
	urlAttrs = map[string]struct{}{
		"action": {}, "cite": {}, "data": {}, "formaction": {}, "href": {},
		"itemid": {}, "itemprop": {}, "itemtype": {}, "ping": {},
		"poster": {}, "src": {},
	}
	allNLTags = map[string]struct{}{
		"head": {}, "link": {}, "meta": {}, "title": {}, "div": {},
	}
	nlTags = map[string]struct{}{
		nameCDATA: {},
		"head":    {}, "link": {}, "meta": {}, "title": {}, "script": {}, "body": {},
		"article": {}, "details": {}, "div": {}, "header": {}, "footer": {}, "form": {},
		"main": {}, "summary": {},
		"h1": {}, "h2": {}, "h3": {}, "h4": {}, "h5": {}, "h6": {},
		"li": {}, "ol": {}, "ul": {}, "dd": {}, "dt": {}, "dl": {},
		"table": {}, "thead": {}, "tbody": {}, "tr": {},
		"section": {}, "input": {},
	}
	// Elements that may be ignored if empty.
	ignoreEmptyTags = map[string]struct{}{
		"div": {}, "span": {}, "code": {}, "kbd": {}, "p": {}, "samp": {},
	}
)

// WriteHTML emit HTML code for the s-expression to the given writer.
func (gen *Generator) WriteHTML(w io.Writer, obj sx.Object) (int, error) {
	enc := myEncoder{gen: gen, pr: printer{w: w}, lastWasTag: true}
	enc.generate(obj)
	return enc.pr.length, enc.pr.err
}

// WriteListHTML emits HTML code for a list of s-expressions to the given writer.
func (gen *Generator) WriteListHTML(w io.Writer, lst *sx.Pair) (int, error) {
	enc := myEncoder{gen: gen, pr: printer{w: w}, lastWasTag: true}
	for elem := lst; elem != nil; elem = elem.Tail() {
		enc.generate(elem.Car())
	}
	return enc.pr.length, enc.pr.err
}

type myEncoder struct {
	gen        *Generator
	pr         printer
	lastWasTag bool
}

func (enc *myEncoder) generate(obj sx.Object) {
	switch o := obj.(type) {
	case sx.String:
		enc.pr.printHTML(string(o))
		enc.lastWasTag = false
	case sx.Number:
		enc.pr.printHTML(string(o.String()))
		enc.lastWasTag = false
	case *sx.Pair:
		if o.IsNil() {
			enc.lastWasTag = false
			return
		}
		if sym, isSymbol := sx.GetSymbol(o.Car()); isSymbol {
			tail := o.Tail()
			if s := sym.GetValue(); s[0] == '@' {
				switch s {
				case nameCDATA:
					enc.writeCDATA(tail)
				case nameNoEscape:
					enc.writeNoEscape(tail)
				case nameInlineComment:
					enc.writeComment(tail)
				case nameBlockComment:
					enc.writeCommentML(tail)
				case nameListSplice:
					enc.generateList(tail)
				case nameDoctype:
					enc.writeDoctype(tail)
				default:
					enc.writeTag(sym, tail)
					return
				}
				enc.lastWasTag = false
				return
			}
			enc.writeTag(sym, tail)
		}
	default:
		enc.lastWasTag = false
	}
}

func (enc *myEncoder) generateList(lst *sx.Pair) {
	for n := lst; n != nil; n = n.Tail() {
		enc.generate(n.Car())
	}
}

func (enc *myEncoder) writeCDATA(elems *sx.Pair) {
	enc.pr.printString("<![CDATA[")
	enc.writeNoEscape(elems)
	enc.pr.printString("]]>")
}

func (enc *myEncoder) writeNoEscape(elems *sx.Pair) {
	for n := elems; n != nil; n = n.Tail() {
		if s, isString := sx.GetString(n.Car()); isString {
			enc.pr.printString(string(s))
		}
	}
}

func (enc *myEncoder) writeComment(elems *sx.Pair) {
	enc.pr.printString("<!--")
	for n := elems; n != nil; n = n.Tail() {
		enc.pr.printString(" ")
		enc.printCommentObj(n.Car())
	}
	enc.pr.printString(" -->")
}
func (enc *myEncoder) writeCommentML(elems *sx.Pair) {
	enc.pr.printString("<!--")
	for n := elems; n != nil; n = n.Tail() {
		enc.pr.printString("\n")
		enc.printCommentObj(n.Car())
	}
	enc.pr.printString("\n-->\n")
}
func (enc *myEncoder) printCommentObj(obj sx.Object) {
	enc.pr.printComment(obj.GoString())
}

func (enc *myEncoder) writeDoctype(elems *sx.Pair) {
	// TODO: check for multiple doctypes, error on second
	enc.pr.printString("<!DOCTYPE html>\n")
	enc.generateList(elems)
}

func (enc *myEncoder) writeTag(sym *sx.Symbol, elems *sx.Pair) {
	symVal := sym.GetValue()
	if _, ignoreEmptyTag := ignoreEmptyTags[symVal]; ignoreEmptyTag && ignoreEmptyStrings(elems) == nil {
		return
	}
	_, isNLTag := nlTags[symVal]
	withNewline := enc.gen.withNewline && isNLTag
	tagName := sym.String()
	if _, isAllNLTags := allNLTags[symVal]; withNewline && (!enc.lastWasTag || isAllNLTags) {
		enc.pr.printStrings("\n<", tagName)
	} else {
		enc.pr.printStrings("<", tagName)
	}
	if attrs := enc.getAttributes(elems); attrs != nil {
		enc.writeAttributes(attrs)
		elems = elems.Tail()
	}
	enc.pr.printString(">")
	if _, isVoid := voidTags[symVal]; isVoid {
		enc.lastWasTag = withNewline
		return
	}

	enc.generateList(elems)
	if withNewline {
		enc.pr.printStrings("</", tagName, ">\n")
	} else {
		enc.pr.printStrings("</", tagName, ">")
	}
	enc.lastWasTag = withNewline
}

func ignoreEmptyStrings(elem *sx.Pair) *sx.Pair {
	for node := elem; node != nil; node = node.Tail() {
		if s, isString := sx.GetString(node.Car()); !isString || s != "" {
			return node
		}
	}
	return nil
}

func (enc *myEncoder) getAttributes(lst *sx.Pair) *sx.Pair {
	pair, isPair := sx.GetPair(lst.Car())
	if !isPair || pair == nil {
		return nil
	}
	sym, isSymbol := sx.GetSymbol(pair.Car())
	if !isSymbol || !sym.IsEqual(SymAttr) {
		return nil
	}
	return pair.Tail()
}

func (enc *myEncoder) writeAttributes(attrs *sx.Pair) {
	length := attrs.Length()
	found := make(map[string]struct{}, length)
	empty := make(map[string]struct{}, length)
	a := make(map[string]string, length)
	for node := attrs; node != nil; node = node.Tail() {
		pair, isPair := sx.GetPair(node.Car())
		if !isPair {
			continue
		}
		sym, isSymbol := sx.GetSymbol(pair.Car())
		if !isSymbol {
			continue
		}
		key := sym.String()
		if _, found := found[key]; found {
			continue
		}
		found[key] = struct{}{}
		if cdr := pair.Cdr(); !sx.IsNil(cdr) {
			var obj sx.Object
			if tail, isTail := sx.GetPair(cdr); isTail {
				obj = tail.Car()
			} else {
				obj = cdr
			}
			var s string
			switch o := obj.(type) {
			case sx.String:
				s = string(o)
			case *sx.Symbol:
				s = o.GetValue()
			case sx.Number:
				s = o.GoString()
			default:
				continue
			}
			a[key] = strings.TrimSpace(getAttributeValue(sym, s))
		} else {
			a[key] = ""
			empty[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(a))
	for key := range a {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		enc.pr.printStrings(" ", key)
		if _, isEmpty := empty[key]; !isEmpty {
			enc.pr.printString(`="`)
			enc.pr.printAttributeValue(a[key])
			enc.pr.printString(`"`)
		}
	}
}

func getAttributeValue(sym *sx.Symbol, value string) string {
	switch getAttributeType(sym) {
	case attrURL:
		return urlEscape(value)
	default:
		return value
	}
}

func getAttributeType(sym *sx.Symbol) attrType {
	name := sym.String()
	if dataName, isData := strings.CutPrefix(name, "data-"); isData {
		name = dataName
		sym = sx.MakeSymbol(name)
	} else if prefix, rest, hasPrefix := strings.Cut(name, ":"); hasPrefix {
		if prefix == "xmlns" {
			return attrURL
		}
		name = rest
		sym = sx.MakeSymbol(name)
	}

	if _, isUrlAttr := urlAttrs[sym.GetValue()]; isUrlAttr {
		return attrURL
	}
	if sym.IsEqual(sx.MakeSymbol("style")) {
		return attrCSS
	}

	// Attribute names starting with "on" (e.g. "onload") are treated as JavaScript values.
	if strings.HasPrefix(name, "on") {
		return attrJS
	}

	// Names that contain something similar to URL are treated as URLs
	if strings.Contains(name, "url") || strings.Contains(name, "uri") || strings.Contains(name, "src") {
		return attrURL
	}
	return attrPlain
}
