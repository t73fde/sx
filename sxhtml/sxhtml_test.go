//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of sx.
//
// sx is licensed under the latest version of the EUPL // (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package sxhtml_test

import (
	"strings"
	"testing"

	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxhtml"
	"zettelstore.de/sx.fossil/sxreader"
)

type testcase struct {
	name string
	src  string
	exp  string
}

func TestSXHTML(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{name: "Empty", src: `()`, exp: ``},
		{name: "JustHTML", src: `(html)`, exp: `<html></html>`},
		{name: "SimpleNested", src: `(p (b "bold") "text")`, exp: `<p><b>bold</b>text</p>`},
		{name: "NoEndTag", src: `(br)`, exp: `<br>`},
		{name: "NoEscape", src: `(@H "&amp;")`, exp: `&amp;`},
		{name: "Escape", src: `"&amp;"`, exp: `&amp;amp;`},
		{name: "DoctypeInline", src: `(@@@@ (html))`, exp: "<!DOCTYPE html>\n<html></html>"},
		{name: "SimpleComment", src: `(@@ "comment")`, exp: `<!-- comment -->`},
		{name: "SimpleCommentEsc", src: `(@@ "esc -->")`, exp: `<!-- esc -&#45;> -->`},
		{name: "CommentWrongMinus", src: `(@@ "-------->")`, exp: `<!-- -&#45;-&#45;-&#45;-&#45;> -->`},
		{name: "SimpleCommentML", src: `(@@@ "line1" "line2")`, exp: "<!--\nline1\nline2\n-->\n"},
		{name: "SimpleCommentMLEsc", src: `(@@@ "line1" "-----")`, exp: "<!--\nline1\n-&#45;-&#45;-\n-->\n"},
		{name: "SimpleHTMLEsc", src: `(p "&")`, exp: `<p>&amp;</p>`},
		{name: "CDATA", src: `(@C "abc")`, exp: `<![CDATA[abc]]>`},

		{name: "NoValueAttr", src: `(p (@ (checked . ())))`, exp: `<p checked></p>`},
		{name: "NoValueAttrSimple", src: `(p (@ (checked)))`, exp: `<p checked></p>`},
		{name: "EmptyValueAttr", src: `(p (@ (checked . "")))`, exp: `<p checked=""></p>`},
		{name: "EmptyValueAttr2", src: `(p (@ (checked "")))`, exp: `<p checked=""></p>`},
		{name: "SpaceValueAttr", src: `(p (@ (checked " ")))`, exp: `<p checked=""></p>`},
		{name: "SingleValueAttr", src: `(p (@ (id . "a")))`, exp: `<p id="a"></p>`},
		{name: "SingleValueAttrNoDOT", src: `(p (@ (id "a")))`, exp: `<p id="a"></p>`},
		{name: "SimpleAttrEsc", src: `(p (@ (name . "\"")))`, exp: `<p name="&quot;"></p>`},
		{name: "SimpleAttrEscNoDOT", src: `(p (@ (name "\"")))`, exp: `<p name="&quot;"></p>`},
		{name: "DoubleAttr", src: `(p (@ (id "1") (id "2")))`, exp: `<p id="1"></p>`},
		{name: "SimpleURLAttr", src: `(a (@ (href . "search?q=%&r=Ä")))`, exp: `<a href="search?q=%25&amp;r=%c3%84"></a>`},
		{name: "SimpleURLAttrNoDOT", src: `(a (@ (href "search?q=%&r=Ä")))`, exp: `<a href="search?q=%25&amp;r=%c3%84"></a>`},
		{name: "SortedAttr", src: `(p (@ (z . z) (a a)))`, exp: `<p a="a" z="z"></p>`},
		{name: "DoubleAttr", src: `(p (@ (a . z) (a a)))`, exp: `<p a="z"></p>`},
		{name: "DeletedAttr", src: `(p (@ (a ()) (z z) (a a)))`, exp: `<p z="z"></p>`},
		{name: "EmptyAttrKey", src: `(p (@ ("" . a)))`, exp: `<p></p>`},
		{name: "NilAttrKey", src: `(p (@ (() . a)))`, exp: `<p></p>`},
		{name: "StringAttr", src: `(p (@ (a . "b")))`, exp: `<p a="b"></p>`},
		{name: "SymbolAttr", src: `(p (@ (a . b)))`, exp: `<p a="b"></p>`},
		{name: "NumberAttr", src: `(p (@ (a . 7)))`, exp: `<p a="7"></p>`},
		{name: "ListAttr", src: `(p (@ (a (1))))`, exp: `<p></p>`},

		{name: "IgnoreEmptyTag", src: `(p)`, exp: ``},
		{name: "IgnoreTagWithEmptyString", src: `(div "")`, exp: ``},
		{name: "IgnoreTagWithEmptyString2", src: `(div "" "")`, exp: ``},
		{name: "NoIgnoreTagWithTagAfterEmptySpace", src: `(div "" (p "A"))`, exp: `<div><p>A</p></div>`},

		{name: "JustAList", src: `(@L "a" "b")`, exp: `ab`},
	}
	checkTestcases(t, testcases, func() *sxhtml.Generator { return sxhtml.NewGenerator() })
}

func TestWithNewline(t *testing.T) {
	testcases := []testcase{
		{name: "HeadBody", src: `(@@@@ (html (head (title "T"))))`, exp: "<!DOCTYPE html>\n<html>\n<head>\n<title>T</title>\n</head>\n</html>"},
	}
	checkTestcases(t, testcases, func() *sxhtml.Generator {
		return sxhtml.NewGenerator(sxhtml.WithNewline)
	})
}

func checkTestcases(t *testing.T, testcases []testcase, newGen func() *sxhtml.Generator) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.src))
			val, err := rd.Read()
			if err != nil {
				t.Error(err)
				return
			}

			gen := newGen()
			var sb strings.Builder
			_, err = gen.WriteHTML(&sb, val)
			if err != nil {
				t.Error(err)
				return
			}

			if got := sb.String(); tc.exp != got {
				t.Errorf("\nSexpr:    %v\nExpected: %v\nGot:      %v", tc.src, tc.exp, got)
			}

			sb.Reset()
			_, err = gen.WriteListHTML(&sb, sx.MakeList(val, sx.String("/"), val))
			if err != nil {
				t.Error(err)
				return
			}

			exp := tc.exp + "/" + tc.exp
			if got := sb.String(); exp != got {
				t.Errorf("\nSexpr:    %v\nExpected: %v\nGot:      %v", tc.src, exp, got)
			}
		})
	}

}
