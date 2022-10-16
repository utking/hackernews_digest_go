package fetcher

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
)

const (
	Hit  = true
	Miss = false
)

func prepareItem() JsonNewsItem {
	item := JsonNewsItem{
		Id:    1,
		Title: "Some Title",
		Url:   "http://some.url/abc",
	}

	return item
}

func TestRunFilterHit(t *testing.T) {
	item := prepareItem()
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "HitTest",
				Value: "Title",
			},
		},
	}
	fetcher := Fetcher{Settings: config, Reverse: false}
	fetcher.filters = fetcher.prepareFilters()

	if len(fetcher.filters) != 1 {
		t.Fatal("Wrong number of filters")
	}

	if fetcher.filterItem(&item) == Miss {
		t.Fatal("Filter missed")
	}
}

func TestRunFilterMiss(t *testing.T) {
	item := prepareItem()
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "MissTest",
				Value: "News Header",
			},
		},
	}
	fetcher := Fetcher{Settings: config, Reverse: false}
	fetcher.filters = fetcher.prepareFilters()

	if len(fetcher.filters) != 1 {
		t.Fatal("Wrong number of filters")
	}

	if fetcher.filterItem(&item) != Miss {
		t.Fatal("Reverse filter did not miss")
	}
}

func TestReverseFilterHit(t *testing.T) {
	item := prepareItem()
	fetcher := Fetcher{Settings: Configuration{
		Filters: []FilterItem{
			{
				Title: "HitTest",
				Value: "Header",
			},
		},
	}, Reverse: true}
	fetcher.filters = fetcher.prepareFilters()

	if len(fetcher.filters) != 1 {
		t.Fatal("Wrong number of filters", len(fetcher.filters))
	}

	if fetcher.filterItem(&item) != Hit {
		t.Fatal("Reverse filter missed")
	}
}

func TestReverseFilterMiss(t *testing.T) {
	item := prepareItem()
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "MissTest",
				Value: "Title",
			},
		},
	}
	fetcher := Fetcher{Settings: config, Reverse: true}
	fetcher.filters = fetcher.prepareFilters()

	if len(fetcher.filters) != 1 {
		t.Fatal("Wrong number of filters")
	}

	if fetcher.filterItem(&item) != Miss {
		t.Fatal("Reverse filter did not miss")
	}
}

func TestPrepareFilters(t *testing.T) {
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "MissTest",
				Value: "some,word",
			},
		},
	}
	fetcher := Fetcher{Settings: config}
	fetcher.filters = fetcher.prepareFilters()

	if len(fetcher.filters) != 2 {
		t.Fatal("Wrong number of filters")
	}

	if fetcher.filters[0] != "some" || fetcher.filters[1] != "word" {
		t.Fatal("Wrong filter values")
	}
}

func TestPrefetch(t *testing.T) {
	expected := "[33214439,33215770]"
	fetcher := Fetcher{Settings: Configuration{ApiBaseUrl: ""}}

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fetcher.Settings.ApiBaseUrl+"/topstories.json",
		httpmock.NewStringResponder(200, expected))

	value, _ := fetcher.prefetch()

	if httpmock.GetTotalCallCount() < 1 {
		t.Errorf("Expected a request while prefetching news items")
	}

	if len(*value) != 2 {
		t.Errorf("Expected 2 prefetched items, got %d", len(*value))
	}

	if (*value)[0] != 33214439 || (*value)[1] != 33215770 {
		t.Errorf("Expected %s prefetched items, got %v", expected, *value)
	}
}

func TestFetchOne(t *testing.T) {
	newsID := int64(33214439)
	expected := `{
		"by": "endorphine",
		"descendants": 1,
		"id": 33214439,
		"kids": [
			33216431
		],
		"score": 289,
		"time": 1665839339,
		"title": "A 24-year-old bug in the Linux Kernel TCP stack (2021)",
		"type": "story",
		"url": "https://engineering.skroutz.gr/blog/uncovering-a-24-year-old-bug-in-the-linux-kernel/"
	}`
	fetcher := Fetcher{Settings: Configuration{ApiBaseUrl: ""}}

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fetcher.Settings.ApiBaseUrl+"/item/33214439.json",
		httpmock.NewStringResponder(200, expected))

	item, err := fetcher.fetchOne(newsID)

	if err != nil {
		t.Error(err)
	}

	if httpmock.GetTotalCallCount() < 1 {
		t.Errorf("Expected a request while prefetching news items")
	}

	if item.Id != newsID {
		t.Errorf("Expected ID to be %d, got %d", newsID, item.Id)
	}

	if item.Title != "A 24-year-old bug in the Linux Kernel TCP stack (2021)" {
		t.Errorf("Expected ID to be %s, got %s", "A 24-year-old bug in the Linux Kernel TCP stack (2021)", item.Title)
	}

	if item.Url != "https://engineering.skroutz.gr/blog/uncovering-a-24-year-old-bug-in-the-linux-kernel/" {
		t.Errorf("Expected URL to be '%s', got '%s'", "https://engineering.skroutz.gr/blog/uncovering-a-24-year-old-bug-in-the-linux-kernel/", item.Url)
	}
}

