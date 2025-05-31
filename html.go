// Copyright (c) The webmention project authors.
// SPDX-License-Identifier: BSD-3-Clause

package webmention

import (
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
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

	var f func(*html.Node) (string, error)
	f = func(n *html.Node) (string, error) {
		if n.Type == html.ElementNode {
			if n.DataAtom == atom.Link || n.DataAtom == atom.A {
				var href, rel string
				var hrefFound, relFound bool
				for _, a := range n.Attr {
					if a.Key == atom.Href.String() {
						href = a.Val
						hrefFound = true
					}
					if a.Key == atom.Rel.String() {
						rel = a.Val
						relFound = true
					}
				}
				if hrefFound && relFound {
					for _, v := range strings.Split(rel, " ") {
						if v == relWebmention || v == relLegacy || v == relLegacySlash {
							return href, nil
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if link, err := f(c); err == nil {
				return link, nil
			}
		}
		return "", ErrNoEndpointFound
	}

	return f(doc)
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
