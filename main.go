package main

import (
	"WebCrawler/config"
	"WebCrawler/database"
	"WebCrawler/internal"
	"WebCrawler/internal/crawler"
	"WebCrawler/internal/fetcher"
	"WebCrawler/internal/frontier"
	"WebCrawler/internal/metrics"
	"WebCrawler/internal/parser"
	"WebCrawler/internal/storage"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db := database.GetDynamoClient(ctx, cfg)
	cache := database.GetRedisClient(cfg)

	err = cache.Ping(ctx).Err()
	if err != nil {
		fmt.Printf("❌ Redis error: %v\n", err)
	} else {
		fmt.Println("✅ Redis connected")
	}

	_, err = db.ListTables(ctx, &dynamodb.ListTablesInput{
		Limit: aws.Int32(1),
	})
	if err != nil {
		fmt.Printf("❌ DynamoDB error: %v\n", err)
	} else {
		fmt.Println("✅ DynamoDB connected")
	}

	fr := frontier.NewFrontier()
	f := fetcher.NewFetcher(10*time.Second, "ShiryuBot/1.0")
	p := parser.NewParser()
	//i := storage.NewInMemoryStore()
	i, err := storage.NewDynamoStore(ctx, db, cache)
	if err != nil {
		log.Fatal(err)
	}
	m := metrics.NewMetricsCollector()
	internal.SeedUrls(ctx, fr, i)

	c := crawler.NewCrawler(fr, f, p, i, m, cache)
	err = c.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
