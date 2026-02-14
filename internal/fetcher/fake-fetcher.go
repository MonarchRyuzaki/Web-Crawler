package fetcher

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type FakeFetcher struct{}

func (f FakeFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	fmt.Printf("FakeFetcher.Fetch\n")
	// 1. The original string
	myString := "hello world"

	// 2. Create an io.Reader from the string
	stringReader := strings.NewReader(myString)
	// The type of stringReader is *strings.Reader, which implements io.Reader.

	// 3. Wrap the io.Reader with io.NopCloser to create an io.ReadCloser
	readCloser := io.NopCloser(stringReader)
	// The type of readCloser is io.ReadCloser.

	return readCloser, nil
}
