package extension

import (
	"WebCrawler/internal/core"
	"context"
	"log"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// UrlExtractor takes an HTML string and returns a slice of all href values.
func UrlExtractor(htmlContent string) []string {
	var links []string

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && UrlFilter(attr.Val) {
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

// UrlFilter only allows links which are http or https.
func UrlFilter(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	return (u.Scheme == "https" || u.Scheme == "http") && len(u.String()) <= 300
}

func CheckSaveAddUrlToFrontier(ctx context.Context, frontier core.Frontier, store core.Storage, links []string) {
	for _, link := range links {
		if ctx.Err() != nil {
			return
		}
		isVisited, err := store.Visited(ctx, link)
		if err != nil {
			log.Printf("Could Not Check URL")
			continue
		}
		if isVisited {
			continue
		}
		err = store.MarkVisited(ctx, link)
		if err != nil {
			log.Printf("Could Not Save URL")
			continue
		}

		go func() {
			err = frontier.AddUrl(link)
			if err != nil {
				return
			}
		}()
	}

}
