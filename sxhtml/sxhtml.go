//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sxhtml.
//
// sxhtml is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxhtml

import (
	"io"
	"sort"
	"strings"

	"zettelstore.de/sx.fossil/sxpf"
)

const (
	// keyVoid marks void HTML elements, i.e. w/o end tag.
	keyVoid = sxpf.Keyword("sxhtml:isVoid")

	// keyIgnoreEmpty marks HTML tags that may be ignored if empty.
	keyIgnoreEmpty = sxpf.Keyword("sxhtml:igEm")

	// keyWithNL marks HTML elements that may start with a new-line character
	keyWithNL = sxpf.Keyword("sxhtml:nl")

	// keyAlwaysNL signals that an new-line character should be emitted if needed.
	keyAlwaysNL = sxpf.Keyword("sxhtml:alwaysNL")

	// keyAttr specifies the type of an HTML attribute value.
	// This controls how attribute values are escaped.
	keyAttr   = sxpf.Keyword("sxhtml:attrType")
	attrPlain = sxpf.Keyword("sxhtml:plain") // No further escape needed
	attrURL   = sxpf.Keyword("sxhtml:url")   // Escape URL
	attrCSS   = sxpf.Keyword("sxhtml:css")   // Special CSS escaping
	attrJS    = sxpf.Keyword("sxhtml:js")    // Escape JavaScript
)

// Names for special symbols.
const (
	NameSymAttr          = "@"
	NameSymCDATA         = "@C"
	NameSymNoEscape      = "@H"
	NameSymInlineComment = "@@"
	NameSymBlockComment  = "@@@"
	NameSymDoctype       = "@@@@"
)

// Generator is the object that allows to generate HTML.
type Generator struct {
	sf          sxpf.SymbolFactory
	symAttr     *sxpf.Symbol
	withNewline bool
}

// Option allows to customize the generator.
type Option func(*Generator)

// WithNewline will add new-line characters before certain tags.
func WithNewline(gen *Generator) { gen.withNewline = true }

// NewGenerator creates a new generator based on a symbol factory.
func NewGenerator(sf sxpf.SymbolFactory, opts ...Option) *Generator {
	gen := Generator{sf: sf}
	for _, opt := range opts {
		opt(&gen)
	}
	if sf == nil {
		return &gen
	}
	gen.symAttr = sf.MustMake(NameSymAttr)

	addBinding := func(symName string, key, val sxpf.Object) {
		if sym := sf.MustMake(symName); sym.Assoc(key) == nil {
			sym.Cons(key, val)
		}
	}

	for _, voidTag := range voidTags {
		addBinding(voidTag, keyVoid, sxpf.Nil())
	}

	for _, urlAttr := range urlAttrs {
		addBinding(urlAttr, keyAttr, attrURL)
	}

	addBinding("style", keyAttr, attrCSS)

	if gen.withNewline {
		for _, alNlTag := range allNLTags {
			addBinding(alNlTag, keyAlwaysNL, sxpf.Nil())
		}
		for _, nlTag := range nlTags {
			addBinding(nlTag, keyWithNL, sxpf.Nil())
		}
	}

	for _, emptyTag := range emptyTags {
		addBinding(emptyTag, keyIgnoreEmpty, sxpf.Nil())
	}

	return &gen
}

// Special elements / attributes
var (
	// Void elements: https://html.spec.whatwg.org/multipage/syntax.html#void-elements
	voidTags = []string{
		"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr"}
	// Attributes with URL values: https://html.spec.whatwg.org/multipage/indices.html#attributes-1
	urlAttrs = []string{
		"action", "cite", "data", "formaction", "href", "itemid", "itemprop", "itemtype", "ping", "poster", "src"}
	allNLTags = []string{
		"head", "link", "meta", "title",
		"div",
	}
	nlTags = []string{
		NameSymCDATA,
		"head", "link", "meta", "title",
		"script",
		"body", "article", "details", "div", "header", "footer", "form", "main", "summary",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"li", "ol", "ul",
		"dd", "dt", "dl",
		"table", "thead", "tbody", "tr",
		"section",
		"input",
	}
	// Elements that may be ignored if empty.
	emptyTags = []string{"div", "span", "code", "kbd", "p", "samp"}
)

// WriteHTML emit HTML code for the s-expression to the given writer.
func (gen *Generator) WriteHTML(w io.Writer, obj sxpf.Object) (int, error) {
	enc := myEncoder{gen: gen, pr: printer{w: w}, lastWasTag: true}
	enc.generate(obj)
	return enc.pr.length, enc.pr.err
}

