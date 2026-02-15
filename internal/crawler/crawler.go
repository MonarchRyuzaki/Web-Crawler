package crawler

import (
	"WebCrawler/internal/core"
	"context"
	"fmt"
	"log"
)

type Crawler struct {
	fetcher core.Fetcher
	parser  core.Parser
	store   core.Storage
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
			save, err := c.store.CheckAndSave(ctx, url, article.TextContent)
			if err != nil {
				return err
			}
			if save {
				log.Printf("Saved Successfully to In Memory Store")
			}
		}
		break
	}
	return nil
}

func NewCrawler(fetcher core.Fetcher, parser core.Parser, store core.Storage) *Crawler {
	return &Crawler{
		fetcher: fetcher,
		parser:  parser,
		store:   store,
	}
}
