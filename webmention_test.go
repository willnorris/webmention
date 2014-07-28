package webmention

import (
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
	_, err = client.SendWebmention(server.URL+"/bad", "", "")
	if err == nil {
		t.Errorf("SendWebmention did not return expected error")
	}
}
