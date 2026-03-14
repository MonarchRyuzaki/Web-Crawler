package publisher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/redis/go-redis/v9"
)

const articleQueueKey = "queue:articles"

type ArticleMessage struct {
	URL        string `json:"url"`
	HashString string `json:"content_hash"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Excerpt    string `json:"excerpt"`
	SiteName   string `json:"site_name"`
	Byline     string `json:"byline"`
	CrawledAt  string `json:"crawled_at"`
	Language   string `json:"language"`
}

func Publish(ctx context.Context, cache *redis.Client, url string, article readability.Article) {
	hash := sha256.Sum256([]byte(article.TextContent))
	msg := ArticleMessage{
		URL:        url,
		HashString: hex.EncodeToString(hash[:]),
		Title:      article.Title,
		Content:    article.TextContent,
		Excerpt:    article.Excerpt,
		SiteName:   article.SiteName,
		Byline:     article.Byline,
		CrawledAt:  time.Now().UTC().Format(time.RFC3339),
		Language:   article.Language,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("publisher: failed to marshal article for %s: %v\n", url, err)
		return
	}

	if err := cache.RPush(ctx, articleQueueKey, data).Err(); err != nil {
		fmt.Printf("publisher: failed to push article for %s: %v\n", url, err)
	}
}
