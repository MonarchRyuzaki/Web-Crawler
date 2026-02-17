package main

import (
	"WebCrawler/internal"
	"WebCrawler/internal/crawler"
	"WebCrawler/internal/fetcher"
	"WebCrawler/internal/frontier"
	"WebCrawler/internal/parser"
	"WebCrawler/internal/storage"
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fr := frontier.NewFrontier()
	f := fetcher.NewFetcher(10*time.Second, "ShiryuBot/1.0")
	p := parser.NewParser()
	i := storage.NewInMemoryStore()
	internal.SeedUrls(ctx, fr, i)

	c := crawler.NewCrawler(fr, f, p, i)
	err = c.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
