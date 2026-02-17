package crawler

import (
	"WebCrawler/internal/core"
	"WebCrawler/internal/extension"
	"context"
	"fmt"
	"log"
)

type Crawler struct {
	frontier core.Frontier
	fetcher  core.Fetcher
	parser   core.Parser
	store    core.Storage
}

func (c *Crawler) Start(ctx context.Context) error {
	//TODO implement me
	// 1. Get Seed URLs
	// 2. Put Seed URLs in Frontier
	// 3. Spawn multiple workers which get a item from Frontier
	log.Printf("Function Crawler.Start\n")
	cnt := 1
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down gracefully...")
			return ctx.Err()
		default:
			url, err := c.frontier.NextUrl(ctx)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}
			body, err := c.fetcher.Fetch(ctx, url)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}
			article, err := c.parser.Parse(body, url)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}
			_ = article
			body.Close()
			save, err := c.store.CheckAndSave(ctx, url, article.TextContent)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}
			if save {
				log.Printf("Saved Successfully to In Memory Store")
			}
			links := extension.UrlExtractor(article.Content)
			extension.CheckSaveAddUrlToFrontier(ctx, c.frontier, c.store, links)
		}
		if cnt >= 10 {
			break
		}
		cnt++
	}
	return nil
}

func NewCrawler(frontier core.Frontier, fetcher core.Fetcher, parser core.Parser, store core.Storage) *Crawler {
	return &Crawler{
		frontier: frontier,
		fetcher:  fetcher,
		parser:   parser,
		store:    store,
	}
}
