package metrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

// MetricsCollector is the concrete implementation
type MetricsCollector struct {
	pagesCrawled int64
	cacheHits    int64
	startTime    time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		pagesCrawled: 0,
		cacheHits:    0,
		startTime:    time.Now(),
	}
}

func (m *MetricsCollector) IncPagesCrawled() {
	atomic.AddInt64(&m.pagesCrawled, 1)
}

func (m *MetricsCollector) IncCacheHits() {
	atomic.AddInt64(&m.cacheHits, 1)
}

func (m *MetricsCollector) GetPagesCrawled() int64 {
	return atomic.LoadInt64(&m.pagesCrawled)
}

func (m *MetricsCollector) GetCacheHits() int64 {
	return atomic.LoadInt64(&m.cacheHits)
}

func (m *MetricsCollector) PrintMetrics() {
	crawled := m.GetPagesCrawled()
	hits := m.GetCacheHits()
	duration := time.Now().Sub(m.startTime)

	hitRate := 0.0
	if crawled > 0 {
		hitRate = float64(hits) / float64(crawled) * 100
	}

	crawlRate := float64(crawled) / duration.Seconds()

	fmt.Printf("Pages Crawled: %d\n", crawled)
	fmt.Printf("Pages Crawled per Second: %.1f\n", crawlRate)
	fmt.Printf("Cache Hits:    %d (%.1f%%)\n", hits, hitRate)
}
