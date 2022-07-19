package fetcher

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

// Constants

const SQL_DRIVER = "sqlite3"
const CREATE_TABLE = `CREATE TABLE IF NOT EXISTS news_items
(
	id INTEGER PRIMARY KEY,
	created_at TEXT NOT NULL,
	news_title TEXT NOT NULL,
	news_url  TEXT NOT NULL
)`
const CRLF = "\r\n"
const DBL_CRLF = CRLF + CRLF
const REGEX_CASE_INSENSITIVE = "(?i)"
const VACUUM = "VACUUM"
const SELECT_ITEMS = "SELECT id FROM news_items"
const INSERT_ITEM = "INSERT INTO news_items VALUES (?,?,?,?)"
const BOUNDARY_STRING = "--boundary-string--"
const EMAIL_MIME_HEADERS = "Content-Type: multipart/alternative; boundary=\"boundary-string\"" + CRLF +
	"MIME-Version: 1.0" + DBL_CRLF
const EMAIL_TEXT_HEADER = "--boundary-string" + CRLF + "Content-Type: text/plain; charset=\"utf-8\"" + CRLF +
	"MIME-Version: 1.0" + CRLF + "Content-Transfer-Encoding: quoted-printable" + CRLF +
	"Content-Disposition: inline" + DBL_CRLF
const EMAIL_HTML_HEADER = "--boundary-string" + CRLF + "Content-Type: text/html; charset=\"utf-8\"" + CRLF +
	"MIME-Version: 1.0" + CRLF + "Content-Transfer-Encoding: quoted-printable" + CRLF +
	"Content-Disposition: inline" + DBL_CRLF
const DIGEST_ITEM_TEXT_TEMPLATE = "* %s - %s" + CRLF
const DIGEST_ITEM_HTML_TEMPLATE = "<li><a href=\"%s\">%s</a></li>" + CRLF
const DIGEST_HTML_TEMPLATE = `<html>
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
</html>%s`

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
	purgeStmt := `DELETE FROM news_items WHERE 
			date(created_at, "unixepoch", "localtime") < date("now", "-#{f.Settings.PurgeAfterDays} days")`
	_, err := f.Db.Exec(purgeStmt)
	if err != nil {
		return err
	}
	_, err = f.Db.Exec(VACUUM)
	return err
}

func (f *Fetcher) prepareDb() error {
	var err error
	f.Db, err = sql.Open(SQL_DRIVER, f.Settings.DatabaseFile)
	if err != nil {
		return err
	}
	_, err = f.Db.Exec(CREATE_TABLE)
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

func (f *Fetcher) filter(prefetched *[]int64, reverse bool) ([]DigestItem, error) {
	digestItems := make([]DigestItem, 0)
	existingIDs := make(map[int64]interface{}, 0)
	newItems := make([]DigestItem, 0)
	anyFilterHit := false

	// Fetch existing items from the DB
	rows, err := f.Db.Query(SELECT_ITEMS)
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
		// Fetch the item
		newItem, err := f.fetchOne(fetchId)
		if err != nil {
			log.Println("FETCH_ONE: ", err)
		}
		// Set a dumb URL and Title for items that don't have a URL
		if newItem.Url == "" {
			newItems = append(newItems, DigestItem{
				id:        newItem.Id,
				createdAt: newItem.Time,
				newsTitle: "-",
				newsUrl:   "-",
			})
		} else {
			// And now the valid items can be processed
			digestItem := DigestItem{
				id:        newItem.Id,
				createdAt: newItem.Time,
				newsTitle: newItem.Title,
				newsUrl:   newItem.Url,
			}
			newItems = append(newItems, digestItem)

			if reverse {
				anyFilterHit = false
				for _, filterItem := range f.filters {
					hit, _ := regexp.MatchString(REGEX_CASE_INSENSITIVE+filterItem, newItem.Title)
					if hit {
						anyFilterHit = true
						break
					}
				}
				if !anyFilterHit {
					digestItems = append(digestItems, digestItem)
				}

			} else {
				for _, filterItem := range f.filters {
					hit, _ := regexp.MatchString(REGEX_CASE_INSENSITIVE+filterItem, newItem.Title)
					if hit {
						digestItems = append(digestItems, digestItem)
						break
					}
				}
			}
		}
	}

	if len(newItems) > 0 {
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
	return digestItems, nil
}

func (f *Fetcher) SendEmail(digest *[]DigestItem) {
	if f.Settings.Smtp.Host == "" {
		log.Println("SMTP Host is empty. Skipping sending the Email")
		return
	}

	headers := make(map[string]string)
	headers["From"] = f.Settings.Smtp.From
	headers["Subject"] = f.Settings.Smtp.Subject

	messageStart := ""
	for k, v := range headers {
		messageStart += fmt.Sprintf("%s: %s" + CRLF, k, v)
	}

	digestItemsHtml := ""
	digestItemsText := "Hi!" + DBL_CRLF
	for _, digestItem := range *digest {
		digestItemsHtml += fmt.Sprintf(DIGEST_ITEM_HTML_TEMPLATE,
			digestItem.newsUrl, digestItem.newsTitle)
		digestItemsText += fmt.Sprintf(DIGEST_ITEM_TEXT_TEMPLATE, digestItem.newsTitle, digestItem.newsUrl)
	}
	mime := EMAIL_MIME_HEADERS
	textHeader := EMAIL_TEXT_HEADER
	htmlHeader := EMAIL_HTML_HEADER
	digestHtml := fmt.Sprintf(DIGEST_HTML_TEMPLATE, digestItemsHtml, time.Now().Format("02 Jan 06 15:04 MST"), DBL_CRLF)

	msg := messageStart + mime + textHeader + digestItemsText + DBL_CRLF + htmlHeader + digestHtml + BOUNDARY_STRING

	c, err := smtp.Dial(fmt.Sprintf("%s:%d", f.Settings.Smtp.Host, f.Settings.Smtp.Port))
	if err != nil {
		log.Fatal("EMAIL: ", err)
	}

	auth := smtp.PlainAuth("", f.Settings.Smtp.Username, f.Settings.Smtp.Password, f.Settings.Smtp.Host)

	if f.Settings.Smtp.UseTls {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         f.Settings.Smtp.Host,
		}
		c.StartTLS(tlsconfig)
	}

	if err = c.Auth(auth); err != nil {
		log.Panic(err)
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
	_, err = fmt.Fprint(wc, msg)
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

func (f *Fetcher) Run(reverseFilters ...bool) Results {
	reverse := false
	if len(reverseFilters) > 0 {
		reverse = reverseFilters[0]
	}

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
	digest, err := f.filter(prefetchedItems, reverse)
	if err != nil {
		log.Fatal("FILTER: ", err)
	}
	results.NewItems = uint(len(digest))
	if len(digest) > 0 {
		if f.Settings.EmailTo != "" {
			f.SendEmail(&digest)
		} else {
			for _, digestItem := range digest {
				fmt.Printf("* %s - %s" + CRLF, digestItem.newsTitle, digestItem.newsUrl)
			}
		}
	}
	return results
}
