package main

import (
	"WebCrawler/internal/crawler"
	"WebCrawler/internal/fetcher"
	"WebCrawler/internal/parser"
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

	c := crawler.NewCrawler(f, p)
	err := c.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
