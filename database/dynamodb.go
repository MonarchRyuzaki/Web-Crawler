package database

import (
	"context"
	"log"
	"sync"

	appConfig "WebCrawler/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	dynamoClient *dynamodb.Client
	dynamoOnce   sync.Once
)

func GetDynamoClient(ctx context.Context, cfg *appConfig.Config) *dynamodb.Client {
	dynamoOnce.Do(func() {
		awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.AWSRegion))
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		var options []func(*dynamodb.Options)
		if cfg.DynamoEndpoint != "" {
			options = append(options, func(o *dynamodb.Options) {
				o.BaseEndpoint = aws.String(cfg.DynamoEndpoint)
			})
		}

		dynamoClient = dynamodb.NewFromConfig(awsCfg, options...)
	})
	return dynamoClient
}
