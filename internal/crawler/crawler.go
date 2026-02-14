package crawler

import (
	"WebCrawler/internal/fetcher"
	"bufio"
	"context"
	"fmt"
	"log"
)

type Crawler struct {
	fetcher *fetcher.Fetcher
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
			body, err := c.fetcher.Fetch(ctx, "https://en.wikipedia.org/wiki/Web_crawler")
			if err != nil {
				return err
			}
			scanner := bufio.NewScanner(body)

			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
			}
			body.Close()
		}
		break
	}
	return nil
}

func NewCrawler(fetcher *fetcher.Fetcher) *Crawler {
	return &Crawler{
		fetcher: fetcher,
	}
}
