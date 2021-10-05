package main

import (
	newsFetcher "awesomeProject/fetcher"
	"log"
)

func main() {
	fetcher := newsFetcher.Fetcher{Settings: newsFetcher.GetConfig()}
	results := fetcher.Run()
	log.Printf("Filters: %d\nFetched new items: %d", results.Filters, results.NewItems)
}