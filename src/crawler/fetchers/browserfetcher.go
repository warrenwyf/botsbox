package fetchers

import (
	"fmt"
	"strings"

	"github.com/headzoo/surf"

	"../../common/util"
)

type BrowserFetcher struct {
	url         string
	method      string
	query       map[string]string
	form        map[string]string
	contentType string
}

func NewBrowserFetcher() *BrowserFetcher {
	return &BrowserFetcher{
		method:      "GET",
		contentType: "html",
	}
}

func (self *BrowserFetcher) Fetch() (*Result, error) {
	if self.method == "GET" {
		url := self.url
		if len(self.query) > 0 {
			url = fmt.Sprintf("%s?%s", url, joinQueryString(self.query))
		}

		browser := surf.NewBrowser()
		err := browser.Open(url)
		if err != nil {
			return nil, err
		}

		return &Result{
			Hash:    self.Hash(),
			Format:  ResultFormat_Browser,
			Content: browser,
		}, nil
	}

	return nil, nil
}

func (self *BrowserFetcher) SetUrl(p *string) {
	self.url = *p
}

func (self *BrowserFetcher) SetMethod(p *string) {
	self.method = *p
}

func (self *BrowserFetcher) SetQuery(p *map[string]string) {
	self.query = *p
}

func (self *BrowserFetcher) SetForm(p *map[string]string) {
	self.form = *p
}

func (self *BrowserFetcher) SetContentType(p *string) {
	self.contentType = *p
}

func (self *BrowserFetcher) Hash() string {
	obj := &map[string]interface{}{
		"url":         self.url,
		"method":      strings.ToUpper(self.method),
		"query":       self.query,
		"form":        self.form,
		"contentType": strings.ToLower(self.contentType),
	}

	return util.Md5(obj)
}
