package fetchers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"strings"
	"time"

	"../../common/util"
)

var httpTransport = &http.Transport{
	DisableKeepAlives: true,
}

type HttpFetcher struct {
	timeout    time.Duration
	url        string
	method     string
	header     map[string]string
	query      map[string]string
	form       map[string]string
	resultType string
	cookies    []*http.Cookie
	userAgent  string
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

	if len(self.userAgent) > 0 {
		req.Header.Add("User-Agent", self.userAgent)
	}

	client := http.Client{
		Transport: httpTransport,
		Timeout:   self.timeout,
	}

	jar, err := cookiejar.New(nil)
	if err == nil {
		u, err := neturl.Parse(self.url)
		if err == nil {
			jar.SetCookies(u, self.cookies)
		}

		client.Jar = jar
	}

	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, errResp
	}
	defer resp.Body.Close()

	b, errRead := util.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}

	result := &Result{
		Hash:        self.Hash(),
		Format:      ResultFormat_Bytes,
		Content:     b,
		ContentType: headerValueIgnoreCase(resp.Header, "Content-Type"),
	}

	if client.Jar != nil {
		u, err := neturl.Parse(url)
		if err == nil {
			result.Cookies = client.Jar.Cookies(u)
			result.CookiesUrl = u
		}
	}

	return result, nil
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

func (self *HttpFetcher) SetUserAgent(ua string) {
	self.userAgent = ua
}

func (self *HttpFetcher) SetCookies(cookies []*http.Cookie) {
	self.cookies = cookies
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
