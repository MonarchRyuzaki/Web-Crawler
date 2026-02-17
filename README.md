# WebCrawler

A concurrent, polite web crawler built in Go, featuring a multi-queue frontier with PageRank-based prioritization, robots.txt compliance, content deduplication, and DNS caching.

## Architecture

![High Level Design](pics/High%20Level%20Design.png)

### Core Components

| Component | Interface | Description |
|-----------|-----------|-------------|
| **Frontier** | `core.Frontier` | URL scheduling with priority (front queues) and politeness (back queues) |
| **Fetcher** | `core.Fetcher` | HTTP client with DNS caching, connection pooling, and robots.txt enforcement |
| **Parser** | `core.Parser` | Extracts readable content and metadata from HTML using [go-readability](https://github.com/go-shiori/go-readability) |
| **Storage** | `core.Storage` | Content deduplication via SHA-256 hashing and visited-URL tracking |
| **Crawler** | `core.Crawler` | Main orchestrator that ties all components together |

### Frontier Design

![URL Frontier](pics/URL%20Frontier.png)

The frontier follows a classic two-stage queue architecture:

1. **Front Queues (Priority)** — 5 queues ranked by domain importance. URLs are assigned to queues based on their domain's [OpenPageRank](https://www.domcop.com/openpagerank/) score. Higher-ranked domains get dequeued more frequently via weighted random selection (exponential decay).

2. **Back Queues (Politeness)** — 5 queues that ensure the crawler doesn't overwhelm any single host. URLs from the same domain are consistently routed to the same back queue.

### Fetcher Features

- **robots.txt compliance** — Parses and respects `Disallow` rules per domain.
- **DNS caching** — Uses [dnscache](https://github.com/rs/dnscache) with periodic refresh to reduce DNS lookups.
- **Connection pooling** — Reuses TCP connections with configurable idle limits.
- **Configurable User-Agent** — Identifies itself as `ShiryuBot/1.0` by default.

### Content Processing

- HTML is parsed into clean, readable text using `go-readability`.
- Extracted links are filtered (HTTP/HTTPS only) and fed back into the frontier.
- Content is deduplicated using SHA-256 hashes before storage.

## Project Structure

```
.
├── main.go                     # Entry point
├── internal/
│   ├── seeder.go               # Seeds initial URLs into the frontier
│   ├── core/
│   │   └── interfaces.go       # Core interfaces (Frontier, Fetcher, Parser, Storage)
│   ├── crawler/
│   │   └── crawler.go          # Main crawl loop orchestrator
│   ├── extension/
│   │   └── link.go             # URL extraction, filtering, and frontier helpers
│   ├── fetcher/
│   │   ├── fetcher.go          # Production HTTP fetcher with robots.txt
│   │   └── fake-fetcher.go     # Mock fetcher for testing
│   ├── frontier/
│   │   └── frontier.go         # Priority + politeness queue system
│   ├── parser/
│   │   └── parser.go           # HTML content parser
│   ├── storage/
│   │   └── InMemoryStore.go    # In-memory visited set and content store
│   └── worker/
│       └── worker.go           # Worker pool (experimental)
└── pkg/
    └── util/
        ├── pagerank.go         # Batched OpenPageRank API client with caching
        ├── queue.go            # Channel-based queue primitive
        └── url.go              # URL parsing helpers
```

## Prerequisites

- **Go 1.25+**
- An [OpenPageRank API key](https://www.domcop.com/openpagerank/) (free tier available)

## Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd WebCrawler
   ```

2. **Create a `.env` file** in the project root:
   ```env
   OPEN_PAGE_RANK_API_KEY=your_api_key_here
   ```

3. **Install dependencies:**
   ```bash
   go mod download
   ```

4. **Run the crawler:**
   ```bash
   go run main.go
   ```

The crawler will start with the seed URL (`https://en.wikipedia.org/wiki/Web_crawler`) and crawl up to 10 pages before stopping. Press `Ctrl+C` for graceful shutdown.

## Configuration

Key parameters can be adjusted in the source:

| Parameter | Location | Default | Description |
|-----------|----------|---------|-------------|
| Seed URLs | `internal/seeder.go` | Wikipedia Web_crawler article | Starting URLs |
| Max pages | `internal/crawler/crawler.go` | 10 | Pages to crawl before stopping |
| HTTP timeout | `main.go` | 10s | Request timeout |
| User-Agent | `main.go` | `ShiryuBot/1.0` | Crawler identification |
| Front queues | `internal/frontier/frontier.go` | 5 | Number of priority queues |
| Back queues | `internal/frontier/frontier.go` | 5 | Number of politeness queues |

## Dependencies

- [go-readability](https://github.com/go-shiori/go-readability) — Content extraction from HTML
- [dnscache](https://github.com/rs/dnscache) — DNS response caching
- [godotenv](https://github.com/joho/godotenv) — Environment variable loading from `.env`
- [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html) — HTML tokenizer and parser

## Future Additions

- **Crawl Delay** — Respect `Crawl-delay` directives from `robots.txt` and introduce configurable per-host delays between requests to improve politeness and reduce the risk of being blocked.
- **Cloud-Native Storage** — Migrate from the current in-memory storage to a cloud-native architecture (e.g., distributed key-value stores, object storage for content, and managed databases for metadata) to support large-scale crawls and persistence across restarts.

## License

This project is for educational purposes.
