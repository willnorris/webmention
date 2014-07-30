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

// DiscoverEndpoint discovers the webmention endpoint for the provided URL.
func (c *Client) DiscoverEndpoint(urlStr string) (string, error) {
	resp, err := c.Client.Get(urlStr)
	if err != nil {
		return "", err
	}
	if code := resp.StatusCode; code < 200 || 300 <= code {
		return "", fmt.Errorf("response error: %v", resp.StatusCode)
	}

	endpoint, err := extractEndpoint(resp)
	if err != nil {
		return "", err
	}

	// resolve relative endpoint URLs
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("error parsing URL %q: %v", urlStr, err)
	}

	e, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("error parsing URL %q: %v", endpoint, err)
	}

	return u.ResolveReference(e).String(), nil
}

func extractEndpoint(resp *http.Response) (string, error) {
	// first check http link headers
	if endpoint := httpLink(resp.Header); endpoint != "" {
		return endpoint, nil
	}

	// then look in the HTML body
	endpoint, err := htmlLink(resp.Body)
	if err != nil {
		return "", err
	}
	return endpoint, nil
}
