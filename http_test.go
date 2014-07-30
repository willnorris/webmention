// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package webmention

import (
	"net/http"
	"testing"
)

func TestHttpLink(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{`<foo>; rel="webmention"`}, "foo"},
		{[]string{`<foo>; rel="a webmention b"`}, "foo"},
		{[]string{`<foo>; rel="http://webmention.org"`}, "foo"},
		{[]string{`<foo>; rel="http://webmention.org/"`}, "foo"},
		{[]string{`<foo>; rel="https://webmention.org"`}, ""},
		{[]string{`<foo>`}, ""},
		{[]string{`<foo>; rel="a", <bar>; rel="webmention"`}, "bar"},
		{[]string{`<foo>; rel="a"`, `<bar>; rel="webmention"`}, "bar"},
		{[]string{`<foo>; rel="webmention", <bar>; rel="webmention"`}, "foo"},
		{[]string{`<foo>; rel="webmention"`, `<bar>; rel="webmention"`}, "foo"},
	}

	for _, tt := range tests {
		headers := make(http.Header)
		for _, i := range tt.input {
			headers.Add("Link", i)
		}
		if got := httpLink(headers); got != tt.want {
			t.Errorf("httpLink(%q) got %v, want %v", headers, got, tt.want)
		}
	}
}
