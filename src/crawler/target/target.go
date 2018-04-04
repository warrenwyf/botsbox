package target

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"../fetchers"
	"../rule"
)

var (
	idSeq uint64
)

type Target struct {
	id        uint64
	tried     int64
	level     uint
	hash      string
	createdAt time.Time

	Url        string
	Method     string
	Header     map[string]string
	Query      map[string]string
	Form       map[string]string
	ResultType string
	ApplyedVar map[string]string

	Age       time.Duration
	Priority  int64
	Timeout   time.Duration
	Retry     int64
	RetryWait time.Duration
	Mtag      string
	Client    string

	Dive map[string]*rule.Entry

	ObjectOutputs []*rule.ObjectOutput
	ListOutputs   []*rule.ListOutput

	result *fetchers.Result
	err    error
}

func NewTargetWithTemplate(template *rule.TargetTemplate) *Target {
	t := NewTarget()

	t.Age = template.Age
	t.Priority = template.Priority
	t.Timeout = template.Timeout
	t.Retry = template.Retry
	t.RetryWait = template.RetryWait
	t.Mtag = template.Mtag
	t.Client = template.Client

	t.Dive = template.Dive

	t.ObjectOutputs = template.ObjectOutputs
	t.ListOutputs = template.ListOutputs

	return t
}

func NewTarget() *Target {
	atomic.AddUint64(&idSeq, 1)

	return &Target{
		id:        idSeq,
		tried:     0,
		level:     0,
		createdAt: time.Now(),

		Method:     "GET",
		Header:     map[string]string{},
		Query:      map[string]string{},
		Form:       map[string]string{},
		ResultType: "html",
		ApplyedVar: map[string]string{},

		Age:       24 * time.Hour,
		Priority:  0,
		Timeout:   0,
		Retry:     3,
		RetryWait: time.Minute,

		Dive: map[string]*rule.Entry{},

		ObjectOutputs: []*rule.ObjectOutput{},
		ListOutputs:   []*rule.ListOutput{},
	}
}

func (self *Target) Higher(compare interface{}) bool { // Used by PriorityQueue
	other := compare.(*Target)

	if self.Priority == other.Priority {
		return self.createdAt.Before(other.createdAt)
	}

	return self.Priority > other.Priority
}

func (self *Target) GetId() uint64 {
	return self.id
}

func (self *Target) GetHash() string {
	return self.hash
}

func (self *Target) GetResult() *fetchers.Result {
	return self.result
}

func (self *Target) GetErr() error {
	return self.err
}

func (self *Target) CanRetry() bool {
	return self.Retry > 0 && self.tried <= self.Retry
}

func (self *Target) Crawl() {
	canTry := self.CanRetry() || self.tried == 0
	if !canTry {
		return
	}

	self.tried++

	var fetcher fetchers.Fetcher

	url := strings.ToLower(self.Url)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {

		client := strings.ToLower(self.Client)
		if client == "browser" {
			browserFetcher := fetchers.NewBrowserFetcher()
			browserFetcher.SetUrl(self.Url)
			browserFetcher.SetMethod(self.Method)
			browserFetcher.SetHeader(self.Header)
			browserFetcher.SetQuery(self.Query)
			browserFetcher.SetForm(self.Form)
			browserFetcher.SetResultType(self.ResultType)
			if self.Timeout > 0 {
				browserFetcher.SetTimeout(self.Timeout)
			}

			fetcher = browserFetcher

		} else {
			httpFetcher := fetchers.NewHttpFetcher()
			httpFetcher.SetUrl(self.Url)
			httpFetcher.SetMethod(self.Method)
			httpFetcher.SetHeader(self.Header)
			httpFetcher.SetQuery(self.Query)
			httpFetcher.SetForm(self.Form)
			httpFetcher.SetResultType(self.ResultType)
			if self.Timeout > 0 {
				httpFetcher.SetTimeout(self.Timeout)
			}

			fetcher = httpFetcher

		}

	}

	if fetcher == nil {
		self.err = errors.New("No supported fetcher")
		return
	}

	self.hash = fetcher.Hash()

	self.result, self.err = fetcher.Fetch()
}
