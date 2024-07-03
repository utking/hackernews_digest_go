package fetcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slices"
)

// Constants

const RegexCaseInsensitive = "(?i)"

// Methods

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
func (f *Fetcher) filter(prefetched *[]int64) (*[]DigestItem, *[]DigestItem, error) {
	var (
		newItems    []DigestItem
		digestItems []DigestItem
	)

	// Find items to pull
	idsToPull, err := f.repository.GetIDsToPull(prefetched)
	if err != nil {
		return nil, nil, err
	}

	// Fetch news items which do not exist in the DB
	for _, fetchID := range idsToPull {
		// Fetch a new
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

			if f.filterItem(&newItem) && f.filterBlacklisted(&newItem) {
				digestItems = append(digestItems, digestItem)
			}
		}
	}

	return &newItems, &digestItems, nil
}

// Run a news item against the blacklisted domains
func (f *Fetcher) filterBlacklisted(newItem *JsonNewsItem) bool {
	if len(f.Settings.BlacklistedDomains) == 0 {
		// No blackist - nothing to check
		return true
	}

	parsedURL, err := url.Parse(newItem.Url)

	if err != nil {
		// Failed to parse the URL - not in the blacklist then
		return true
	}

	return !slices.Contains(f.Settings.BlacklistedDomains, parsedURL.Host)
}

// Run a news item against all the configured filters
func (f *Fetcher) filterItem(newItem *JsonNewsItem) bool {
	if f.Reverse {
		anyFilterHit := false

		for _, filter := range f.filters {
			if hit, _ := regexp.MatchString(RegexCaseInsensitive+filter, newItem.Title); hit {
				return false
			}
		}

		return !anyFilterHit
	}

	for _, filterItem := range f.filters {
		if hit, _ := regexp.MatchString(RegexCaseInsensitive+filterItem, newItem.Title); hit {
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

// Send to Telegram from the provided news list and send it
func (f *Fetcher) SendTelegram(digest *[]DigestItem) {
	telegram := DigestTelegram{tgConfig: f.Settings.Telegram}
	telegram.SendTelegram(digest, f.Settings.Telegram)
}

func (f *Fetcher) Vacuum() error {
	// Vacuum is part of the SetUp phase; so run it and exit
	if err := f.setUpRepository(); err != nil {
		return err
	}

	f.repository.Close()

	return nil
}

func (f *Fetcher) setUpRepository() error {
	f.repository = DataRepository{dbConfig: f.Settings.Database, purgeAfter: f.Settings.PurgeAfterDays, reverse: f.Reverse}
	return f.repository.Init()
}

// The main runner function
func (f *Fetcher) Run() (*Results, error) {
	f.filters = f.prepareFilters()

	if err := f.setUpRepository(); err != nil {
		return nil, err
	}

	defer f.repository.Close()

	prefetchedItems, err := f.prefetch()
	if err != nil {
		return nil, err
	}

	filteredItems, digest, err := f.filter(prefetchedItems)
	if err != nil {
		return nil, err
	}

	// Add newly fetched items into the repository
	if len(*filteredItems) > 0 {
		if err := f.repository.UpdateItems(filteredItems); err != nil {
			return nil, fmt.Errorf("could not update the repository")
		}
	}

	results := &Results{
		NewItems: len(*digest),
		Filters:  len(f.filters),
	}

	if len(*digest) > 0 {
		switch {
		case f.Settings.Telegram.Token != "" && f.Settings.Telegram.ChatId != "":
			f.SendTelegram(digest)
		case f.Settings.EmailTo != "":
			f.SendEmail(digest)
		default:
			// Print out to console
			for _, digestItem := range *digest {
				fmt.Printf("* %s - %s\n", digestItem.newsTitle, digestItem.newsUrl)
			}
		}
	}

	return results, nil
}
