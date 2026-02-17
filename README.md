# WebCrawler

A concurrent, polite web crawler built in Go, featuring a multi-queue frontier with PageRank-based prioritization, robots.txt compliance, content deduplication, and DNS caching. Built as an educational deep-dive into the internals of how production web crawlers work.

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

---

## Deep Dive

### URL Lifecycle

A URL flows through these stages before its content is stored:

1. **Seed / Discovery** — A URL is either a seed or extracted from a previously crawled page.
2. **URL Seen?** — Checked *before* crawling to avoid wasted work (e.g., page X → A and page Y → A would otherwise enqueue A twice). Uses an in-memory visited set.
3. **Frontier** — URL is prioritized (front queues) and assigned to a polite back queue.
4. **robots.txt Check** — Before fetching, the URL is validated against the domain's cached `robots.txt` rules.
5. **Fetch** — HTTP GET with DNS caching, timeouts, and connection pooling.
6. **Content Seen?** — The downloaded content is hashed (SHA-256) and compared against previously stored hashes to detect duplicates.
7. **Store & Extract** — Content is stored and links are extracted and fed back into step 1.

### Frontier Design

![URL Frontier](pics/URL%20Frontier.png)

The frontier follows a classic two-stage queue architecture:

**Front Queues (Priority)**
- 5 queues (FQ1–FQ5) ranked by domain importance.
- URLs are assigned based on their domain's [OpenPageRank](https://www.domcop.com/openpagerank/) score (0–10 scale).
- The **Front Queue Selector** is biased toward higher-priority queues using weighted random selection with exponential decay (e.g., FQ1 is selected far more often than FQ5).

**Back Queues (Politeness)**
- 5 back queues (scalable) that enforce per-host politeness.
- A **domain-to-queue map** ensures all URLs from the same domain are routed to the same queue.
- New domains are assigned to the back queue with the fewest items.
- The **Back Queue Selector** picks queues in a round-robin manner, ensuring no single host is hammered.

### Fetcher & robots.txt

The fetcher respects `robots.txt` — the "house rules" posted at the root of every website. On discovering a new domain, it fetches and caches `/robots.txt` once, then checks `Disallow` paths before every request.

**Why bother, even for short crawls?**

- **Avoids spider traps** — Infinite generated pages (e.g., `/events/2026/02`, `/events/2026/03`…) are usually disallowed. Without this, the crawler wastes its entire run on junk.
- **Skips junk paths** — Directories like `/cgi-bin/`, `/tmp/`, `/private/` are noise that wastes storage.
- **Avoids honeypots** — Some sites have hidden links specifically to catch bots that ignore `robots.txt`, resulting in permanent IP bans.

**Implementation:** Simple string prefix matching against cached `Disallow` paths per domain.

### DNS Caching

Every HTTP request requires a DNS lookup. Since a crawler revisits the same domains frequently, DNS responses are cached using [dnscache](https://github.com/rs/dnscache) with periodic refresh to avoid OS DNS lock contention and reduce latency.

### Content Deduplication

After downloading, page content must be checked for duplicates before storage. Several approaches were considered:

| Method | Bits | Verdict |
|--------|------|---------|
| CRC32 | 32 | Too many collisions |
| MD5 | 128 | Technically broken |
| SHA-256 | 256 | ✅ Used — strong, no practical collisions |
| SimHash | Variable | Good for near-duplicates, but requires scanning/indexing existing hashes — too complex for current scope |

The current implementation uses **SHA-256** for exact-match deduplication. Near-duplicate detection (SimHash) is deferred to a future iteration.

### Content Parsing

Raw HTML is not stored directly. The fetched page is passed through [go-readability](https://github.com/go-shiori/go-readability) to strip boilerplate (navbars, footers, ads) and extract the actual readable content before storage. It also handles **Relative → Absolute conversion** — e.g., `/wiki/Food` becomes `https://en.wikipedia.org/wiki/Food`

### URL Filtering & Link Extraction

The link extractor (`internal/extension/link.go`) handles:

- **Scheme filtering** — Only HTTP/HTTPS URLs are kept
- **Spider trap avoidance** — URLs with excessive length are discarded (e.g., `/foo/bar/foo/bar/foo/bar/…`)


---

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
- **Cloud-Native Storage** — Replace `MemoryStore` with DynamoDB for persistence and `MemoryQueue` with Redis (`RPUSH`/`LPOP`) for distributed queuing. Containerize with Docker to spin up multiple worker instances sharing the same Redis frontier.
- **Bloom Filter + DynamoDB for URL-Seen** — Two-tier URL deduplication: a Bloom filter in RAM as a fast first pass (if "no" → guaranteed new, skip DB read), with DynamoDB as the source of truth for false positives. This saves ~90% of DB reads since most discovered URLs are new.
- **Near-Duplicate Detection (SimHash)** — Move beyond exact-match SHA-256 to SimHash fingerprinting for detecting pages with near-identical content.
- **Sitemap Parsing** — Parse `sitemap.xml` from `robots.txt` to bulk-discover URLs in a single request instead of one-by-one link following, improving coverage of deep/orphan pages.
- **Smart Parameter Stripping** — Automatically detect and strip useless query parameters (session IDs, tracking params like `utm_source`) using heuristics: if the same URL path with different param values produces identical content hashes, flag that parameter as noise for the domain.
- **Metrics & Observability** — Real-time dashboard with crawl rate, queue depth, error counts, and per-domain statistics.
- **Edge Case Hardening** — Handle redirects (301/302), enforce `MaxBodySize` limits to skip oversized pages, and add retry logic with exponential backoff.

## Design Decisions & Trade-offs

| Decision | Reasoning |
|----------|-----------|
| **SHA-256 over SimHash** | SimHash is better for near-duplicates but requires comparing against all existing hashes (needs indexing). SHA-256 is a simple map lookup for exact matches — good enough for v1. |
| **URL-seen check before crawl** | If page X links to A and page Y also links to A, checking *after* fetch would waste a fetch. Checking *before* enqueuing avoids duplicate work. |
| **In-memory over DynamoDB** | Faster iteration during development. The interface-driven design means swapping to DynamoDB later requires zero changes to the crawler logic. |
| **Bloom filter consideration** | For large-scale crawls, a Bloom filter in front of the DB saves ~90% of read operations since most URLs are new. Deferred until the cloud-native phase. |
| **No sitemap parsing** | Standard link-following provides enough URLs for the current scope. Sitemap parsing is an optimization for *coverage*, not *correctness*. |

## License

This project is for educational purposes.
