package fetcher

import (
	"testing"

	"github.com/jarcoal/httpmock"
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
	filters := fetcher.prepareFilters()

	if len(filters) != 1 {
		t.Fatal("Wrong number of filters")
	}

	if fetcher.RunFilter(&item) {
		t.Fatal("Filter did not work as expected")
	}
}

func TestRunFilterMiss(t *testing.T) {
	item := prepareItem()
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "MissTest",
				Value: "Titles",
			},
		},
	}
	fetcher := Fetcher{Settings: config, Reverse: false}
	filters := fetcher.prepareFilters()

	if len(filters) != 1 {
		t.Fatal("Wrong number of filters")
	}

	if fetcher.RunFilter(&item) {
		t.Fatal("Filter did not work as expected")
	}
}

func TestReverseFilterHit(t *testing.T) {
	item := prepareItem()
	fetcher := Fetcher{Settings: Configuration{
		Filters: []FilterItem{
			{
				Title: "HitTest",
				Value: "Title",
			},
		},
	}, Reverse: true}
	filters := fetcher.prepareFilters()

	if len(filters) != 1 {
		t.Fatal("Wrong number of filters", len(filters))
	}

	if !fetcher.RunFilter(&item) {
		t.Fatal("Filter did not work as expected")
	}
}

func TestReverseFilterMiss(t *testing.T) {
	item := prepareItem()
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "MissTest",
				Value: "Titles",
			},
		},
	}
	fetcher := Fetcher{Settings: config, Reverse: true}
	filters := fetcher.prepareFilters()

	if len(filters) != 1 {
		t.Fatal("Wrong number of filters")
	}

	if !fetcher.RunFilter(&item) {
		t.Fatal("Filter did not work as expected")
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
	filters := fetcher.prepareFilters()

	if len(filters) != 2 {
		t.Fatal("Wrong number of filters")
	}

	if filters[0] != "some" || filters[1] != "word" {
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
		"descendants": 34,
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
