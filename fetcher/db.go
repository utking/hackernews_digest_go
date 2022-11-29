package fetcher

import (
	"github.com/jmoiron/sqlx"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Constants

const (
	TableName   = "news_items"
	CreateTable = `CREATE TABLE IF NOT EXISTS %s
(
	id INTEGER PRIMARY KEY,
	created_at INTEGER NOT NULL,
	news_title TEXT NOT NULL,
	news_url  TEXT NOT NULL
)`

	DblCrLf          = CRLF + CRLF
	SQLiteVacuum     = "VACUUM"
	MySQLVacuum      = "SELECT 1"
	SelectItems      = "SELECT id FROM %s"
	InsertItems      = "INSERT INTO %s VALUES (?,?,?,?)"
	SQLitePurgeItems = "DELETE FROM %s WHERE date(created_at, \"unixepoch\", \"localtime\") < date(\"now\", \"-%d days\")"
	MySQLPurgeItems  = "DELETE FROM %s WHERE FROM_UNIXTIME(created_at) <= (NOW() - INTERVAL %d DAY)"
)

var PurgeItems string
var Vacuum string

type DataRepository struct {
	db         *sqlx.DB
	tbl_prefix string
	dbConfig   Database
	reverse    bool
	purgeAfter uint
}

// Remove news items older than `purgeAfter` days
func (repo *DataRepository) purgeOld() error {
	purgeStmt := fmt.Sprintf(PurgeItems, repo.tbl_prefix+TableName, repo.purgeAfter)

	if _, err := repo.db.Exec(purgeStmt); err != nil {
		return err
	}

	_, err := repo.db.Exec(Vacuum)

	return err
}

// Open a database file and purge old news items from it
func (repo *DataRepository) prepareDB() error {
	var err error

	switch repo.dbConfig.Driver {
	case "sqlite3":
		repo.db, err = sqlx.Open(repo.dbConfig.Driver, repo.dbConfig.Database)
		PurgeItems = SQLitePurgeItems
		Vacuum = SQLiteVacuum
	case "mysql":
		repo.db, err = sqlx.Open(repo.dbConfig.Driver,
			fmt.Sprintf("%s:%s@%s/%s", repo.dbConfig.Username,
				repo.dbConfig.Password, repo.dbConfig.Address, repo.dbConfig.Database))
		PurgeItems = MySQLPurgeItems
		Vacuum = MySQLVacuum
	default:
		return fmt.Errorf("wrong repository driver")
	}

	if err != nil {
		return err
	}

	if _, err := repo.db.Exec(fmt.Sprintf(CreateTable, repo.tbl_prefix+TableName)); err != nil {
		return err
	}

	if err := repo.purgeOld(); err != nil {
		return err
	}

	return nil
}

// Entry point for initializing a database
func (repo *DataRepository) Init() error {
	if repo.reverse {
		repo.tbl_prefix = "reverse_"
	}

	if err := repo.prepareDB(); err != nil {
		return err
	}

	return nil
}

func contains(list []int64, val int64) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}

	return false
}

// Pull existing news items' IDs
func (repo *DataRepository) GetIDsToPull(prefetched *[]int64) ([]int64, error) {
	var (
		itemsToCheck []int64
		existingIDs  []int64
	)

	query, args, err := sqlx.In(fmt.Sprintf(SelectItems+" WHERE id IN (?)", repo.tbl_prefix+TableName), *prefetched)

	if err != nil {
		return itemsToCheck, err
	}

	rows, err := repo.db.Query(query, args...)

	if err != nil {
		return itemsToCheck, err
	}

	defer rows.Close()

	for rows.Next() {
		var curID int64

		if err := rows.Scan(&curID); err != nil {
			return itemsToCheck, err
		}

		existingIDs = append(existingIDs, curID)
	}

	for _, p := range *prefetched {
		if !contains(existingIDs, p) {
			itemsToCheck = append(itemsToCheck, p)
		}
	}

	return itemsToCheck, nil
}

// Add the provided news items to the database
func (repo *DataRepository) UpdateItems(newItems *[]DigestItem) error {
	stmt, err := repo.db.Prepare(fmt.Sprintf(InsertItems, repo.tbl_prefix+TableName))

	if err != nil {
		return err
	}

	for _, newItem := range *newItems {
		if _, err := stmt.Exec(newItem.id, newItem.createdAt, newItem.newsTitle, newItem.newsUrl); err != nil {
			return err
		}
	}

	return nil
}

// Close the database
func (repo *DataRepository) Close() {
	repo.db.Close()
}
