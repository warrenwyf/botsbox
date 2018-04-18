package fetchers

import (
	"net/http"
	"net/url"
)

const (
	ResultFormat_Bytes = iota
)

type Fetcher interface {
	Fetch() (result *Result, err error)
	Hash() string // Used to compare two fetchers are equal or not
}

type Result struct {
	Hash string

	Format      int
	Content     interface{}
	ContentType string

	Cookies    []*http.Cookie
	CookiesUrl *url.URL
}

func (r *Result) ToString() string {
	switch r.Format {
	case ResultFormat_Bytes:
		return string(r.Content.([]byte))
	}

	return ""
}
