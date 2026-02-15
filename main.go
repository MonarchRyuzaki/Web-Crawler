package main

import (
	"WebCrawler/internal/crawler"
	"WebCrawler/internal/fetcher"
	"WebCrawler/internal/parser"
	"WebCrawler/internal/storage"
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
	p := parser.NewParser()
	i := storage.NewInMemoryStore()

	c := crawler.NewCrawler(f, p, i)
	err := c.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