// WriteListHTML emits HTML code for a list of s-expressions to the given writer.
func (gen *Generator) WriteListHTML(w io.Writer, lst *sxpf.Pair) (int, error) {
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

func (enc *myEncoder) generate(obj sxpf.Object) {
	switch o := obj.(type) {
	case sxpf.String:
		enc.pr.printHTML(o.String())
		enc.lastWasTag = false
	case *sxpf.Pair:
		if o.IsNil() {
			enc.lastWasTag = false
			return
		}
		if sym, isSymbol := sxpf.GetSymbol(o.Car()); isSymbol {
			tail := o.Tail()
			if s := sym.String(); s[0] == '@' {
				switch s {
				case NameSymCDATA:
					enc.writeCDATA(tail)
				case NameSymNoEscape:
					enc.writeNoEscape(tail)
				case NameSymInlineComment:
					enc.writeComment(tail)
				case NameSymBlockComment:
					enc.writeCommentML(tail)
				case NameSymDoctype:
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

func (enc *myEncoder) generateList(lst *sxpf.Pair) {
	for n := lst; n != nil; n = n.Tail() {
		enc.generate(n.Car())
	}
}

func (enc *myEncoder) writeCDATA(elems *sxpf.Pair) {
	enc.pr.printString("<![CDATA[")
	enc.writeNoEscape(elems)
	enc.pr.printString("]]>")
}

func (enc *myEncoder) writeNoEscape(elems *sxpf.Pair) {
	for n := elems; n != nil; n = n.Tail() {
		if s, isString := sxpf.GetString(n.Car()); isString {
			enc.pr.printString(s.String())
		}
	}
}

func (enc *myEncoder) writeComment(elems *sxpf.Pair) {
	enc.pr.printString("<!--")
	for n := elems; n != nil; n = n.Tail() {
		commentVal := n.Car() // TODO: check for valid node types
		enc.pr.printString(" ")
		enc.pr.printComment(commentVal.String())
	}
	enc.pr.printString(" -->")
}
func (enc *myEncoder) writeCommentML(elems *sxpf.Pair) {
	enc.pr.printString("<!--")
	for n := elems; n != nil; n = n.Tail() {
		commentVal := n.Car() // TODO: check for valid node types
		enc.pr.printString("\n")
		enc.pr.printComment(commentVal.String())
	}
	enc.pr.printString("\n-->\n")
}

func (enc *myEncoder) writeDoctype(elems *sxpf.Pair) {
	// TODO: check for multiple doctypes, error on second
	enc.pr.printString("<!DOCTYPE html>\n")
	enc.generateList(elems)
}

func (enc *myEncoder) writeTag(sym *sxpf.Symbol, elems *sxpf.Pair) {
	if sym.Assoc(keyIgnoreEmpty) != nil && ignoreEmptyStrings(elems) == nil {
		return
	}
	withNewline := enc.gen.withNewline && sym.Assoc(keyWithNL) != nil
	tagName := sym.String()
	if withNewline && (!enc.lastWasTag || sym.Assoc(keyAlwaysNL) != nil) {
		enc.pr.printStrings("\n<", tagName)
	} else {
		enc.pr.printStrings("<", tagName)
	}
	if attrs := enc.getAttributes(elems); attrs != nil {
		enc.writeAttributes(attrs)
		elems = elems.Tail()
	}
	enc.pr.printString(">")
	if sym.Assoc(keyVoid) != nil {
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

func ignoreEmptyStrings(elem *sxpf.Pair) *sxpf.Pair {
	for node := elem; node != nil; node = node.Tail() {
		if s, isString := sxpf.GetString(node.Car()); !isString || s != "" {
			return node
		}
	}
	return nil
}

func (enc *myEncoder) getAttributes(lst *sxpf.Pair) *sxpf.Pair {
	pair, isPair := sxpf.GetPair(lst.Car())
	if !isPair || pair == nil {
		return nil
	}
	sym, isSymbol := sxpf.GetSymbol(pair.Car())
	if !isSymbol || !sym.IsEqual(enc.gen.symAttr) {
		return nil
	}
	return pair.Tail()
}

func (enc *myEncoder) writeAttributes(attrs *sxpf.Pair) {
	length := attrs.Length()
	found := make(map[string]struct{}, length)
	empty := make(map[string]struct{}, length)
	a := make(map[string]string, length)
	sf := enc.gen.sf
	for node := attrs; node != nil; node = node.Tail() {
		pair, isPair := sxpf.GetPair(node.Car())
		if !isPair {
			continue
		}
		sym, isSymbol := sxpf.GetSymbol(pair.Car())
		if !isSymbol {
			continue
		}
		key := sym.String()
		if _, found := found[key]; found {
			continue
		}
		found[key] = struct{}{}
		if cdr := pair.Cdr(); !sxpf.IsNil(cdr) {
			var obj sxpf.Object
			if tail, isTail := sxpf.GetPair(cdr); isTail {
				obj = tail.Car()
			} else {
				obj = cdr
			}
			if sxpf.False.IsEql(obj) || sxpf.IsList(obj) {
				continue
			}
			a[key] = strings.TrimSpace(getAttributeValue(sym, obj, sf))
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

func getAttributeValue(sym *sxpf.Symbol, value sxpf.Object, sf sxpf.SymbolFactory) string {
	switch getAttributeType(sym, sf) {
	case attrURL:
		return urlEscape(value.String())
	default:
		return value.String()
	}
}

func getAttributeType(sym *sxpf.Symbol, sf sxpf.SymbolFactory) sxpf.Keyword {
	name := sym.String()
	if dataName, isData := strings.CutPrefix(name, "data-"); isData {
		name = dataName
		sym = sf.MustMake(name)
	} else if prefix, rest, hasPrefix := strings.Cut(name, ":"); hasPrefix {
		if prefix == "xmlns" {
			return attrURL
		}
		name = rest
		sym = sf.MustMake(name)
	}

	if p := sym.Assoc(keyAttr); p != nil {
		return p.Cdr().(sxpf.Keyword)
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
