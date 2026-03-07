package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-shiori/go-readability"
	"github.com/redis/go-redis/v9"
)

const (
	tableName        = "CrawledContent"
	pkAttribute      = "url"
	skAttribute      = "content_hash"
	bloomFilterKey   = "bf:visited_urls"
	bloomFilterError = 0.01 // 1% false positive rate
	bloomFilterCap   = 1_000_000
)

type DynamoStore struct {
	db    *dynamodb.Client
	cache *redis.Client
}

func NewDynamoStore(ctx context.Context, db *dynamodb.Client, cache *redis.Client) (*DynamoStore, error) {
	store := &DynamoStore{db: db, cache: cache}
	if err := store.initBloomFilter(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// initBloomFilter creates the Redis bloom filter if it doesn't already exist.
// BF.RESERVE is idempotent-safe via XX but here we just ignore "item exists" errors.
func (d *DynamoStore) initBloomFilter(ctx context.Context) error {
	err := d.cache.Do(ctx, "BF.RESERVE", bloomFilterKey, bloomFilterError, bloomFilterCap).Err()
	if err != nil && err.Error() != "ERR item exists" {
		return fmt.Errorf("failed to init bloom filter: %w", err)
	}
	return nil
}

// Visited checks:
// 1. Redis Bloom Filter — if false, URL is definitely new (no further checks needed)
// 2. DynamoDB — ground truth for positives (bloom filter may false-positive)
func (d *DynamoStore) Visited(ctx context.Context, url string) (bool, error) {
	// Layer 1: Redis Bloom Filter
	result, err := d.cache.Do(ctx, "BF.EXISTS", bloomFilterKey, url).Bool()
	if err != nil {
		return false, fmt.Errorf("bloom filter check failed: %w", err)
	}
	if !result {
		return false, nil // definitely not visited
	}

	// Layer 2: DynamoDB — confirm the bloom filter positive
	out, err := d.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			pkAttribute: &types.AttributeValueMemberS{Value: url},
		},
	})
	if err != nil {
		return false, fmt.Errorf("dynamodb GetItem failed: %w", err)
	}

	return out.Item != nil, nil
}

// CheckAndSave stores the page only if the content hash is new.
// Returns true if saved, false if duplicate.
func (d *DynamoStore) CheckAndSave(ctx context.Context, url string, article readability.Article) (bool, error) {
	hash := sha256.Sum256([]byte(article.TextContent))
	hashString := hex.EncodeToString(hash[:])

	_, err := d.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		ConditionExpression: aws.String("attribute_not_exists(#url)"),
		ExpressionAttributeNames: map[string]string{
			"#url": pkAttribute,
		},
		Item: map[string]types.AttributeValue{
			pkAttribute:  &types.AttributeValueMemberS{Value: url},
			skAttribute:  &types.AttributeValueMemberS{Value: hashString},
			"title":      &types.AttributeValueMemberS{Value: article.Title},
			"content":    &types.AttributeValueMemberS{Value: article.TextContent},
			"excerpt":    &types.AttributeValueMemberS{Value: article.Excerpt},
			"site_name":  &types.AttributeValueMemberS{Value: article.SiteName},
			"language":   &types.AttributeValueMemberS{Value: article.Language},
			"byline":     &types.AttributeValueMemberS{Value: article.Byline},
			"crawled_at": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})

	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return false, nil // already exists
		}
		return false, fmt.Errorf("dynamodb PutItem failed: %w", err)
	}

	return true, nil
}

// MarkVisited adds the URL to the Redis Bloom Filter.
// DynamoDB write is handled by CheckAndSave — no double write needed here.
func (d *DynamoStore) MarkVisited(ctx context.Context, url string) error {
	if err := d.cache.Do(ctx, "BF.ADD", bloomFilterKey, url).Err(); err != nil {
		return fmt.Errorf("bloom filter add failed: %w", err)
	}
	return nil
}
