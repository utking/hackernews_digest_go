package fetcher

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Constants

const TABLE_NAME = "news_items"
const CREATE_TABLE = `CREATE TABLE IF NOT EXISTS %s
(
	id INTEGER PRIMARY KEY,
	created_at INTEGER NOT NULL,
	news_title TEXT NOT NULL,
	news_url  TEXT NOT NULL
)`

const DBL_CRLF = CRLF + CRLF
const SQLITE_VACUUM = "VACUUM"
const MYSQL_VACUUM = "SELECT 1"
const SELECT_ITEMS = "SELECT id FROM %s"
const INSERT_ITEM = "INSERT INTO %s VALUES (?,?,?,?)"
const SQLITE_PURGE_ITEMS = "DELETE FROM %s WHERE date(created_at, \"unixepoch\", \"localtime\") < date(\"now\", \"-%d days\")"
const MYSQL_PURGE_ITEMS = "DELETE FROM %s WHERE FROM_UNIXTIME(created_at) <= (NOW() - INTERVAL %d DAY)"

var PURGE_ITEMS string
var VACUUM string

type DataRepository struct {
	dbConfig   Database
	purgeAfter uint
	db         *sql.DB
	reverse    bool
	tbl_prefix string
}

// Remove news items older than `purgeAfter` days
func (repo *DataRepository) purgeOld() error {
	purgeStmt := fmt.Sprintf(PURGE_ITEMS, repo.tbl_prefix+TABLE_NAME, repo.purgeAfter)
	_, err := repo.db.Exec(purgeStmt)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(VACUUM)
	return err
}

// Open a database file and purge old news items from it
func (repo *DataRepository) prepareDb() error {
	var err error
	if repo.dbConfig.Driver == "sqlite3" {
		repo.db, err = sql.Open(repo.dbConfig.Driver, repo.dbConfig.Database)
		PURGE_ITEMS = SQLITE_PURGE_ITEMS
		VACUUM = SQLITE_VACUUM
	} else if repo.dbConfig.Driver == "mysql" {
		repo.db, err = sql.Open(repo.dbConfig.Driver,
			fmt.Sprintf("%s:%s@%s/%s", repo.dbConfig.Username,
				repo.dbConfig.Password, repo.dbConfig.Address, repo.dbConfig.Database))
		PURGE_ITEMS = MYSQL_PURGE_ITEMS
		VACUUM = MYSQL_VACUUM
	}
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(fmt.Sprintf(CREATE_TABLE, repo.tbl_prefix+TABLE_NAME))
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
	if repo.reverse {
		repo.tbl_prefix = "reverse_"
	}
	err := repo.prepareDb()
	if err != nil {
		log.Fatal("PREPARE DB: ", err)
	}
}

// Pull existing news items' IDs
func (repo *DataRepository) GetExistingIDs() (map[int64]interface{}, error) {
	existingIDs := make(map[int64]interface{}, 0)
	rows, err := repo.db.Query(fmt.Sprintf(SELECT_ITEMS, repo.tbl_prefix+TABLE_NAME))
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
func (repo *DataRepository) UpdateItems(newItems []DigestItem) {
	stmt, err := repo.db.Prepare(fmt.Sprintf(INSERT_ITEM, repo.tbl_prefix+TABLE_NAME))
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
	repo.db.Close()
}
