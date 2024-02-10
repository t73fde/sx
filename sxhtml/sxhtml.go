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
var (
	SymAttr          = sx.MakeSymbol("@")
	SymCDATA         = sx.MakeSymbol("@C")
	SymNoEscape      = sx.MakeSymbol("@H")
	SymListSplice    = sx.MakeSymbol("@L")
	SymInlineComment = sx.MakeSymbol("@@")
	SymBlockComment  = sx.MakeSymbol("@@@")
	SymDoctype       = sx.MakeSymbol("@@@@")
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
	voidTags = map[sx.Symbol]bool{
		sx.MakeSymbol("area"): true, sx.MakeSymbol("base"): true, sx.MakeSymbol("br"): true,
		sx.MakeSymbol("col"): true, sx.MakeSymbol("embed"): true, sx.MakeSymbol("hr"): true,
		sx.MakeSymbol("img"): true, sx.MakeSymbol("input"): true, sx.MakeSymbol("link"): true,
		sx.MakeSymbol("meta"): true, sx.MakeSymbol("source"): true,
		sx.MakeSymbol("track"): true, sx.MakeSymbol("wbr"): true,
	}
	// Attributes with URL values: https://html.spec.whatwg.org/multipage/indices.html#attributes-1
	urlAttrs = map[sx.Symbol]bool{
		sx.MakeSymbol("action"): true, sx.MakeSymbol("cite"): true, sx.MakeSymbol("data"): true,
		sx.MakeSymbol("formaction"): true, sx.MakeSymbol("href"): true,
		sx.MakeSymbol("itemid"): true, sx.MakeSymbol("itemprop"): true,
		sx.MakeSymbol("itemtype"): true, sx.MakeSymbol("ping"): true,
		sx.MakeSymbol("poster"): true, sx.MakeSymbol("src"): true,
	}
	allNLTags = map[sx.Symbol]bool{
		sx.MakeSymbol("head"): true, sx.MakeSymbol("link"): true, sx.MakeSymbol("meta"): true,
		sx.MakeSymbol("title"): true, sx.MakeSymbol("div"): true,
	}
	nlTags = map[sx.Symbol]bool{
		SymCDATA:              true,
		sx.MakeSymbol("head"): true, sx.MakeSymbol("link"): true, sx.MakeSymbol("meta"): true,
		sx.MakeSymbol("title"): true, sx.MakeSymbol("script"): true, sx.MakeSymbol("body"): true,
		sx.MakeSymbol("article"): true, sx.MakeSymbol("details"): true, sx.MakeSymbol("div"): true,
		sx.MakeSymbol("header"): true, sx.MakeSymbol("footer"): true, sx.MakeSymbol("form"): true,
		sx.MakeSymbol("main"): true, sx.MakeSymbol("summary"): true,
		sx.MakeSymbol("h1"): true, sx.MakeSymbol("h2"): true, sx.MakeSymbol("h3"): true,
		sx.MakeSymbol("h4"): true, sx.MakeSymbol("h5"): true, sx.MakeSymbol("h6"): true,
		sx.MakeSymbol("li"): true, sx.MakeSymbol("ol"): true, sx.MakeSymbol("ul"): true,
		sx.MakeSymbol("dd"): true, sx.MakeSymbol("dt"): true, sx.MakeSymbol("dl"): true,
		sx.MakeSymbol("table"): true, sx.MakeSymbol("thead"): true, sx.MakeSymbol("tbody"): true,
		sx.MakeSymbol("tr"): true, sx.MakeSymbol("section"): true, sx.MakeSymbol("input"): true,
	}
	// Elements that may be ignored if empty.
	emptyTags = map[sx.Symbol]bool{
		sx.MakeSymbol("div"): true, sx.MakeSymbol("span"): true, sx.MakeSymbol("code"): true,
		sx.MakeSymbol("kbd"): true, sx.MakeSymbol("p"): true, sx.MakeSymbol("samp"): true,
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
			if s := sym.GoString(); s[0] == '@' {
				switch sym {
				case SymCDATA:
					enc.writeCDATA(tail)
				case SymNoEscape:
					enc.writeNoEscape(tail)
				case SymInlineComment:
					enc.writeComment(tail)
				case SymBlockComment:
					enc.writeCommentML(tail)
				case SymListSplice:
					enc.generateList(tail)
				case SymDoctype:
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

func (enc *myEncoder) writeTag(sym sx.Symbol, elems *sx.Pair) {
	if emptyTags[sym] && ignoreEmptyStrings(elems) == nil {
		return
	}
	withNewline := enc.gen.withNewline && nlTags[sym]
	tagName := sym.String()
	if withNewline && (!enc.lastWasTag || allNLTags[sym]) {
		enc.pr.printStrings("\n<", tagName)
	} else {
		enc.pr.printStrings("<", tagName)
	}
	if attrs := enc.getAttributes(elems); attrs != nil {
		enc.writeAttributes(attrs)
		elems = elems.Tail()
	}
	enc.pr.printString(">")
	if voidTags[sym] {
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
			case sx.Symbol:
				s = o.GoString()
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

func getAttributeValue(sym sx.Symbol, value string) string {
	switch getAttributeType(sym) {
	case attrURL:
		return urlEscape(value)
	default:
		return value
	}
}

func getAttributeType(sym sx.Symbol) attrType {
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

	if urlAttrs[sym] {
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
