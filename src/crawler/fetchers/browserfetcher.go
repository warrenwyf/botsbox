package fetchers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"../../common/browser"
	"../../common/mhtml"
	"../../common/util"
)

type BrowserFetcher struct {
	timeout    time.Duration
	url        string
	method     string
	header     map[string]string
	query      map[string]string
	form       map[string]string
	resultType string
}

func NewBrowserFetcher() *BrowserFetcher {
	return &BrowserFetcher{
		timeout:    5 * time.Minute,
		method:     "GET",
		header:     map[string]string{},
		query:      map[string]string{},
		form:       map[string]string{},
		resultType: "html",
	}
}

func (self *BrowserFetcher) Fetch() (*Result, error) {
	url := self.url
	if len(self.query) > 0 {
		url = fmt.Sprintf("%s?%s", url, joinQueryString(self.query))
	}

	page := browser.GetBrowser().CreatePage()
	if page == nil {
		return nil, errors.New("Can not create page")
	}
	defer page.Close()

	page.Load(url, self.timeout)
	html := mhtml.GetHtml(page.ExportMHtml(self.timeout))

	if html == nil {
		return nil, errors.New("Nothing got via browser")
	}

	return &Result{
		Hash:    self.Hash(),
		Format:  ResultFormat_Bytes,
		Content: html,
	}, nil
}

func (self *BrowserFetcher) SetTimeout(v time.Duration) {
	self.timeout = v
}

func (self *BrowserFetcher) SetUrl(v string) {
	self.url = v
}

func (self *BrowserFetcher) SetMethod(v string) {
	self.method = v
}

func (self *BrowserFetcher) SetHeader(v map[string]string) {
	self.header = v
}

func (self *BrowserFetcher) SetQuery(v map[string]string) {
	self.query = v
}

func (self *BrowserFetcher) SetForm(v map[string]string) {
	self.form = v
}

func (self *BrowserFetcher) SetResultType(v string) {
	self.resultType = v
}

func (self *BrowserFetcher) Hash() string {
	obj := &map[string]interface{}{
		"url":        self.url,
		"method":     strings.ToUpper(self.method),
		"header":     self.header,
		"query":      self.query,
		"form":       self.form,
		"resultType": strings.ToLower(self.resultType),
	}

	return util.Md5(obj)
}
