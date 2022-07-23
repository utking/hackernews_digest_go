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
		t.Fatalf("ApiBaseUrl value [%s] is wrong", cfg.ApiBaseUrl)
	}
	if cfg.Database.Driver != "sqlite3" {
		t.Fatalf("Database driver [%s] value is wrong", cfg.Database.Driver)
	}
	if cfg.Database.Address != "tcp(127.0.0.1:3306)" {
		t.Fatalf("Database Address [%s] value is wrong", cfg.Database.Address)
	}
	if cfg.Database.Username != "user" {
		t.Fatalf("Database username [%s] value is wrong", cfg.Database.Username)
	}
	if cfg.Database.Password != "password" {
		t.Fatalf("Database password [%s] value is wrong", cfg.Database.Password)
	}
	if cfg.Database.Database != "./hackernews_db.sqlite" {
		t.Fatalf("DatabaseFile value [%s] is wrong", cfg.Database.Database)
	}
	if cfg.PurgeAfterDays != 30 {
		t.Fatalf("PurgeAfterDays value [%d] is wrong", cfg.PurgeAfterDays)
	}
	if cfg.EmailTo != "to@example.com" {
		t.Fatalf("EmailTo value [%s] is wrong", cfg.EmailTo)
	}
	if len(cfg.Filters) == 0 {
		t.Fatal("Filters list has no records")
	}
	if cfg.Filters[0].Title != "SQL" {
		t.Fatalf("Filter title [%s] is wrong", cfg.Filters[0].Title)
	}
	if cfg.Filters[0].Value != "sql" {
		t.Fatalf("Filter value [%s] is wrong", cfg.Filters[0].Value)
	}
}

func TestMissingFileGetConfiguration(t *testing.T) {
	_, err := GetConfig("no-file.json")
	if err == nil {
		t.Fatal("Wrong configuration could be loaded")
	}
}
