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

const RegexCaseInsensitive = "(?i)"

// Methods

func (fe *FetchError) Error() string {
	return fmt.Sprintf("Error: %s", "HackerNews data could not be fetched")
}

type Fetcher struct {
	filters    []string
	Settings   Configuration
	repository DataRepository
	Reverse    bool
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
	var result []int64

	prefetchURL := fmt.Sprintf("%s/topstories.json", f.Settings.ApiBaseUrl)
	request, _ := http.NewRequest(http.MethodGet, prefetchURL, http.NoBody)
	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		return &result, err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &result, err
	}

	return &result, nil
}

// Fetch one news item as a JSON object
func (f *Fetcher) fetchOne(id int64) (JsonNewsItem, error) {
	var result JsonNewsItem

	prefetchURL := fmt.Sprintf("%s/item/%d.json", f.Settings.ApiBaseUrl, id)
	resp, err := http.Get(prefetchURL)

	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

// Load IDs for news items that are already in the repository. For those prefetched IDs
// those that are not in the repository yet, fetch them one by one and run against
// the set of filters. For the reverse'd filters, the news item must be in none of them.
func (f *Fetcher) filter(prefetched *[]int64) ([]DigestItem, error) {
	var (
		newItems    []DigestItem
		digestItems []DigestItem
	)

	// Fetch existing items from the DB
	existingIDs, err := f.repository.GetExistingIDs()
	if err != nil {
		return Digest{}, err
	}

	// Fetch news items which do not exist in the DB
	for _, fetchID := range *prefetched {
		if _, ok := existingIDs[fetchID]; ok {
			continue
		}

		// Fetch the item
		newItem, err := f.fetchOne(fetchID)
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

			if f.RunFilter(&newItem) {
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
func (f *Fetcher) RunFilter(newItem *JsonNewsItem) bool {
	if f.Reverse {
		anyFilterHit := false

		for _, filterItem := range f.filters {
			hit, _ := regexp.MatchString(RegexCaseInsensitive+filterItem, newItem.Title)
			if hit {
				anyFilterHit = true
				break
			}
		}

		return !anyFilterHit
	}

	for _, filterItem := range f.filters {
		hit, _ := regexp.MatchString(RegexCaseInsensitive+filterItem, newItem.Title)
		if hit {
			return true
		}
	}

	return false
}

// Compile an email from the provided news list and send it
func (f *Fetcher) SendEmail(digest *[]DigestItem) {
	subjectPostfix := ""

	if f.Reverse {
		subjectPostfix = " Reversed"
	}

	mailer := DigestMailer{smtpConfig: f.Settings.Smtp}
	mailer.SendEmail(digest, f.Settings.EmailTo, f.Settings.Smtp.Subject+subjectPostfix)
}

func (f *Fetcher) Vacuum() {
	f.setUpRepository()
	f.repository.Close()
}

func (f *Fetcher) setUpRepository() {
	f.repository = DataRepository{dbConfig: f.Settings.Database, purgeAfter: f.Settings.PurgeAfterDays, reverse: f.Reverse}
	f.repository.Init()
}

// The main runner function
func (f *Fetcher) Run() (*Results, error) {
	f.filters = f.prepareFilters()
	f.setUpRepository()

	defer f.repository.Close()

	prefetchedItems, err := f.prefetch()
	if err != nil {
		return nil, err
	}

	digest, err := f.filter(prefetchedItems)
	if err != nil {
		return nil, err
	}

	results := Results{
		NewItems: len(digest),
		Filters:  len(f.filters),
	}

	if len(digest) > 0 {
		if f.Settings.EmailTo != "" {
			f.SendEmail(&digest)
		} else {
			// Print out to console
			for _, digestItem := range digest {
				fmt.Printf("* %s - %s"+CRLF, digestItem.newsTitle, digestItem.newsUrl)
			}
		}
	}

	return &results, nil
}
