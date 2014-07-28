// Package webmention provides functions for discovering the webmention
// endpoint for URLs, and sending webmentions according to
// http://webmention.org/.
package webmention

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	relWebmention  = "webmention"
	relLegacy      = "http://webmention.org"
	relLegacySlash = "http://webmention.org/"
)

// Client is a webmention client that can discover webmention endpoints and send webmentions.
type Client struct {
	*http.Client
}

// New constructs a new webmention Client using the provided http.Client.  If a
// nil http.Client is provided, http.DefaultClient is used.
func New(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{Client: client}
}

// SendWebmention sends a webmention to endpoint, indicating that source has mentioned target.
func (c *Client) SendWebmention(endpoint, source, target string) (*http.Response, error) {
	resp, err := c.Client.PostForm(endpoint, url.Values{
		"source": []string{source},
		"target": []string{target},
	})
	if err != nil {
		return resp, err
	}
	if code := resp.StatusCode; code < 200 || 300 <= code {
		return resp, fmt.Errorf("response error: %v", resp.StatusCode)
	}
	return resp, nil
}
