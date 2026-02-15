package extractor

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// LinkExtractor takes an HTML string and returns a slice of all href values.
func LinkExtractor(htmlContent string) []string {
	var links []string

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && LinkFilter(attr.Val) {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return links
}

// LinkFilter only allows links which are http or https.
func LinkFilter(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	return u.Scheme == "https" || u.Scheme == "http"
}
