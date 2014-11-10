// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package webmention

import (
	"io"
	"strings"

	"code.google.com/p/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// htmlLink parses r as HTML and returns the URL of the first link that
// contains a webmention rel value.  HTML <link> elements are preferred,
// falling back to <a> elements if no webmention <link> elements are found.
func htmlLink(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode {
			if n.DataAtom == atom.Link || n.DataAtom == atom.A {
				var href, rel string
				for _, a := range n.Attr {
					if a.Key == atom.Href.String() {
						href = a.Val
					}
					if a.Key == atom.Rel.String() {
						rel = a.Val
					}
				}
				if len(href) > 0 && len(rel) > 0 {
					for _, v := range strings.Split(rel, " ") {
						if v == relWebmention || v == relLegacy || v == relLegacySlash {
							return href
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if link := f(c); link != "" {
				return link
			}
		}
		return ""
	}

	return f(doc), nil
}

// parseLinks parses r as HTML and returns all URLs linked to (from either a
// <link> or <a> element).  If non-empty, rootSelector is a CSS selector
// identifying the root node(s) to search in for links.
//
// TODO: return full links rather than just URLs, since other metadata may be useful
func parseLinks(r io.Reader, rootSelector string) ([]string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var sel cascadia.Selector
	if rootSelector != "" {
		sel, err = cascadia.Compile(rootSelector)
		if err != nil {
			return nil, err
		}
	}

	var urls []string

	var f func(*html.Node, bool)
	f = func(n *html.Node, capture bool) {
		capture = capture || sel.Match(n)
		if capture {
			if n.Type == html.ElementNode && (n.Data == "link" || n.Data == "a") {
				for _, a := range n.Attr {
					if a.Key == "href" {
						urls = append(urls, a.Val)
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, capture)
		}
	}

	// if no selector specified, capture everything
	capture := (sel == nil)

	f(doc, capture)
	return urls, nil
}
