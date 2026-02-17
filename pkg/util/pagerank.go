package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// RankResult is what the caller receives.
type RankResult struct {
	Domain string
	Rank   int     // 0-10
	Score  float64 // The decimal score
	Error  error
}

// rankRequest is the internal message passed to the batcher.
type rankRequest struct {
	domain string
	respCh chan RankResult // The "return address" for the result
}

// BatchRanker is the service handling the logic.
type BatchRanker struct {
	apiKey    string
	client    *http.Client
	queue     chan rankRequest // Incoming requests land here
	done      chan struct{}    // For shutting down gracefully
	batchSize int

	cache map[string]RankResult
	mu    sync.Mutex
}

func NewBatchRanker(apiKey string) *BatchRanker {
	b := &BatchRanker{
		apiKey:    apiKey,
		client:    &http.Client{Timeout: 5 * time.Second},
		queue:     make(chan rankRequest, 1000), // Buffer for bursty traffic
		done:      make(chan struct{}),
		batchSize: 100, // API limit
		cache:     make(map[string]RankResult),
	}
	// Start the background worker immediately
	go b.startWorker()
	return b
}

func (b *BatchRanker) getRankResult(domain string) (RankResult, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	res, ok := b.cache[domain]
	return res, ok
}

func (b *BatchRanker) setRankResult(domain string, rankResult RankResult) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cache[domain] = rankResult
}

// GetRank queues the domain and returns a channel that will eventually hold the result.
// This is non-blocking.
func (b *BatchRanker) GetRank(domain string) <-chan RankResult {
	respCh := make(chan RankResult, 1) // Buffered so sender doesn't block
	req := rankRequest{
		domain: domain,
		respCh: respCh,
	}
	if res, ok := b.getRankResult(domain); ok {
		respCh <- res
	} else {
		select {
		case b.queue <- req:
			// Request queued successfully
		default:
			// Queue full - fail fast
			respCh <- RankResult{Error: fmt.Errorf("service overloaded")}
		}
	}
	return respCh
}

func (b *BatchRanker) startWorker() {
	ticker := time.NewTicker(1 * time.Second)
	var batch []rankRequest

	for {
		select {
		case req := <-b.queue:
			batch = append(batch, req)
			// Optimization: If we hit 100 items, fire immediately! Don't wait for ticker.
			if len(batch) >= b.batchSize {
				b.processBatch(batch)
				batch = nil                   // Reset batch
				ticker.Reset(1 * time.Second) // Reset timer so we don't fire twice
			}

		case <-ticker.C:
			if len(batch) > 0 {
				b.processBatch(batch)
				batch = nil
			}

		case <-b.done:
			return
		}
	}
}

func (b *BatchRanker) processBatch(requests []rankRequest) {
	// 1. Extract domains for API call
	domains := make([]string, len(requests))
	// Create a map to quickly find the waiting channel for a domain
	// Note: Handling duplicates (e.g., user asks for google.com twice) requires a slice map
	waiters := make(map[string][]chan RankResult)

	for i, req := range requests {
		domains[i] = req.domain
		waiters[req.domain] = append(waiters[req.domain], req.respCh)
	}

	// 2. Call the API
	fmt.Printf("Batch processing %d domains...\n", len(domains))
	apiResults, err := b.callOpenPageRank(domains)

	// 3. Handle Global Error (Network fail, etc)
	if err != nil {
		for _, req := range requests {
			req.respCh <- RankResult{Error: err}
		}
		return
	}

	// 4. Distribute Results
	// We loop through API results and notify waiting channels
	for _, res := range apiResults {
		waitingChans := waiters[res.Domain]
		for _, ch := range waitingChans {
			rankVal := 0
			if res.PageRankInteger != 0 {
				rankVal = res.PageRankInteger
			}
			result := RankResult{
				Domain: res.Domain,
				Rank:   rankVal,
				Score:  res.PageRankDecimal,
				Error:  nil,
			}

			b.cache[res.Domain] = result

			ch <- result
		}
		// Remove from map so we know who is handled
		delete(waiters, res.Domain)
	}

	// 5. Handle "Missing" Domains (API didn't return them)
	for domain, chans := range waiters {
		for _, ch := range chans {
			ch <- RankResult{Domain: domain, Error: fmt.Errorf("domain not found in API response")}
		}
	}
}

// --- Internal API Client Logic (Same as before, slightly adapted) ---

type apiResponse struct {
	Response []struct {
		Domain          string  `json:"domain"`
		PageRankInteger int     `json:"page_rank_integer"`
		PageRankDecimal float64 `json:"page_rank_decimal"`
	} `json:"response"`
}

func (b *BatchRanker) callOpenPageRank(domains []string) ([]struct {
	Domain          string  `json:"domain"`
	PageRankInteger int     `json:"page_rank_integer"`
	PageRankDecimal float64 `json:"page_rank_decimal"`
}, error) {

	u := "https://openpagerank.com/api/v1.0/getPageRank"
	params := url.Values{}
	for _, d := range domains {
		params.Add("domains[]", d)
	}

	req, _ := http.NewRequest("GET", u+"?"+params.Encode(), nil)
	req.Header.Set("API-OPR", b.apiKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Response, nil
}
