package fetcher

import "testing"

func PrepareItem() JsonNewsItem {
	item := JsonNewsItem{
		Id:    1,
		Title: "Some Title",
		Url:   "http://some.url/abc",
	}
	return item
}

func TestRunFilterHit(t *testing.T) {
	item := PrepareItem()
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
	item := PrepareItem()
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
	item := PrepareItem()
	config := Configuration{
		Filters: []FilterItem{
			{
				Title: "HitTest",
				Value: "Title",
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

func TestReverseFilterMiss(t *testing.T) {
	item := PrepareItem()
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
