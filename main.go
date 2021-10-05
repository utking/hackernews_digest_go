package main

import (
	newsFetcher "github.com/utking/hackernews_digest_go/fetcher"
	"log"
)

func main() {
	fetcher := newsFetcher.Fetcher{Settings: newsFetcher.GetConfig()}
	results := fetcher.Run()
	log.Printf("Filters: %d\nFetched new items: %d", results.Filters, results.NewItems)
}
