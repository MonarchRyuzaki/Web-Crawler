package main

import (
	"WebCrawler/internal/crawler"
	"context"
	"time"
)

func main() {
	c := crawler.Crawler{}
	err := c.Start(context.TODO())
	if err != nil {
		return
	}
	time.Sleep(1 * time.Millisecond)
}