func TestFetchOneBroken(t *testing.T) {
	newsID := int64(33214439)
	expected := `some-response`
	fetcher := Fetcher{Settings: Configuration{ApiBaseUrl: ""}}

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fetcher.Settings.ApiBaseUrl+"/item/33214439.json",
		httpmock.NewStringResponder(200, expected))

	item, err := fetcher.fetchOne(newsID)

	if err == nil {
		t.Error("Expected to fail a wrong response parsing")
	}

	if httpmock.GetTotalCallCount() < 1 {
		t.Errorf("Expected a request while prefetching news items")
	}

	if item.Id != 0 {
		t.Errorf("Expected ID to be %d, got %d", 0, item.Id)
	}
}

func TestVacuum(t *testing.T) {
	fetcher := Fetcher{Settings: Configuration{ApiBaseUrl: "", Database: Database{Driver: "sqlite3", Database: ":memory:"}}}

	err := fetcher.Vacuum()

	if err != nil {
		t.Errorf("Vacuum should not throw any errors, %v", err)
	}
}

func TestFilterItems(t *testing.T) {
	expectedPrefetched := "[33214440,33215770,33215771]"
	oldNewsItemID := int64(33214440)
	expectedDigest := map[int64]string{
		33214440: `{
			"by": "author",
			"descendants": 1,
			"id": 33214440,
			"kids": [
				33216444
			],
			"score": 10,
			"time": 1665839487,
			"title": "Some Title #1",
			"type": "story",
			"url": "https://some-host/1-story-uri/"
		}`,
		33215770: `{
			"by": "author-2",
			"descendants": 1,
			"id": 33215770,
			"kids": [
				33215771
			],
			"score": 20,
			"time": 1665839234,
			"title": "Some Title #2",
			"type": "story",
			"url": "https://some-host/2-story-uri/"
		}`,
		33215771: `{
			"by": "author-3",
			"descendants": 1,
			"id": 33215771,
			"kids": [
				33215772
			],
			"score": 21,
			"time": 1665839235,
			"title": "Some News",
			"type": "story",
			"url": ""
		}`,
	}

	fetcher := Fetcher{Settings: Configuration{
		ApiBaseUrl: "",
		Filters:    []FilterItem{{Title: "Test filter", Value: "title"}},
		Database:   Database{Driver: "sqlite3", Database: ":memory:"},
	}}

	fetcher.filters = fetcher.prepareFilters()

	if err := fetcher.setUpRepository(); err != nil {
		t.Errorf("Errof while initializing the repository, %v", err)
	}

	// Add "old" news items to the repository
	err := fetcher.repository.UpdateItems(&[]DigestItem{
		{
			newsTitle: "Existing Title",
			newsUrl:   "http://host",
			id:        oldNewsItemID,
		},
	})

	if err != nil {
		t.Errorf("Counld not add an old news item")
	}

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	// Mock the top stories
	httpmock.RegisterResponder("GET", fetcher.Settings.ApiBaseUrl+"/topstories.json",
		httpmock.NewStringResponder(200, expectedPrefetched))
	// Mock stories
	for id, resp := range expectedDigest {
		httpmock.RegisterResponder("GET", fmt.Sprintf(fetcher.Settings.ApiBaseUrl+"/item/%d.json", id),
			httpmock.NewStringResponder(200, resp))
	}

	prefetched, err := fetcher.prefetch()

	if err != nil {
		t.Errorf("Error while prefetching news items, %v", err)
	}

	if len(*prefetched) != 3 {
		t.Errorf("Expected 3 prefetched items, got %d", len(*prefetched))
	}

	if (*prefetched)[0] != 33214440 || (*prefetched)[1] != 33215770 || (*prefetched)[2] != 33215771 {
		t.Errorf("Expected %s prefetched items, got %v", expectedPrefetched, *prefetched)
	}

	unfiltered, filtered, err := fetcher.filter(prefetched)

	if err != nil {
		t.Errorf("Error while filtering news items, %v", err)
	}

	if unfiltered == nil || filtered == nil {
		t.Errorf("filtered and digest must not be nil")
	} else if len(*unfiltered) != 2 {
		t.Errorf("Expected 2 filtered items, got %d", len(*unfiltered))
	}

	if len(*filtered) != 1 {
		t.Fatalf("Expected 1 digest item, got %d", len(*filtered))
	}

	if (*filtered)[0].id != 33215770 {
		t.Errorf("Expected a news item with ID %d, got %d", 33215770, (*filtered)[0].id)
	}
}

