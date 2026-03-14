package internal

import (
	"WebCrawler/internal/core"
	"WebCrawler/internal/extension"
	"context"
)

func SeedUrls(ctx context.Context, frontier core.Frontier, store core.Storage) {
	links := []string{"https://en.wikipedia.org/wiki/Dog"}
	extension.CheckSaveAddUrlToFrontier(ctx, frontier, store, links)
}
