package main

import (
	"fmt"
	newsFetcher "github.com/utking/hackernews_digest_go/fetcher"
)

func main() {
	fetcher := newsFetcher.Fetcher{Settings: newsFetcher.GetConfig()}
	results := fetcher.Run()
	fmt.Printf("Filters: %d\nFetched new items: %d\n", results.Filters, results.NewItems)
}
