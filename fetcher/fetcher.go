package fetcher

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

// Data Types

type PrefetchResults []int64
type DigestItem struct {
	id        int64
	createdAt int64
	newsTitle string
	newsUrl   string
}

type JsonNewsItem struct {
	Id    int64  `json:"id"`
	Time  int64  `json:"time"`
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
}
type Digest []DigestItem
type FetchError struct{}
type Results struct {
	NewItems uint
	Filters  uint
}

// Methods

func (fe *FetchError) Error() string {
	return fmt.Sprintf("Error: %s", "HackerNews data could not be fetched")
}

type Fetcher struct {
	Settings Configuration
	Db       *sql.DB
	filters  []string
}

func (f *Fetcher) purgeOld() error {
	purgeStmt := fmt.Sprintf(`DELETE FROM news_items WHERE 
			date(created_at, "unixepoch", "localtime") < date("now", "-#{f.Settings.PurgeAfterDays} days")`)
	_, err := f.Db.Exec(purgeStmt)
	if err != nil {
		return err
	}
	_, err = f.Db.Exec("VACUUM")
	return err
}

func (f *Fetcher) prepareDb() error {
	var err error
	f.Db, err = sql.Open("sqlite3", f.Settings.DatabaseFile)
	if err != nil {
		return err
	}
	_, err = f.Db.Exec(`CREATE TABLE IF NOT EXISTS news_items
            (
                id INTEGER PRIMARY KEY,
                created_at TEXT NOT NULL,
                news_title TEXT NOT NULL,
                news_url  TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	err = f.purgeOld()
	if err != nil {
		return err
	}
	return nil
}

func (f *Fetcher) prepareFilters() []string {
	var resultFilters []string
	for _, filter := range f.Settings.Filters {
		filterStrings := strings.Split(filter.Value, ",")
		resultFilters = append(resultFilters, filterStrings...)
	}
	return resultFilters
}

func (f *Fetcher) prefetch() (*[]int64, error) {
	result := make([]int64, 0)
	prefetchUrl := fmt.Sprintf("%s/topstories.json", f.Settings.ApiBaseUrl)
	resp, err := http.Get(prefetchUrl)
	if err != nil {
		return &result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return &result, err
	}
	return &result, nil
}

func (f *Fetcher) fetchOne(id int64) (JsonNewsItem, error) {
	var result JsonNewsItem
	prefetchUrl := fmt.Sprintf("%s/item/%d.json", f.Settings.ApiBaseUrl, id)
	resp, err := http.Get(prefetchUrl)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (f *Fetcher) filter(prefetched *[]int64) ([]DigestItem, error) {
	digestItems := make([]DigestItem, 0)
	filteredItems := make([]int64, 0)
	existingIDs := make(map[int64]interface{}, 0)
	newItems := make([]DigestItem, 0)

	// Fetch existing items from the DB
	rows, err := f.Db.Query("SELECT id FROM news_items")
	if err != nil {
		return Digest{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var curId int64
		err = rows.Scan(&curId)
		if err != nil {
			return Digest{}, err
		}
		existingIDs[curId] = 0
	}
	// Fetch news items which do not exist in the DB
	for _, fetchId := range *prefetched {
		_, ok := existingIDs[fetchId]
		if ok {
			continue
		}
		filteredItems = append(filteredItems, fetchId)
		// Fetch the item
		newItem, err := f.fetchOne(fetchId)
		if err != nil {
			log.Println("FETCH_ONE: ", err)
		}
		if newItem.Url == "" {
			newItems = append(newItems, DigestItem{
				id:        newItem.Id,
				createdAt: newItem.Time,
				newsTitle: "-",
				newsUrl:   "-",
			})
		} else {
			digestItem := DigestItem{
				id:        newItem.Id,
				createdAt: newItem.Time,
				newsTitle: newItem.Title,
				newsUrl:   newItem.Url,
			}
			newItems = append(newItems, digestItem)
			for _, filterItem := range f.filters {
				hit, _ := regexp.MatchString("(?i)"+filterItem, newItem.Title)
				if hit {
					digestItems = append(digestItems, digestItem)
					break
				}
			}
		}
	}
	if len(newItems) > 0 {
		stmt, err := f.Db.Prepare("INSERT INTO news_items VALUES (?,?,?,?)")
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
	return digestItems, nil
}

func (f *Fetcher) SendEmail(digest *[]DigestItem) {
	if f.Settings.Smtp.Host == "" {
		log.Println("SMTP Host is empty. Skipping sending the Email")
		return
	}

	subject := "Subject: HackerNews Digest\n"

	digestItemsHtml := ""
	digestItemsText := "Hi!\n\n"
	for _, digestItem := range *digest {
		digestItemsHtml += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n",
			digestItem.newsUrl, digestItem.newsTitle)
		digestItemsText += fmt.Sprintf("* %s - %s\n", digestItem.newsTitle, digestItem.newsUrl)
	}
	mime := "Content-Type: multipart/alternative; boundary=\"boundary-string\"\nMIME-Version: 1.0\n\n"
	textHeader := "--boundary-string\nContent-Type: text/plain; charset=\"utf-8\"\nMIME-Version: 1.0\n" +
		"Content-Transfer-Encoding: quoted-printable\nContent-Disposition: inline\n\n"
	htmlHeader := "--boundary-string\nContent-Type: text/html; charset=\"utf-8\"\nMIME-Version: 1.0\n" +
		"Content-Transfer-Encoding: quoted-printable\nContent-Disposition: inline\n\n"
	digestHtml := fmt.Sprintf(`<html>
          <head>HackerNews Digest</head>
          <body>
            <p>Hi!</p>
            <div>
            <ul>
            %s
            </ul>
            </div>
            <p>Generated: %s</p>
          </body>
        </html>%s`, digestItemsHtml, time.Now().Format("02 Jan 06 15:04 MST"), "\n\n")

	msg := subject + mime + textHeader + digestItemsText + "\n\n" + htmlHeader + digestHtml + "--boundary-string--"

	c, err := smtp.Dial(fmt.Sprintf("%s:%d", f.Settings.Smtp.Host, f.Settings.Smtp.Port))
	if err != nil {
		log.Fatal("EMAIL: ", err)
	}
	if err := c.Mail(f.Settings.Smtp.From); err != nil {
		log.Fatal("EMAIL_SENDER: ", err)
	}
	if err := c.Rcpt(f.Settings.EmailTo); err != nil {
		log.Fatal("EMAIL_RECEIVER: ", err)
	}
	wc, err := c.Data()
	if err != nil {
		log.Fatal("EMAIL_START_CONTENT: ", err)
	}
	_, err = fmt.Fprintf(wc, msg)
	if err != nil {
		log.Fatal("EMAIL_SET_CONTENT: ", err)
	}
	err = wc.Close()
	if err != nil {
		log.Fatal("EMAIL_CLOSE_CONTENT: ", err)
	}
	err = c.Quit()
	if err != nil {
		log.Fatal("EMAIL_QUIT: ", err)
	}
}

func (f *Fetcher) Run() Results {
	f.filters = f.prepareFilters()
	results := Results{NewItems: 0, Filters: uint(len(f.filters))}
	err := f.prepareDb()
	if err != nil {
		log.Fatal("PREPARE DB: ", err)
	}
	defer f.Db.Close()
	var prefetchedItems *[]int64
	prefetchedItems, err = f.prefetch()
	if err != nil {
		log.Fatal("PREFETCH: ", err)
	}
	digest, err := f.filter(prefetchedItems)
	if err != nil {
		log.Fatal("FILTER: ", err)
	}
	results.NewItems = uint(len(digest))
	if len(digest) > 0 {
		if f.Settings.EmailTo != "" {
			f.SendEmail(&digest)
		} else {
			for _, digestItem := range digest {
				fmt.Printf("* %s - %s\n", digestItem.newsTitle, digestItem.newsUrl)
			}
		}
	}
	return results
}
