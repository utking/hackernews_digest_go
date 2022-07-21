package main

import (
	"fmt"
	"log"

	newsFetcher "github.com/utking/hackernews_digest_go/fetcher"
)

func main() {
	config, err := newsFetcher.GetConfig()
	if err != nil {
		log.Fatalln(err)
	}
	fetcher := newsFetcher.Fetcher{Settings: config}
	results := fetcher.Run()
	fmt.Printf("Filters: %d\nFetched new items: %d\n", results.Filters, results.NewItems)
}
