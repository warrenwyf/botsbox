package target

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tidwall/gjson"

	"../../common/util"
	"../fetchers"
)

var (
	idSeq uint64
)

type Target struct {
	id    uint64
	tried int64
	level uint

	Age       time.Duration
	Priority  int64
	Retry     int64
	RetryWait time.Duration
	Mtag      string
	Client    string

	Url         string
	Method      string
	Query       map[string]string
	Form        map[string]string
	ContentType string

	Dive map[string]*Entry

	ObjectOutputs []*ObjectOutput
	ListOutputs   []*ListOutput

	result *fetchers.Result
	err    error
}

func NewTarget() *Target {
	atomic.AddUint64(&idSeq, 1)

	return &Target{
		id:    idSeq,
		tried: 0,
		level: 0,

		Age:       24 * time.Hour,
		Priority:  0,
		Retry:     3,
		RetryWait: time.Minute,

		Method:      "GET",
		Query:       map[string]string{},
		Form:        map[string]string{},
		ContentType: "html",

		Dive: map[string]*Entry{},

		ObjectOutputs: []*ObjectOutput{},
		ListOutputs:   []*ListOutput{},
	}
}

func NewTargetWithJson(elem *gjson.Result) *Target {
	t := NewTarget()

	ageElem := elem.Get("$age")
	if ageElem.Exists() {
		age, err := util.ParseDuration(ageElem.String())
		if err == nil {
			t.Age = age
		}
	}

	priorityElem := elem.Get("$priority")
	if priorityElem.Exists() {
		priority := priorityElem.Int()
		if priority >= 0 {
			t.Priority = priority
		}
	}

	retryElem := elem.Get("$retry")
	if retryElem.Exists() {
		retry := retryElem.Int()
		t.Retry = retry // Minus means do not retry
	}

	retryWaitElem := elem.Get("$retryWait")
	if retryWaitElem.Exists() {
		retryWait, err := util.ParseDuration(retryWaitElem.String())
		if err == nil {
			t.RetryWait = retryWait
		}
	}

	clientElem := elem.Get("$client")
	if clientElem.Exists() {
		t.Client = clientElem.String()
	}

	diveElem := elem.Get("$dive")
	if diveElem.Exists() {
		diveElem.ForEach(func(kElem, vElem gjson.Result) bool {
			t.Dive[kElem.String()] = NewEntryWithJson(&vElem)
			return true
		})
	}

	outputsElem := elem.Get("$outputs")
	if outputsElem.Exists() {
		outputsElem.ForEach(func(_, outputElem gjson.Result) bool {
			if outputElem.Get("$each").Exists() { // ListOutput
				t.ListOutputs = append(t.ListOutputs, NewListOutputWithJson(&outputElem))
			} else {
				t.ObjectOutputs = append(t.ObjectOutputs, NewObjectOutputWithJson(&outputElem))
			}

			return true
		})
	}

	return t
}

func (self *Target) GetId() uint64 {
	return self.id
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
			browserFetcher.SetQuery(self.Query)
			browserFetcher.SetForm(self.Form)
			browserFetcher.SetContentType(self.ContentType)

			fetcher = browserFetcher

		} else {
			httpFetcher := fetchers.NewHttpFetcher()
			httpFetcher.SetUrl(self.Url)
			httpFetcher.SetMethod(self.Method)
			httpFetcher.SetQuery(self.Query)
			httpFetcher.SetForm(self.Form)
			httpFetcher.SetContentType(self.ContentType)

			fetcher = httpFetcher

		}

	}

	if fetcher == nil {
		self.err = errors.New("No supported fetcher")
		return
	}

	self.result, self.err = fetcher.Fetch()
}
