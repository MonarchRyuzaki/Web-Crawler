package crawler

import (
	"WebCrawler/internal/worker"
	"context"
	"fmt"
)

type Crawler struct{}

func (c *Crawler) Start(ctx context.Context) error {
	//TODO implement me
	// 1. Get Seed URLs
	// 2. Put Seed URLs in Frontier
	// 3. Spawn multiple workers which get a item from Frontier
	fmt.Printf("Function Crawler.Start\n")
	w := worker.Worker{}
	w.StartWorkers(5)
	return nil
}
