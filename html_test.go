// Copyright (c) The webmention project authors.
// SPDX-License-Identifier: BSD-3-Clause

package webmention

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHtmlLink(t *testing.T) {
	tests := []struct {
		input, want string
		wantErr     error
	}{
		// basic links
		{`<link href="foo" rel="webmention">`, "foo", nil},
		{`<a href="foo" rel="webmention">`, "foo", nil},
		// different attribute order
		{`<link rel="webmention" href="foo">`, "foo", nil},
		// line breaks inside element
		{`<link
			rel="webmention" 
			href="foo">`, "foo", nil},
		// multiple rel values
		{`<link rel="a webmention b" href="foo">`, "foo", nil},
		// legacy rel value
		{`<link rel="http://webmention.org" href="foo">`, "foo", nil},
		// legacy rel value with slash
		{`<link rel="http://webmention.org/" href="foo">`, "foo", nil},
		// invalid legacy rel value
		{`<link rel="https://webmention.org" href="foo">`, "", ErrNoEndpointFound},
		// no rel value
		{`<link href="foo">`, "", ErrNoEndpointFound},
		// multiple links, only one for webmention
		{`<a href="foo" rel="web"><a href="bar" rel="webmention">`, "bar", nil},
		// multiple webmention links, return first
		{`<a href="foo" rel="webmention"><a href="bar" rel="webmention">`, "foo", nil},
	}

	for _, tt := range tests {
		buf := bytes.NewBufferString(tt.input)
		if got, err := htmlLink(buf); err != tt.wantErr {
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
		} else if want := tt.want; !cmp.Equal(got, want) {
			t.Errorf("parseLinks(%q, %q) returned %v, want %v", tt.input, tt.sel, got, want)
		}
	}
}
