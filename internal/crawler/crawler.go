package crawler

import (
	"WebCrawler/internal/core"
	"WebCrawler/internal/extension"
	"context"
	"fmt"
	"io"
	"log"
	"time"
)

type Crawler struct {
	frontier core.Frontier
	fetcher  core.Fetcher
	parser   core.Parser
	store    core.Storage
}

type fetchResult struct {
	url  string
	body io.ReadCloser
}

const NUMBER_OF_FETCHER_WORKERS = 5
const NUMBER_OF_PROCESSING_WORKERS = 5

func (c *Crawler) Start(ctx context.Context) error {
	log.Printf("Function Crawler.Start\n")
	workStream := make(chan fetchResult, 50)
	for i := 0; i < NUMBER_OF_FETCHER_WORKERS; i++ {
		go func() {
			for {
				time.Sleep(10 * time.Millisecond)
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
				workStream <- fetchResult{url: url, body: body}
			}
		}()
	}
	for i := 0; i < NUMBER_OF_PROCESSING_WORKERS; i++ {
		go func() {
			for res := range workStream {
				article, err := c.parser.Parse(res.body, res.url)
				if err != nil {
					fmt.Printf("%v", err)
					continue
				}
				err = res.body.Close()
				if err != nil {
					fmt.Printf("%v", err)
					continue
				}
				save, err := c.store.CheckAndSave(ctx, res.url, article.TextContent)
				if err != nil {
					fmt.Printf("%v", err)
					continue
				}
				if save {
					links := extension.UrlExtractor(article.Content)
					extension.CheckSaveAddUrlToFrontier(ctx, c.frontier, c.store, links)
				}
			}
		}()
	}
	<-ctx.Done()
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
