// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package webmention

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testServer() (*http.ServeMux, *httptest.Server, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	cleanup := func() {
		server.Close()
	}
	return mux, server, cleanup
}

func TestClient_SendWebmention(t *testing.T) {
	mux, server, cleanup := testServer()
	defer cleanup()

	source, target := "S", "T"

	mux.HandleFunc("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		if got := r.PostFormValue("source"); got != source {
			t.Errorf("request contained source: %v, want %v", got, source)
		}
		if got := r.PostFormValue("target"); got != target {
			t.Errorf("request contained target: %v, want %v", got, target)
		}
	})

	client := New(nil)
	resp, err := client.SendWebmention(server.URL+"/endpoint", source, target)
	resp.Body.Close()
	if err != nil {
		t.Errorf("SendWebmention returned error: %v", err)
	}

	// ensure 404 response is returned as error
	resp, err = client.SendWebmention(server.URL+"/bad", "", "")
	resp.Body.Close()
	if err == nil {
		t.Errorf("SendWebmention did not return expected error")
	}
}

func TestClient_DiscoverEndpoint(t *testing.T) {
	mux, server, cleanup := testServer()
	defer cleanup()
	client := New(nil)

	// valid request with link
	mux.HandleFunc("/good", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<link href="/endpoint" rel="webmention">`)
	})

	want := server.URL + "/endpoint" // want absolute URL
	if got, err := client.DiscoverEndpoint(server.URL + "/good"); err != nil {
		t.Errorf("DiscoverEndpoint(%q) returned error: %v", server.URL+"/good", err)
	} else if got != want {
		t.Errorf("DiscoverEndpoint(%q) returned %v, want %v", server.URL+"/good", got, want)
	}

	// valid request with no link
	mux.HandleFunc("/nolink", func(w http.ResponseWriter, r *http.Request) {
	})

	want = ""
	if got, err := client.DiscoverEndpoint(server.URL + "/nolink"); err != errNoWebmentionRel {
		t.Errorf("DiscoverEndpoint(%q) returned error: %v", server.URL+"/nolink", err)
	} else if got != want {
		t.Errorf("DiscoverEndpoint(%q) returned %v, want %v", server.URL+"/nolink", got, want)
	}

	// empty endpoint is a valid relative URL pointing to the page itself
	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<link href="" rel="webmention">`)
	})

	want = server.URL + "/empty" // want absolute URL
	if got, err := client.DiscoverEndpoint(server.URL + "/empty"); err != nil {
		t.Errorf("DiscoverEndpoint(%q) returned error: %v", server.URL+"/empty", err)
	} else if got != want {
		t.Errorf("DiscoverEndpoint(%q) returned %v, want %v", server.URL+"/empty", got, want)
	}

	// ensure 404 response is returned as error
	if _, err := client.DiscoverEndpoint(server.URL + "/bad"); err == nil {
		t.Errorf("DiscoverEndpoint(%q) did not return expected error", server.URL+"/bad")
	}
}

func TestExtractEndpoint(t *testing.T) {
	tests := []struct {
		resp string // raw response header and body
		want string // wanted endpoint URL
	}{
		{
			`Link: </endpoint>; rel="webmention"

`,
			"/endpoint",
		},
		{
			`
<link href="/endpoint" rel="webmention">`,
			"/endpoint",
		},
		{
			`Link: </endpoint1>; rel="webmention"

<link href="/endpoint2" rel="webmention">`,
			"/endpoint1",
		},
	}

	for _, tt := range tests {
		raw := "HTTP/1.1 200 OK\n" + tt.resp
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBufferString(raw)), nil)
		if err != nil {
			t.Errorf("error reading response %q: %v", raw, err)
		}

		if got, err := extractEndpoint(resp); err != nil {
			t.Errorf("extractEndpoint(%q) returned error: %v", raw, err)
		} else if got != tt.want {
			t.Errorf("extractEndpoint(%q) returned %v, want %v", raw, got, tt.want)
		}
	}
}

func TestDiscoverLinks(t *testing.T) {
	mux, server, cleanup := testServer()
	defer cleanup()
	client := New(nil)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html>
<head><link href="/a"></head>
<body><a href="http://example.com/"></a></body>
</html>`)
	})

	// no selector
	got, err := client.DiscoverLinks(server.URL, "")
	if err != nil {
		t.Errorf("DiscoverLinks returned error: %v", err)
	}
	want := []string{server.URL + "/a", "http://example.com/"}
	if !cmp.Equal(got, want) {
		t.Errorf("DiscoverLinks returned %v, want %v", got, want)
	}

	// with selector
	got, err = client.DiscoverLinks(server.URL, "body")
	if err != nil {
		t.Errorf("DiscoverLinks returned error: %v", err)
	}
	want = []string{"http://example.com/"}
	if !cmp.Equal(got, want) {
		t.Errorf("DiscoverLinks returned %v, want %v", got, want)
	}
}
