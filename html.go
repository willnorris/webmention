// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package webmention

import (
	"io"
	"strings"

	"code.google.com/p/go.net/html"
)

// htmlLink parses r as HTML and returns the URL of the first link that
// contains a webmention rel value.  HTML <link> elements are preferred,
// falling back to <a> elements if no webmention <link> elements are found.
func htmlLink(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	// the first webmention link found in an <a> element, used only if no
	// webmention <link> elements are found.
	var aLink string

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode {
			if n.Data == "link" {
				if href := parseLinkNode(n); href != "" {
					return href
				}
			}
			if n.Data == "a" && aLink == "" {
				aLink = parseLinkNode(n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if link := f(c); link != "" {
				return link
			}
		}
		return ""
	}

	link := f(doc)
	if link == "" {
		link = aLink
	}
	return link, nil
}

// parseLinkNode returns the href value of n if it contains a webmention rel value.
func parseLinkNode(n *html.Node) string {
	if n == nil {
		return ""
	}

	var href, rel string
	for _, a := range n.Attr {
		if a.Key == "href" {
			href = a.Val
		}
		if a.Key == "rel" {
			rel = a.Val
		}
	}
	for _, v := range strings.Split(rel, " ") {
		if v == relWebmention || v == relLegacy || v == relLegacySlash {
			return href
		}
	}
	return ""
}
