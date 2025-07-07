package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	newsFetcher "github.com/utking/hackernews_digest_go/fetcher"
)

var (
	err     error
	config  newsFetcher.Configuration
	results *newsFetcher.Results
)

func main() {
	args := newsFetcher.ArgParser{}
	cwd := "."

	if err = args.Parse(); err != nil {
		log.Fatalln(err)
	}

	if len(args.Config) > 0 && args.Config[0] != '/' {
		if cwd, err = os.Getwd(); err != nil {
			log.Fatalln("Cannot find what directory we are in")
		}
	}

	if config, err = newsFetcher.GetConfig(filepath.Join(cwd, args.Config)); err != nil {
		log.Fatalln(err)
	}

	fetcher := newsFetcher.Fetcher{Settings: config, Reverse: args.Reverse}

	if args.Vacuum {
		fmt.Printf("Removing records older than %d days\n", config.PurgeAfterDays)
		if err = fetcher.Vacuum(); err != nil {
			log.Fatalln(err)
		}

		return
	}

	if results, err = fetcher.Run(); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Filters: %d\nFetched new items: %d\n", results.Filters, results.NewItems)
}
