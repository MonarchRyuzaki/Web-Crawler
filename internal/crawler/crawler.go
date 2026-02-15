package crawler

import (
	"WebCrawler/internal/fetcher"
	"WebCrawler/internal/parser"
	"context"
	"fmt"
	"log"
)

type Crawler struct {
	fetcher *fetcher.Fetcher
	parser  *parser.MyParser
}

func (c *Crawler) Start(ctx context.Context) error {
	//TODO implement me
	// 1. Get Seed URLs
	// 2. Put Seed URLs in Frontier
	// 3. Spawn multiple workers which get a item from Frontier
	log.Printf("Function Crawler.Start\n")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down gracefully...")
			return ctx.Err()
		default:
			baseUrl := "https://en.wikipedia.org"
			path := "/wiki/Web_crawler"
			url := baseUrl + path
			body, err := c.fetcher.Fetch(ctx, url)
			if err != nil {
				return err
			}
			article, err := c.parser.Parse(body, url)
			if err != nil {
				return err
			}
			_ = article
			body.Close()
		}
		break
	}
	return nil
}

func NewCrawler(fetcher *fetcher.Fetcher, parser *parser.MyParser) *Crawler {
	return &Crawler{
		fetcher: fetcher,
		parser:  parser,
	}
}
