// Copyright (c) The webmention project authors.
// SPDX-License-Identifier: BSD-3-Clause

package webmention

import (
	"fmt"
	"net/http"

	"willnorris.com/go/webmention/third_party/header"
)

// ErrNoEndpointFound is returned when no endpoint can be found for a certain
// target URL.
var ErrNoEndpointFound = fmt.Errorf("no endpoint found")

// httpLink parses headers and returns the URL of the first link that contains
// a webmention rel value.
func httpLink(headers http.Header) (string, error) {
	for _, h := range header.ParseList(headers, "Link") {
		link := header.ParseLink(h)
		for _, v := range link.Rel {
			if v == relWebmention || v == relLegacy || v == relLegacySlash {
				return link.Href, nil
			}
		}
	}
	return "", ErrNoEndpointFound
}
