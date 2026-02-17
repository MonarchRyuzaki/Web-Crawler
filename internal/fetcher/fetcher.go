package fetcher

import (
	"WebCrawler/pkg/util"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/dnscache"
)

type Fetcher struct {
	client    *http.Client
	userAgent string
	// forbiddenPath stores domain as key and relative url as the values
	forbiddenPath map[string][]string

	mu sync.Mutex
}

func (f *Fetcher) getForbiddenPath(domain string) ([]string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	paths, ok := f.forbiddenPath[domain]
	return paths, ok
}

func (f *Fetcher) setForbiddenPath(domain string, paths []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.forbiddenPath[domain] = paths
}

func DNSResolver() *dnscache.Resolver {
	r := &dnscache.Resolver{}
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for range t.C {
			fmt.Println("[DNS] Refreshing cache from upstream...")
			r.Refresh(true)
		}
	}()
	return r
}

// NewFetcher creates a new Fetcher with safe defaults for a crawler.
// timeout: The hard limit for the entire request (DNS + Connect + Wait + Read)
func NewFetcher(timeout time.Duration, userAgent string) *Fetcher {
	r := DNSResolver()
	return &Fetcher{
		userAgent:     userAgent,
		forbiddenPath: make(map[string][]string),
		client: &http.Client{
			Timeout: timeout,
			// Connection Pooling
			Transport: &http.Transport{
				DisableKeepAlives: false,
				DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
					host, port, err := net.SplitHostPort(addr)
					if err != nil {
						return nil, err
					}
					ips, err := r.LookupHost(ctx, host)
					if err != nil {
						return nil, err
					}
					fmt.Printf("[DIAL] Attempting connection to %s using IPs %v\n", host, ips)
					for _, ip := range ips {
						var dialer net.Dialer
						conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
						if err == nil {
							return conn, nil
						}
					}
					return nil, fmt.Errorf("dial failed")
				},
				MaxIdleConns:        100, // Keep 100 connections open in the pool
				MaxIdleConnsPerHost: 10,  // Keep 10 open specifically for "wikipedia.org"
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// doRequest is the single source of truth for making HTTP calls.
func (f *Fetcher) doRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %v for %s", resp.StatusCode, url)
	}

	return resp.Body, nil
}

func (f *Fetcher) handleRobots(ctx context.Context, domain string) error {
	robotsURL := "https://" + domain + "/robots.txt"

	body, err := f.doRequest(ctx, robotsURL)
	if err != nil {
		return err
	}
	defer body.Close()

	var disallowed []string
	scanner := bufio.NewScanner(body)
	inTargetBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lowerLine := strings.ToLower(line)
		if strings.HasPrefix(lowerLine, "user-agent:") {
			agent := strings.TrimSpace(line[len("user-agent:"):])
			inTargetBlock = agent == "*" || agent == f.userAgent
			continue
		}

		if inTargetBlock && strings.HasPrefix(lowerLine, "disallow:") {
			path := strings.TrimSpace(line[len("disallow:"):])
			if path != "" {
				disallowed = append(disallowed, path)
			}
		}
	}

	f.setForbiddenPath(domain, disallowed)
	log.Printf("Domain %v has the disallowed paths : %v\n", domain, disallowed)
	return scanner.Err()
}

func (f *Fetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	domain, _ := util.GetDomain(url)

	if _, ok := f.getForbiddenPath(domain); !ok {
		if err := f.handleRobots(ctx, domain); err != nil {
			// Log error but perhaps continue if robots.txt is missing (404)
			// For now, we follow your original logic of returning the error
			return nil, err
		}
	}

	currentPath, err := util.GetPath(url)
	if err != nil {
		return nil, fmt.Errorf("could not parse path: %w", err)
	}

	fp, _ := f.getForbiddenPath(domain)

	for _, forbidden := range fp {
		// Standard robots.txt behavior: "Disallow: /tmp" matches "/tmp", "/tmp/", and "/tmp/file.html"
		if strings.HasPrefix(currentPath, forbidden) {
			return nil, fmt.Errorf("access denied by robots.txt: %s", url)
		}
	}

	return f.doRequest(ctx, url)
}
