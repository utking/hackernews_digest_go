package fetcher

import (
	"os"
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	os.Setenv("CONFIG_FILENAME", "../config.example.json")
	cfg, err := GetConfig()
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
	if cfg.ReverseFilters != false {
		t.Fatal("ReverseFilters value is wrong")
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

func TestDefaultFileGetConfiguration(t *testing.T) {
	_, err := GetConfig()
	if err != nil {
		t.Fatalf("Wrong filename configuration could be loaded: %s", err)
	}
}

func TestMissingFileGetConfiguration(t *testing.T) {
	os.Setenv("CONFIG_FILENAME", "no-file.json")
	_, err := GetConfig()
	if err == nil {
		t.Fatal("Wrong configuration could be loaded")
	}
}