func TestBlacklistedDomains(t *testing.T) {
	expectedPrefetched := "[33214440,33215770,33215771]"
	expectedDigest := map[int64]string{
		33214440: `{
			"by": "author",
			"descendants": 1,
			"id": 33214440,
			"kids": [
				33216444
			],
			"score": 10,
			"time": 1665839487,
			"title": "Some Title #1",
			"type": "story",
			"url": "https://somehost/1-story-uri/"
		}`,
		33215770: `{
			"by": "author-2",
			"descendants": 1,
			"id": 33215770,
			"kids": [
				33215771
			],
			"score": 20,
			"time": 1665839234,
			"title": "Some Title #2",
			"type": "story",
			"url": "https://some-host/2-story-uri/"
		}`,
		33215771: `{
			"by": "author-3",
			"descendants": 1,
			"id": 33215771,
			"kids": [
				33215772
			],
			"score": 21,
			"time": 1665839235,
			"title": "Some News",
			"type": "story",
			"url": ""
		}`,
	}

	fetcher := Fetcher{Settings: Configuration{
		ApiBaseUrl:         "",
		BlacklistedDomains: []string{"some-host"},
		Filters:            []FilterItem{{Title: "Test filter", Value: "title"}},
		Database:           Database{Driver: "sqlite3", Database: ":memory:"},
	}}

	fetcher.filters = fetcher.prepareFilters()

	if err := fetcher.setUpRepository(); err != nil {
		t.Errorf("Errof while initializing the repository, %v", err)
	}

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	// Mock the top stories
	httpmock.RegisterResponder("GET", fetcher.Settings.ApiBaseUrl+"/topstories.json",
		httpmock.NewStringResponder(200, expectedPrefetched))
	// Mock stories
	for id, resp := range expectedDigest {
		httpmock.RegisterResponder("GET", fmt.Sprintf(fetcher.Settings.ApiBaseUrl+"/item/%d.json", id),
			httpmock.NewStringResponder(200, resp))
	}

	prefetched, err := fetcher.prefetch()

	if err != nil {
		t.Errorf("Error while prefetching news items, %v", err)
	}

	if len(*prefetched) != 3 {
		t.Errorf("Expected 3 prefetched items, got %d", len(*prefetched))
	}

	if (*prefetched)[0] != 33214440 || (*prefetched)[1] != 33215770 || (*prefetched)[2] != 33215771 {
		t.Errorf("Expected %s prefetched items, got %v", expectedPrefetched, *prefetched)
	}

	unfiltered, filtered, err := fetcher.filter(prefetched)

	if err != nil {
		t.Errorf("Error while filtering news items, %v", err)
	}

	if unfiltered == nil || filtered == nil {
		t.Errorf("filtered and digest must not be nil")
	} else if len(*unfiltered) != 3 {
		t.Errorf("Expected 3 filtered items, got %d", len(*unfiltered))
	}

	if len(*filtered) != 1 {
		t.Fatalf("Expected 1 digest item, got %d", len(*filtered))
	}

	if (*filtered)[0].id != 33214440 {
		t.Errorf("Expected a news item with ID %d, got %d", 33214440, (*filtered)[0].id)
	}
}

func TestRun(t *testing.T) {
	expectedPrefetched := "[33214440,33215770,33215771]"
	expectedDigest := map[int64]string{
		33214440: `{
			"by": "author",
			"descendants": 1,
			"id": 33214440,
			"kids": [
				33216444
			],
			"score": 10,
			"time": 1665839487,
			"title": "Some Title #1",
			"type": "story",
			"url": "https://somehost/1-story-uri/"
		}`,
		33215770: `{
			"by": "author-2",
			"descendants": 1,
			"id": 33215770,
			"kids": [
				33215771
			],
			"score": 20,
			"time": 1665839234,
			"title": "Some Title #2",
			"type": "story",
			"url": "https://some-host/2-story-uri/"
		}`,
		33215771: `{
			"by": "author-3",
			"descendants": 1,
			"id": 33215771,
			"kids": [
				33215772
			],
			"score": 21,
			"time": 1665839235,
			"title": "Some News",
			"type": "story",
			"url": ""
		}`,
	}

	fetcher := Fetcher{Settings: Configuration{
		ApiBaseUrl:         "",
		EmailTo:            "",
		BlacklistedDomains: []string{"some-host"},
		Filters:            []FilterItem{{Title: "Test filter", Value: "title"}},
		Database:           Database{Driver: "sqlite3", Database: ":memory:"},
	}}

	fetcher.filters = fetcher.prepareFilters()

	if err := fetcher.setUpRepository(); err != nil {
		t.Errorf("Errof while initializing the repository, %v", err)
	}

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	// Mock the top stories
	httpmock.RegisterResponder("GET", fetcher.Settings.ApiBaseUrl+"/topstories.json",
		httpmock.NewStringResponder(200, expectedPrefetched))
	// Mock stories
	for id, resp := range expectedDigest {
		httpmock.RegisterResponder("GET", fmt.Sprintf(fetcher.Settings.ApiBaseUrl+"/item/%d.json", id),
			httpmock.NewStringResponder(200, resp))
	}

	results, err := fetcher.Run()

	if err != nil {
		t.Errorf("Error while executing Run, %v", err)
	}

	if results.Filters != 1 {
		t.Errorf("Should be 1 filter, got %d", results.Filters)
	}

	if results.NewItems != 1 {
		t.Errorf("Should be 1 news item in the digest, got %d", results.NewItems)
	}
}
