// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package webmention provides functions for discovering the webmention
// endpoint for URLs, and sending webmentions according to
// http://webmention.org/.
package webmention // import "willnorris.com/go/webmention"

import (
	"fmt"
	"io"
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
	headEndpoint, err := c.discoverRequest(http.MethodHead, urlStr)
	if err == nil && headEndpoint != "" {
		return headEndpoint, nil
	}

	getEndpoint, err := c.discoverRequest(http.MethodGet, urlStr)
	if err == nil && getEndpoint != "" {
		return getEndpoint, nil
	}

	return "", err
}

func (c *Client) discoverRequest(method, urlStr string) (string, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code < 200 || 300 <= code {
		_, _ = io.Copy(io.Discard, resp.Body)
		return "", fmt.Errorf("response error: %v", resp.StatusCode)
	}

	endpoint, err := extractEndpoint(resp)
	if err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return "", err
	}

	urls, err := resolveReferences(resp.Request.URL.String(), endpoint)
	if err != nil {
		return "", err
	}
	return urls[0], nil
}

func extractEndpoint(resp *http.Response) (string, error) {
	// first check http link headers
	if endpoint, err := httpLink(resp.Header); err == nil {
		return endpoint, nil
	}

	// then look in the HTML body
	endpoint, err := htmlLink(resp.Body)
	if err != nil {
		return "", err
	}
	return endpoint, nil
}

// DiscoverLinks discovers URLs that the provided resource links to.  These are
// candidates for sending webmentions to.  If non-empty, sel is a CSS selector
// identifying the root node(s) to search in for links.
func (c *Client) DiscoverLinks(urlStr string, sel string) ([]string, error) {
	resp, err := c.Client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	if code := resp.StatusCode; code < 200 || 300 <= code {
		return nil, fmt.Errorf("response error: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	return DiscoverLinksFromReader(resp.Body, urlStr, sel)
}

// DiscoverLinksFromReader discovers URLs in the HTML read from 'r'. Relative
// URLs found the HTML are resolved against 'baseURL'. These are candidates for
// sending webmentions to.  If non-empty, sel is a CSS selector identifying the
// root node(s) to search in for links.
func DiscoverLinksFromReader(r io.Reader, baseURL string, sel string) ([]string, error) {
	// TODO: should we include HTTP header links?
	links, err := parseLinks(r, sel)
	if err != nil {
		return nil, err
	}

	urls, err := resolveReferences(baseURL, links...)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

// resolveReferences resolves each URL in refs into an absolute URL relative to
// base.  If base is not a valid URL, an error is returned.  If one of the
// values in refs is not a valid URL, it is skipped.
func resolveReferences(base string, refs ...string) ([]string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, r := range refs {
		u, err := url.Parse(r)
		if err != nil {
			continue
		}
		urls = append(urls, b.ResolveReference(u).String())
	}
	return urls, nil
}
