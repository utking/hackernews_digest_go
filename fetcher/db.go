package fetcher

import (
	"database/sql"
	"log"
)

// Constants

const SQL_DRIVER = "sqlite3"
const CREATE_TABLE = `CREATE TABLE IF NOT EXISTS news_items
(
	id INTEGER PRIMARY KEY,
	created_at TEXT NOT NULL,
	news_title TEXT NOT NULL,
	news_url  TEXT NOT NULL
)`

const DBL_CRLF = CRLF + CRLF
const VACUUM = "VACUUM"
const SELECT_ITEMS = "SELECT id FROM news_items"
const INSERT_ITEM = "INSERT INTO news_items VALUES (?,?,?,?)"

type DataRepository struct {
	dbConfig   string
	purgeAfter uint
	Db         *sql.DB
}

// Remove news items older than `purgeAfter` days
func (repo *DataRepository) purgeOld() error {
	purgeStmt := `DELETE FROM news_items WHERE 
			date(created_at, "unixepoch", "localtime") < date("now", "-#{f.purgeAfter} days")`
	_, err := repo.Db.Exec(purgeStmt)
	if err != nil {
		return err
	}
	_, err = repo.Db.Exec(VACUUM)
	return err
}

// Open a database file and purge old news items from it
func (repo *DataRepository) prepareDb() error {
	var err error
	repo.Db, err = sql.Open(SQL_DRIVER, repo.dbConfig)
	if err != nil {
		return err
	}
	_, err = repo.Db.Exec(CREATE_TABLE)
	if err != nil {
		return err
	}
	err = repo.purgeOld()
	if err != nil {
		return err
	}
	return nil
}

// Entry point for initializing a database
func (repo *DataRepository) Init() {
	err := repo.prepareDb()
	if err != nil {
		log.Fatal("PREPARE DB: ", err)
	}
}

// Pull existing news items' IDs
func (repo *DataRepository) GetExistingIDs() (map[int64]interface{}, error) {
	existingIDs := make(map[int64]interface{}, 0)
	rows, err := repo.Db.Query(SELECT_ITEMS)
	if err != nil {
		return existingIDs, err
	}
	defer rows.Close()
	for rows.Next() {
		var curId int64
		err = rows.Scan(&curId)
		if err != nil {
			return existingIDs, err
		}
		existingIDs[curId] = 0
	}
	return existingIDs, nil
}

// Add the provided news items to the database
func (f *DataRepository) UpdateItems(newItems []DigestItem) {
	stmt, err := f.Db.Prepare(INSERT_ITEM)
	if err != nil {
		log.Fatal("PREPARE: ", err)
	} else {
		for _, newItem := range newItems {
			_, err := stmt.Exec(newItem.id, newItem.createdAt, newItem.newsTitle, newItem.newsUrl)
			if err != nil {
				log.Fatal("INSERT: ", err)
			}
		}
	}
}

// Close the database
func (repo *DataRepository) Close() {
	repo.Db.Close()
}
