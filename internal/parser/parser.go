package parser

import (
	"fmt"
	"io"
	"log"
	"net/url"

	"github.com/go-shiori/go-readability"
)

type MyParser struct {
}

func NewParser() *MyParser {
	return &MyParser{}
}

// parseContent Parses the messy HTML page into the fields and it also converts all the relative urls to absolute urls
func (m *MyParser) parseContent(content io.Reader, u string) (article readability.Article, err error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		log.Fatalf("error parsing url")
	}
	article, err = readability.FromReader(content, parsedUrl)
	if err != nil {
		log.Fatalf("failed to parse %s: %v\n", u, err)
	}

	fmt.Printf("URL     : %s\n", u)
	fmt.Printf("Title   : %s\n", article.Title)
	fmt.Printf("Author  : %s\n", article.Byline)
	fmt.Printf("Length  : %d\n", article.Length)
	fmt.Printf("Excerpt : %s\n", article.Excerpt)
	fmt.Printf("SiteName: %s\n", article.SiteName)
	fmt.Printf("Image   : %s\n", article.Image)
	fmt.Printf("Favicon : %s\n", article.Favicon)
	fmt.Printf("Content : %s\n", article.TextContent)
	fmt.Printf("HTMLContent: %s\n", article.Content)
	fmt.Println()
	return article, err
}

func (m *MyParser) Parse(content io.Reader, u string) (article readability.Article, err error) {
	article, err = m.parseContent(content, u)

	return article, err
}
