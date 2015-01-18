// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package header

import (
	"net/http"
	"reflect"
	"testing"
)

var getHeaderListTests = []struct {
	s string
	l []string
}{
	{s: `a`, l: []string{`a`}},
	{s: `a, b , c `, l: []string{`a`, `b`, `c`}},
	{s: `a,, b , , c `, l: []string{`a`, `b`, `c`}},
	{s: `a,b,c`, l: []string{`a`, `b`, `c`}},
	{s: ` a b, c d `, l: []string{`a b`, `c d`}},
	{s: `"a, b, c", d `, l: []string{`"a, b, c"`, "d"}},
	{s: `","`, l: []string{`","`}},
	{s: `"\""`, l: []string{`"\""`}},
	{s: `" "`, l: []string{`" "`}},
}

func TestGetHeaderList(t *testing.T) {
	for _, tt := range getHeaderListTests {
		header := http.Header{"Foo": {tt.s}}
		if l := ParseList(header, "foo"); !reflect.DeepEqual(tt.l, l) {
			t.Errorf("ParseList for %q = %q, want %q", tt.s, l, tt.l)
		}
	}
}

var parseValueAndParamsTests = []struct {
	s      string
	value  string
	params map[string]string
}{
	{`text/html`, "text/html", map[string]string{}},
	{`text/html  `, "text/html", map[string]string{}},
	{`text/html ; `, "text/html", map[string]string{}},
	{`tExt/htMl`, "text/html", map[string]string{}},
	{`tExt/htMl; fOO=";"; hellO=world`, "text/html", map[string]string{
		"hello": "world",
		"foo":   `;`,
	}},
	{`text/html; foo=bar, hello=world`, "text/html", map[string]string{"foo": "bar"}},
	{`text/html ; foo=bar `, "text/html", map[string]string{"foo": "bar"}},
	{`text/html ;foo=bar `, "text/html", map[string]string{"foo": "bar"}},
	{`text/html; foo="b\ar"`, "text/html", map[string]string{"foo": "bar"}},
	{`text/html; foo="bar\"baz\"qux"`, "text/html", map[string]string{"foo": `bar"baz"qux`}},
	{`text/html; foo="b,ar"`, "text/html", map[string]string{"foo": "b,ar"}},
	{`text/html; foo="b;ar"`, "text/html", map[string]string{"foo": "b;ar"}},
	{`text/html; FOO="bar"`, "text/html", map[string]string{"foo": "bar"}},
	{`form-data; filename="file.txt"; name=file`, "form-data", map[string]string{"filename": "file.txt", "name": "file"}},
}

func TestParseLink(t *testing.T) {
	tests := []struct {
		s    string
		want Link
	}{
		{`</foo>; rel="a"`, Link{"/foo", []string{"a"}}},
		{`</foo>; rel="a b"; rel="c"`, Link{"/foo", []string{"a", "b"}}},
	}

	for _, tt := range tests {
		if got := ParseLink(tt.s); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ParseLink(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}
