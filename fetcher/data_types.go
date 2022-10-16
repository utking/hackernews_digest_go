package fetcher

// Data Types

type FilterItem struct {
	Title string
	Value string
}

type PrefetchResults []int64

type DigestItem struct {
	newsTitle string
	newsUrl   string
	id        int64
	createdAt int64
}

type JsonNewsItem struct {
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
	Id    int64  `json:"id"`
	Time  int64  `json:"time"`
}

type Digest []DigestItem

type FetchError struct{}

type Results struct {
	NewItems int
	Filters  int
}

// Constants

const CRLF = "\r\n"
