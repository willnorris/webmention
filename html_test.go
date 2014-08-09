// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package webmention

import (
	"bytes"
	"reflect"
	"testing"
)

func TestHtmlLink(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		// basic links
		{`<link href="foo" rel="webmention">`, "foo"},
		{`<a href="foo" rel="webmention">`, "foo"},
		// different attribute order
		{`<link rel="webmention" href="foo">`, "foo"},
		// line breaks inside element
		{`<link
			rel="webmention" 
			href="foo">`, "foo"},
		// multiple rel values
		{`<link rel="a webmention b" href="foo">`, "foo"},
		// legacy rel value
		{`<link rel="http://webmention.org" href="foo">`, "foo"},
		// legacy rel value with slash
		{`<link rel="http://webmention.org/" href="foo">`, "foo"},
		// invalid legacy rel value
		{`<link rel="https://webmention.org" href="foo">`, ""},
		// no rel value
		{`<link href="foo">`, ""},
		// multiple links, only one for webmention
		{`<a href="foo" rel="web"><a href="bar" rel="webmention">`, "bar"},
		// multiple webmention links, return first
		{`<a href="foo" rel="webmention"><a href="bar" rel="webmention">`, "foo"},
	}

	for _, tt := range tests {
		buf := bytes.NewBufferString(tt.input)
		if got, err := htmlLink(buf); err != nil {
			t.Errorf("htmlLink(%q) returned error: %v", tt.input, err)
		} else if want := tt.want; got != want {
			t.Errorf("htmlLink(%q) returned %v, want %v", tt.input, got, want)
		}
	}
}

func TestParseLinks(t *testing.T) {
	tests := []struct {
		input string
		sel   string
		want  []string
	}{
		{`<a href="a">`, "", []string{"a"}},
		{`<a href="a"><a href="b">`, "", []string{"a", "b"}},
		{`<a href="a"><link href="b">`, "", []string{"a", "b"}},

		// with selector
		{`<link href="a"><main><a href="b"></main>`, "main", []string{"b"}},
		{`<link href="a"><div class="h-entry"><a href="b"></div>`, ".h-entry", []string{"b"}},
	}

	for _, tt := range tests {
		buf := bytes.NewBufferString(tt.input)
		if got, err := parseLinks(buf, tt.sel); err != nil {
			t.Errorf("parseLinks(%q, %q) returned error: %v", tt.input, tt.sel, err)
		} else if want := tt.want; !reflect.DeepEqual(got, want) {
			t.Errorf("parseLinks(%q, %q) returned %v, want %v", tt.input, tt.sel, got, want)
		}
	}
}
