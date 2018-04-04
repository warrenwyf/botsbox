package fetchers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"../../common/util"
)

type HttpFetcher struct {
	timeout    time.Duration
	url        string
	method     string
	header     map[string]string
	query      map[string]string
	form       map[string]string
	resultType string
}

func NewHttpFetcher() *HttpFetcher {
	return &HttpFetcher{
		timeout:    120 * time.Second,
		method:     "GET",
		header:     map[string]string{},
		query:      map[string]string{},
		form:       map[string]string{},
		resultType: "html",
	}
}

func (self *HttpFetcher) Fetch() (*Result, error) {
	url := self.url
	if len(self.query) > 0 {
		url = fmt.Sprintf("%s?%s", url, joinQueryString(self.query))
	}

	var body io.Reader = nil
	if self.method == "POST" {
		form := neturl.Values{}
		for k, v := range self.form {
			form.Add(k, v)
		}
		body = strings.NewReader(form.Encode())
	}

	req, errReq := http.NewRequest(self.method, url, body)
	if errReq != nil {
		return nil, errReq
	}

	for k, v := range self.header {
		req.Header.Add(k, v)
	}

	if self.method == "POST" {
		if len(req.Header.Get("Content-Type")) == 0 {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	client := http.Client{
		Timeout: self.timeout,
	}

	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, errResp
	}
	defer resp.Body.Close()

	bytes, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}

	return &Result{
		Hash:        self.Hash(),
		Format:      ResultFormat_Bytes,
		Content:     bytes,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
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

func (self *HttpFetcher) SetHeader(v map[string]string) {
	self.header = v
}

func (self *HttpFetcher) SetQuery(v map[string]string) {
	self.query = v
}

func (self *HttpFetcher) SetForm(v map[string]string) {
	self.form = v
}

func (self *HttpFetcher) SetResultType(v string) {
	self.resultType = v
}

func (self *HttpFetcher) Hash() string {
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
