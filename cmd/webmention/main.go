// Copyright 2014 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// The webmention binary is a command line utiltiy for sending webmentions to
// the URLs linked to by a given webpage.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wsxiaoys/terminal/color"
	"willnorris.com/go/webmention"
)

var (
	client *webmention.Client
	url    string

	selector = flag.String("selector", ".h-entry", "CSS Selector limiting where to look for links")
)

func main() {
	flag.Parse()

	client = webmention.New(nil)
	url = flag.Arg(0)
	if url == "" {
		fatalf("Must provide url")
	}

	fmt.Printf("Searching for links from %q to send webmentions to...\n\n", url)
	dl, err := client.DiscoverLinks(url, *selector)
	if err != nil {
		fatalf("error discovering links for %q: %v", url, err)
	}
	var links []link
	for _, l := range dl {
		links = append(links, link{url: l})
	}

	selectLinks(links)
	sendWebmentions(links)
}

type link struct {
	url  string
	ping bool
}

func selectLinks(links []link) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Select links to send webmentions to:")
		for i, link := range links {
			x := " "
			if link.ping {
				x = "x"
			}
			fmt.Printf("  [%s]: %2d. %v\n", x, i, link.url)
		}

		fmt.Print("\nEnter space separated IDs of links to toggle, [a]ll or [n]one: ")
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		fmt.Println()

		switch input {
		case "":
			return
		case "a", "all":
			for i := range links {
				links[i].ping = true
			}
		case "n", "none":
			for i := range links {
				links[i].ping = false
			}
		default:
			for _, a := range strings.Split(input, " ") {
				i, err := strconv.Atoi(a)
				if err != nil || i > len(links) {
					continue
				}
				links[i].ping = !links[i].ping
			}
		}
	}
}

func sendWebmentions(links []link) {
	fmt.Println("Sending webmentions...")
	for _, l := range links {
		if !l.ping {
			continue
		}

		fmt.Printf("  %v ... ", l.url)
		endpoint, err := client.DiscoverEndpoint(l.url)
		if err != nil {
			errorf("%v", err)
			continue
		} else if endpoint == "" {
			color.Println("@{!r}no webmention support@|")
			continue
		}

		_, err = client.SendWebmention(endpoint, url, l.url)
		if err != nil {
			errorf("%v", err)
			continue
		}
		color.Println("@gsent@|")
	}
}

func fatalf(format string, args ...interface{}) {
	errorf(format, args...)
	os.Exit(1)
}

func errorf(format string, args ...interface{}) {
	color.Fprintf(os.Stderr, "@{!r}ERROR:@| ")
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
