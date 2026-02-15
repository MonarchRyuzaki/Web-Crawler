package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type InMemoryStore struct {
	// contentStore : url => Content
	contentStore map[string]string

	// urlVisited : url => true/false
	urlVisited map[string]bool

	// contentHashes : hash(url) => url (for debug purpose)
	contentHashes map[string]string

	mu sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		contentStore:  make(map[string]string),
		urlVisited:    make(map[string]bool),
		contentHashes: make(map[string]string),
	}
}

func (i *InMemoryStore) Visited(ctx context.Context, url string) (bool, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	isPresent := i.urlVisited[url]
	return isPresent, nil
}

func (i *InMemoryStore) CheckAndSave(ctx context.Context, url string, content string) (bool, error) {
	hash := sha256.Sum256([]byte(content))

	hashString := hex.EncodeToString(hash[:])

	i.mu.Lock()
	defer i.mu.Unlock()

	if _, exists := i.contentHashes[hashString]; exists {
		return false, nil
	}

	i.contentHashes[hashString] = url
	i.contentStore[url] = content

	return true, nil
}

func (i *InMemoryStore) MarkVisited(ctx context.Context, url string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.urlVisited[url] = true
	return nil
}
