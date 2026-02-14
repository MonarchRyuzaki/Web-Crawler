package main

import (
	"WebCrawler/internal/crawler"
	"WebCrawler/internal/fetcher"
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	f := fetcher.NewFetcher(10*time.Second, "ShiryuBot/1.0")

	c := crawler.NewCrawler(f)
	err := c.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
