package core

import (
	"context"
	"io"
)

// Crawler is the main engine that orchestrates the flow.
type Crawler interface {
	Start(ctx context.Context) error
}

// Fetcher is responsible for making the actual HTTP requests.
// Implementation: Initially standard net/http, later with Rotated Proxies.
type Fetcher interface {
	Fetch(ctx context.Context, url string) (io.ReadCloser, error)
}

// Parser handles the raw content processing.
// Implementation: html.Parse to extract links and text.
type Parser interface {
	Parse(content io.Reader, baseUrl string) (links []string, text string, err error)
}

// Storage handles persistence of content and deduplication.
// Implementation: Initially In-Memory Map, later DynamoDB + Bloom Filter.
type Storage interface {
	// Visited checks if we have already crawled this URL (Bloom Filter check)
	Visited(ctx context.Context, url string) (bool, error)

	// Save stores the crawled content (DynamoDB)
	Save(ctx context.Context, url string, content string) error

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
