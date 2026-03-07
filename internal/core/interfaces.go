package core

import (
	"context"
	"io"

	"github.com/go-shiori/go-readability"
)

// Crawler is the main engine that orchestrates the flow.
type Crawler interface {
	Start(ctx context.Context) error
}

// Fetcher is responsible for making the actual HTTP requests.
type Fetcher interface {
	// Fetch takes the absolute url and returns the body when status = 200
	Fetch(ctx context.Context, url string) (io.ReadCloser, error)
}

// Parser handles the raw content processing.
type Parser interface {
	Parse(content io.Reader, url string) (article readability.Article, err error)
}

// Storage handles persistence of content and deduplication.
// Implementation: Initially In-Memory Map, later DynamoDB + Bloom Filter.
type Storage interface {
	// Visited checks if we have already crawled this URL (Bloom Filter check)
	// In actual bloom filter, if the filter says maybe ie yes, check the db too in this function
	// The impl. for above comment is skipped for in Memory store
	Visited(ctx context.Context, url string) (bool, error)

	// CheckAndSave stores the crawled content and returns true if successfully saved and false if already present (DynamoDB)
	CheckAndSave(ctx context.Context, url string, article readability.Article) (bool, error)

	// MarkVisited explicitly adds a URL to the visited set (Bloom Filter add)
	MarkVisited(ctx context.Context, url string) error
}

// Frontier manages the scheduling of URLs (Priority + Politeness).
// Implementation: Initially Channels/Slices, later Redis Lists.
type Frontier interface {
	// AddUrl pushes a discovered link into the system
	AddUrl(url string) error

	// NextUrl retrieves the next URL that is "polite" to crawl right now
	// This blocks until a URL is ready.
	NextUrl(ctx context.Context) (string, error)
}

// Metrics is the collector for different type of metrics
type Metrics interface {
	IncPagesCrawled()
	IncCacheHits()
	GetPagesCrawled() int64
	GetCacheHits() int64
	PrintMetrics()
}
