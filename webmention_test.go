package webmention

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
	_, err := client.SendWebmention(server.URL+"/endpoint", source, target)
	if err != nil {
		t.Errorf("SendWebmention returned error: %v", err)
	}

	// ensure 404 response is returned as error
	if _, err = client.SendWebmention(server.URL+"/bad", "", ""); err == nil {
		t.Errorf("SendWebmention did not return expected error")
	}
}

func TestClient_DiscoverEndpoint(t *testing.T) {
	mux, server, cleanup := testServer()
	defer cleanup()

	mux.HandleFunc("/good", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<link href="/endpoint" rel="webmention">`)
	})

	client := New(nil)
	want := server.URL + "/endpoint" // want absolute URL
	if got, err := client.DiscoverEndpoint(server.URL + "/good"); err != nil {
		t.Errorf("DiscoverEndpoint(%q) returned error: %v", server.URL+"/good", err)
	} else if got != want {
		t.Errorf("DiscoverEndpoint(%q) returned %v, want %v", server.URL+"/good", got, want)
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
