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

// Provides some function to pretty-print objects.

import (
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxeval"
)

// Pretty writes the first argument to stdout.
var Pretty = sxeval.Builtin{
	Name:     "pp",
	MinArity: 1,
	MaxArity: 1,
	TestPure: nil,
	Fn1: func(_ *sxeval.Environment, arg sx.Object, _ *sxeval.Frame) (sx.Object, error) {
		_, err := Print(os.Stdout, arg)
		return sx.Nil(), err
	},
}

const (
	defaultWidth  = 80
	defaultIndent = 2
	maxHeadAlign  = 10 // longest head symbol whose arguments are aligned below its first argument
)

// Options controls the layout of the pretty printer.
type Options struct {
	Width  int // maximum line width; values <= 0 select defaultWidth
	Indent int // indentation of bodies and long heads; values <= 0 select defaultIndent

	// SpecialForms maps a head symbol to the number of arguments that stay on
	// the first line; the remaining elements are the body, indented by Indent.
	// A nil or empty map disables special forms, which is the default.
	// The map must not be modified during printing.
	SpecialForms map[string]int
}

// Print object to given writer in a pretty way, using default options.
func Print(w io.Writer, obj sx.Object) (int, error) {
	return PrintWith(w, obj, Options{})
}

// PrintWidth prints object to given writer, trying to keep lines within width columns.
func PrintWidth(w io.Writer, obj sx.Object, width int) (int, error) {
	return PrintWith(w, obj, Options{Width: width})
}

// PrintWith prints object to given writer, using the given options.
func PrintWith(w io.Writer, obj sx.Object, opts Options) (int, error) {
	if opts.Width <= 0 {
		opts.Width = defaultWidth
	}
	if opts.Indent <= 0 {
		opts.Indent = defaultIndent
	}
	p := printer{w: w, width: opts.Width, indent: opts.Indent, special: opts.SpecialForms}
	p.node(build(obj), 0)
	p.str("\n")
	return p.written, p.err
}

// node is the intermediate tree that knows its width in flat layout.
type node struct {
	text   string  // text of an atom
	items  []*node // elements of a list
	tail   *node   // tail of a dotted list, or nil
	width  int     // width in runes if printed flat
	isList bool
}

func build(obj sx.Object) *node {
	if sx.IsNil(obj) {
		return &node{text: "()", width: 2}
	}
	pair, isPair := sx.GetPair(obj)
	if !isPair {
		var sb strings.Builder
		_, _ = sx.Print(&sb, obj) // strings.Builder never fails
		s := sb.String()
		return &node{text: s, width: utf8.RuneCountInString(s)}
	}

	n := &node{isList: true}
	width := 2 // parentheses
	for {
		item := build(pair.Car())
		n.items = append(n.items, item)
		width += item.width

		cdr := pair.Cdr()
		if sx.IsNil(cdr) {
			break
		}
		if next, ok := sx.GetPair(cdr); ok {
			pair = next
			continue
		}
		n.tail = build(cdr)
		width += 3 + n.tail.width // " . " + tail
		break
	}
	width += len(n.items) - 1 // separating spaces
	n.width = width
	return n
}

type printer struct {
	w       io.Writer
	width   int
	indent  int
	special map[string]int
	col     int
	written int
	err     error
}

func (p *printer) str(s string) {
	if p.err != nil {
		return
	}
	n, err := io.WriteString(p.w, s)
	p.written += n
	p.err = err
	p.col += utf8.RuneCountInString(s)
}

// 80 spaces
const spaces = "                                                                                "

func (p *printer) newline(indent int) {
	p.str("\n")
	p.col = 0
	for indent > len(spaces) {
		p.str(spaces)
		indent -= len(spaces)
	}
	p.str(spaces[:indent])
}

// node prints n; trail is the number of closing parentheses that must follow on the same line.
func (p *printer) node(n *node, trail int) {
	if !n.isList || p.col+n.width+trail <= p.width {
		p.flat(n)
		return
	}
	p.broken(n, trail)
}

func (p *printer) flat(n *node) {
	if !n.isList {
		p.str(n.text)
		return
	}
	p.str("(")
	for i, item := range n.items {
		if i > 0 {
			p.str(" ")
		}
		p.flat(item)
	}
	if n.tail != nil {
		p.str(" . ")
		p.flat(n.tail)
	}
	p.str(")")
}

func (p *printer) broken(n *node, trail int) {
	start := p.col
	p.str("(")
	items := n.items
	align := start + 1
	pos := 0
	special := false
	if head := items[0]; !head.isList && len(items) > 1 {
		p.str(head.text)
		pos = 1
		if k, ok := p.special[head.text]; ok && len(items) > k+1 {
			for ; pos <= k; pos++ {
				p.str(" ")
				p.node(items[pos], 0)
			}
			align = start + p.indent
			special = true
			p.newline(align) // first body element starts on its own line
		} else if head.width <= maxHeadAlign {
			align = start + 2 + head.width
			p.str(" ")
		} else {
			align = start + p.indent
			p.newline(align)
		}
	}

	fill := !special && n.tail == nil && allAtoms(items[pos:])
	for i := pos; i < len(items); i++ {
		item := items[i]
		t := 0
		if i == len(items)-1 && n.tail == nil {
			t = trail + 1 // closing parenthesis of this list
		}
		if i > pos {
			if fill && p.col+1+item.width+t <= p.width {
				p.str(" ")
			} else {
				p.newline(align)
			}
		}
		p.node(item, t)
	}

	if n.tail != nil {
		p.newline(align)
		p.str(". ")
		p.node(n.tail, trail+1)
	}
	p.str(")")
}

func allAtoms(items []*node) bool {
	for _, item := range items {
		if item.isList {
			return false
		}
	}
	return true
}
