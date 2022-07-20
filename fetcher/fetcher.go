package fetcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Constants

const REGEX_CASE_INSENSITIVE = "(?i)"

// Methods

func (fe *FetchError) Error() string {
	return fmt.Sprintf("Error: %s", "HackerNews data could not be fetched")
}

type Fetcher struct {
	Settings   Configuration
	repository DataRepository
	filters    []string
}

// Parse the filters configuration and return it as a flat array of strings
func (f *Fetcher) prepareFilters() []string {
	var resultFilters []string
	for _, filter := range f.Settings.Filters {
		filterStrings := strings.Split(filter.Value, ",")
		resultFilters = append(resultFilters, filterStrings...)
	}
	return resultFilters
}

// Get top stories' IDs
func (f *Fetcher) prefetch() (*[]int64, error) {
	result := make([]int64, 0)
	prefetchUrl := fmt.Sprintf("%s/topstories.json", f.Settings.ApiBaseUrl)
	resp, err := http.Get(prefetchUrl)
	if err != nil {
		return &result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return &result, err
	}
	return &result, nil
}

// Fetch one news item as a JSON object
func (f *Fetcher) fetchOne(id int64) (JsonNewsItem, error) {
	var result JsonNewsItem
	prefetchUrl := fmt.Sprintf("%s/item/%d.json", f.Settings.ApiBaseUrl, id)
	resp, err := http.Get(prefetchUrl)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Load IDs for news items that are already in the repository. For those prefetched IDs
// those that are not in the repository yet, fetch them one by one and run against
// the set of filters. For the `reverse`'d filters, the news item must be in none of them.
func (f *Fetcher) filter(prefetched *[]int64, reverse bool) ([]DigestItem, error) {
	digestItems := make([]DigestItem, 0)
	newItems := make([]DigestItem, 0)

	// Fetch existing items from the DB
	existingIDs, err := f.repository.GetExistingIDs()
	if err != nil {
		return Digest{}, err
	}
	// Fetch news items which do not exist in the DB
	for _, fetchId := range *prefetched {
		_, ok := existingIDs[fetchId]
		if ok {
			continue
		}
		// Fetch the item
		newItem, err := f.fetchOne(fetchId)
		if err != nil {
			log.Println("FETCH_ONE: ", err)
		}
		// Set a dumb URL and Title for items that don't have a URL
		if newItem.Url == "" {
			newItems = append(newItems, DigestItem{
				id:        newItem.Id,
				createdAt: newItem.Time,
				newsTitle: "-",
				newsUrl:   "-",
			})
		} else {
			// And now the valid items can be processed
			digestItem := DigestItem{
				id:        newItem.Id,
				createdAt: newItem.Time,
				newsTitle: newItem.Title,
				newsUrl:   newItem.Url,
			}
			newItems = append(newItems, digestItem)

			if f.RunFilter(&newItem, reverse) {
				digestItems = append(digestItems, digestItem)
			}
		}
	}
	// Add newly fetched items into the repository
	if len(newItems) > 0 {
		f.repository.UpdateItems(newItems)
	}

	return digestItems, nil
}

// Run a news item against all the configured filters
func (f *Fetcher) RunFilter(newItem *JsonNewsItem, reverse bool) bool {
	if reverse {
		anyFilterHit := false
		for _, filterItem := range f.filters {
			hit, _ := regexp.MatchString(REGEX_CASE_INSENSITIVE+filterItem, newItem.Title)
			if hit {
				anyFilterHit = true
				break
			}
		}
		if !anyFilterHit {
			return true
		}

	} else {
		for _, filterItem := range f.filters {
			hit, _ := regexp.MatchString(REGEX_CASE_INSENSITIVE+filterItem, newItem.Title)
			if hit {
				return true
			}
		}
	}
	return false
}

// Compile an email from the provided news list and send it
func (f *Fetcher) SendEmail(digest *[]DigestItem) {
	mailer := DigestMailer{smtpConfig: f.Settings.Smtp}
	mailer.SendEmail(digest, f.Settings.EmailTo)
}

// The main runner function
func (f *Fetcher) Run(reverseFilters ...bool) Results {
	reverse := false
	if len(reverseFilters) > 0 {
		reverse = reverseFilters[0]
	}

	f.filters = f.prepareFilters()
	f.repository = DataRepository{dbConfig: f.Settings.DatabaseFile, purgeAfter: f.Settings.PurgeAfterDays}
	f.repository.Init()
	defer f.repository.Close()

	prefetchedItems, err := f.prefetch()
	if err != nil {
		log.Fatal("PREFETCH: ", err)
	}
	digest, err := f.filter(prefetchedItems, reverse)
	if err != nil {
		log.Fatal("FILTER: ", err)
	}
	results := Results{
		NewItems: uint(len(digest)),
		Filters: uint(len(f.filters)),
	}
	if len(digest) > 0 {
		if f.Settings.EmailTo != "" {
			f.SendEmail(&digest)
		} else {
			for _, digestItem := range digest {
				fmt.Printf("* %s - %s"+CRLF, digestItem.newsTitle, digestItem.newsUrl)
			}
		}
	}
	return results
}
