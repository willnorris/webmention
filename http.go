package webmention

import (
	"net/http"

	"willnorris.com/go/webmention/third_party/header"
)

// httpLink parses headers and returns the URL of the first link that contains
// a webmention rel value.
func httpLink(headers http.Header) string {
	for _, h := range header.ParseList(headers, "Link") {
		link := header.ParseLink(h)
		for _, v := range link.Rel {
			if v == relWebmention || v == relLegacy || v == relLegacySlash {
				return link.Href
			}
		}
	}
	return ""
}
