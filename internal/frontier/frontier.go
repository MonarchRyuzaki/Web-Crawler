package frontier

import (
	"WebCrawler/pkg/util"
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"sync"
	"time"
)

type Frontier struct {
	frontQueue     []util.Queue
	backQueue      []util.Queue
	pageRankClient *util.BatchRanker
	tempNextUrl    util.Queue

	backQueueRoutingTable map[string]int
	mu                    sync.Mutex
}

const NUMBER_OF_FRONT_QUEUE = 5
const NUMBER_OF_BACK_QUEUE = 5

func NewFrontier() *Frontier {
	f := Frontier{
		frontQueue:            make([]util.Queue, NUMBER_OF_FRONT_QUEUE),
		backQueue:             make([]util.Queue, NUMBER_OF_BACK_QUEUE),
		backQueueRoutingTable: make(map[string]int),
		pageRankClient:        util.NewBatchRanker(os.Getenv("OPEN_PAGE_RANK_API_KEY")),
		tempNextUrl:           make(util.Queue, 1000),
	}
	for i := range NUMBER_OF_FRONT_QUEUE {
		f.frontQueue[i] = make(util.Queue, 1000)
	}
	for i := range NUMBER_OF_BACK_QUEUE {
		f.backQueue[i] = make(util.Queue, 1000)
	}
	go f.frontQueueSelectorDequeue()
	go f.backQueueSelector()
	return &f
}

func (f *Frontier) prioritizer(url string) (int, error) {
	domain, err := util.GetDomain(url)
	if err != nil {
		return -1, err
	}
	future := f.pageRankClient.GetRank(domain)

	result := <-future

	if result.Error != nil {
		fmt.Printf("Failed %s: %v\n", url, result.Error)
		return NUMBER_OF_FRONT_QUEUE - 1, nil
	} else {
		//fmt.Printf("✅ Got Result! %s => Rank: %d (Score: %.2f)\n", result.Domain, result.Rank, result.Score)
	}

	if result.Score >= 10.0 {
		return 0, nil
	}
	if result.Score <= 0.0 {
		return NUMBER_OF_FRONT_QUEUE - 1, nil
	}

	// The Formula
	// Example: Rank 8.5, Queues 5
	// (10 - 8.5) / 10 = 0.15
	// 0.15 * 5 = 0.75
	// floor(0.75) = 0
	index := int(math.Floor(((10.0 - result.Score) / 10.0) * float64(NUMBER_OF_FRONT_QUEUE)))

	if index >= NUMBER_OF_FRONT_QUEUE {
		return NUMBER_OF_FRONT_QUEUE - 1, nil
	}
	return index, nil
}

func (f *Frontier) AddUrl(url string) error {
	fqIdx, err := f.prioritizer(url)
	if err != nil {
		return err
	}
	f.frontQueue[fqIdx].Enqueue(url)
	return nil
}

func (f *Frontier) frontQueueSelectorDequeue() {
	// Generate weights dynamically using a decay factor (e.g., 0.5)
	// Weight = factor ^ index
	for {
		time.Sleep(100 * time.Millisecond)
		weights := make([]float64, NUMBER_OF_FRONT_QUEUE)
		totalWeight := 0.0
		factor := 0.5

		for i := 0; i < NUMBER_OF_FRONT_QUEUE; i++ {
			weights[i] = math.Pow(factor, float64(i))
			totalWeight += weights[i]
		}

		roll := rand.Float64() * totalWeight
		selectedIndex := -1
		cursor := 0.0
		for i, w := range weights {
			cursor += w
			if roll < cursor {
				selectedIndex = i
				break
			}
		}

		url := f.frontQueue[selectedIndex].Dequeue()
		if url != "" {
			//fmt.Printf("Dequeue from Front Queue #%v : %v\n", selectedIndex, url)
			f.backQueueRouter(url)
		}
	}
}

func (f *Frontier) backQueueRouter(url string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	idx, ok := f.backQueueRoutingTable[url]
	if !ok {
		idx = rand.IntN(NUMBER_OF_BACK_QUEUE)
		f.backQueueRoutingTable[url] = idx
	}

	f.backQueue[idx].Enqueue(url)
	//fmt.Printf("Enqueue to Back Queue #%v : %v\n", idx, url)
}

func (f *Frontier) backQueueSelector() {
	for {
		time.Sleep(100 * time.Millisecond)
		for i := range NUMBER_OF_BACK_QUEUE {
			url := f.backQueue[i].Dequeue()
			if url != "" {
				f.tempNextUrl <- url
			}

		}
	}
}

func (f *Frontier) NextUrl(ctx context.Context) (string, error) {
	return <-f.tempNextUrl, nil
}
