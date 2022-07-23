package fetcher

import (
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	cfg, err := GetConfig("../config.example.json")
	if err != nil {
		t.Fatalf("Could not load configuration: %s", err)
	}
	if cfg.ApiBaseUrl != "https://hacker-news.firebaseio.com/v0" {
		t.Fatal("ApiBaseUrl value is wrong")
	}
	if cfg.DatabaseFile != "./hackernews_db.sqlite" {
		t.Fatal("DatabaseFile value is wrong")
	}
	if cfg.PurgeAfterDays != 30 {
		t.Fatal("PurgeAfterDays value is wrong")
	}
	if cfg.EmailTo != "to@example.com" {
		t.Fatal("EmailTo value is wrong")
	}
	if len(cfg.Filters) == 0 {
		t.Fatal("Filters list has no records")
	}
	if cfg.Filters[0].Title != "SQL" {
		t.Fatal("Filter title is wrong")
	}
	if cfg.Filters[0].Value != "sql" {
		t.Fatal("Filter value is wrong")
	}
}

func TestMissingFileGetConfiguration(t *testing.T) {
	_, err := GetConfig("no-file.json")
	if err == nil {
		t.Fatal("Wrong configuration could be loaded")
	}
}
