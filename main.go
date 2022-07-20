package main

import (
	"fmt"
	"os"

	newsFetcher "github.com/utking/hackernews_digest_go/fetcher"
)

func main() {
	reversedFilters := len(os.Args) > 1 && (os.Args[1] == "-r")
	fetcher := newsFetcher.Fetcher{Settings: newsFetcher.GetConfig()}
	results := fetcher.Run(reversedFilters)
	fmt.Printf("Filters: %d\nFetched new items: %d\n", results.Filters, results.NewItems)
}
