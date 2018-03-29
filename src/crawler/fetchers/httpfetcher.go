package fetchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"../../common/util"
)

type HttpFetcher struct {
	timeout     time.Duration
	url         string
	method      string
	query       map[string]string
	form        map[string]string
	contentType string
}

func NewHttpFetcher() *HttpFetcher {
	return &HttpFetcher{
		timeout:     120 * time.Second,
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

		client := http.Client{
			Timeout: self.timeout,
		}
		resp, err := client.Get(url)
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

func (self *HttpFetcher) SetTimeout(v time.Duration) {
	self.timeout = v
}

func (self *HttpFetcher) SetUrl(v string) {
	self.url = v
}

func (self *HttpFetcher) SetMethod(v string) {
	self.method = v
}

func (self *HttpFetcher) SetQuery(v map[string]string) {
	self.query = v
}

func (self *HttpFetcher) SetForm(v map[string]string) {
	self.form = v
}

func (self *HttpFetcher) SetContentType(v string) {
	self.contentType = v
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
