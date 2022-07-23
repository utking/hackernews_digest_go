package main

import (
	"fmt"
	"log"

	newsFetcher "github.com/utking/hackernews_digest_go/fetcher"
)

func main() {
	args := newsFetcher.ArgParser{}
	err := args.Parse()
	if err != nil {
		log.Fatalln(err)
	}
	config, err := newsFetcher.GetConfig(args.Config)
	if err != nil {
		log.Fatalln(err)
	}
	fetcher := newsFetcher.Fetcher{Settings: config, Reverse: args.Reverse}
	if args.Vacuum {
		fmt.Printf("Removing records older than %d days\n", config.PurgeAfterDays)
		fetcher.Vacuum()
	} else {
		results := fetcher.Run()
		fmt.Printf("Filters: %d\nFetched new items: %d\n", results.Filters, results.NewItems)
	}
}
