// Copyright (c) The webmention project authors.
// SPDX-License-Identifier: BSD-3-Clause

package webmention

import (
	"net/http"
	"testing"
)

func TestHttpLink(t *testing.T) {
	tests := []struct {
		input   []string
		want    string
		wantErr error
	}{
		{[]string{`<foo>; rel="webmention"`}, "foo", nil},
		{[]string{`<foo>; rel="a webmention b"`}, "foo", nil},
		{[]string{`<foo>; rel="http://webmention.org"`}, "foo", nil},
		{[]string{`<foo>; rel="http://webmention.org/"`}, "foo", nil},
		{[]string{`<foo>; rel="https://webmention.org"`}, "", ErrNoEndpointFound},
		{[]string{`<foo>`}, "", ErrNoEndpointFound},
		{[]string{`<foo>; rel="a", <bar>; rel="webmention"`}, "bar", nil},
		{[]string{`<foo>; rel="a"`, `<bar>; rel="webmention"`}, "bar", nil},
		{[]string{`<foo>; rel="webmention", <bar>; rel="webmention"`}, "foo", nil},
		{[]string{`<foo>; rel="webmention"`, `<bar>; rel="webmention"`}, "foo", nil},
		{[]string{`<>; rel="webmention"`}, "", nil},
	}

	for _, tt := range tests {
		headers := make(http.Header)
		for _, i := range tt.input {
			headers.Add("Link", i)
		}
		if got, gotErr := httpLink(headers); got != tt.want || gotErr != tt.wantErr {
			t.Errorf("httpLink(%q) got %v (error %v), want %v (error %v)", headers, got, gotErr, tt.want, tt.wantErr)
		}
	}
}
