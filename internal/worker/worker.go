package worker

import (
	fetcher2 "WebCrawler/internal/fetcher"
	"bufio"
	"context"
	"fmt"
	_ "io"
)

type Worker struct{}

func (w *Worker) StartWorkers(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf("Attempting to Start Worker ID = %d\n", i)
		go func(id int) {
			fmt.Printf("[Worker %d]: Trying to Fetch\n", id)
			fetcher := fetcher2.FakeFetcher{}
			fetch, err := fetcher.Fetch(context.TODO(), "")
			if err != nil {
				return
			}
			defer fetch.Close()

			scanner := bufio.NewScanner(fetch)

			// Read line by line
			for scanner.Scan() {
				line := scanner.Text() // This only holds one line in memory at a time
				fmt.Printf("[Worker %d] processing line: %s\n", id, line)
			}

			if err := scanner.Err(); err != nil {
				fmt.Printf("Error during scan: %v\n", err)
			}

			// 3. Convert bytes to string for printing
			//fmt.Printf("[Worker %d]: %s\n", id, string(body))
		}(i)
	}
}
