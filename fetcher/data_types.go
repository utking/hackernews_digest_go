package fetcher

// Data Types

type FilterItem struct {
	Title string
	Value string
}

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

const CRLF = "\r\n"
