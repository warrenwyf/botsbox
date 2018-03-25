package fetchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"../../common/util"
)

type HttpFetcher struct {
	url         string
	method      string
	query       map[string]string
	form        map[string]string
	contentType string
}

func NewHttpFetcher() *HttpFetcher {
	return &HttpFetcher{
		method:      "GET",
		contentType: "html",
	}
}

func (self *HttpFetcher) Fetch() (*Result, error) {
	if self.method == "GET" {
		url := self.url
		if len(self.query) > 0 {
			url = fmt.Sprintf("%s?%s", url, joinQueryString(self.query))
		}

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return &Result{
			Hash:    self.Hash(),
			Format:  ResultFormat_Bytes,
			Content: bytes,
		}, nil
	}

	return nil, nil
}

func (self *HttpFetcher) SetUrl(p *string) {
	self.url = *p
}

func (self *HttpFetcher) SetMethod(p *string) {
	self.method = *p
}

func (self *HttpFetcher) SetQuery(p *map[string]string) {
	self.query = *p
}

func (self *HttpFetcher) SetForm(p *map[string]string) {
	self.form = *p
}

func (self *HttpFetcher) SetContentType(p *string) {
	self.contentType = *p
}

func (self *HttpFetcher) Hash() string {
	obj := &map[string]interface{}{
		"url":         self.url,
		"method":      strings.ToUpper(self.method),
		"query":       self.query,
		"form":        self.form,
		"contentType": strings.ToLower(self.contentType),
	}

	return util.Md5(obj)
}
