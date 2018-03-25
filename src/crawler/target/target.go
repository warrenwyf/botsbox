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

type Entry struct {
	Name string
	Url  string
}

func NewEntryWithJson(elem gjson.Result) *Entry {
	entry := &Entry{}

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		entry.Name = nameElem.String()
	}

	urlElem := elem.Get("$url")
	if urlElem.Exists() {
		entry.Url = urlElem.String()
	}

	return entry
}

type Output struct {
	Name string
	Data map[string]string
}

func NewOutputWithJson(elem gjson.Result) *Output {
	output := &Output{
		Data: map[string]string{},
	}

	nameElem := elem.Get("$name")
	if nameElem.Exists() {
		output.Name = nameElem.String()
	}

	dataElem := elem.Get("$data")
	if dataElem.Exists() {
		mapElem := dataElem.Map()
		if mapElem != nil {
			for k, v := range mapElem {
				output.Data[k] = v.String()
			}
		}
	}

	return output
}

type Target struct {
	id    uint64
	tried int64
	level uint

	Age       time.Duration
	Priority  int64
	Retry     int64
	RetryWait time.Duration
	Mtag      string

	Url         string
	Method      string
	Query       map[string]string
	Form        map[string]string
	ContentType string

	Dive    map[string]*Entry
	Outputs []*Output

	result *fetchers.Result
	err    error
}

func NewTarget() *Target {
	atomic.AddUint64(&idSeq, 1)

	return &Target{
		id:    idSeq,
		tried: 0,
		level: 0,

		Age:       time.Duration(24) * time.Hour,
		Priority:  0,
		Retry:     3,
		RetryWait: time.Minute,

		Method:      "GET",
		ContentType: "html",

		Dive:    map[string]*Entry{},
		Outputs: []*Output{},
	}
}

func NewTargetWithJson(elem gjson.Result) *Target {
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

	diveElem := elem.Get("$dive")
	if diveElem.Exists() {
		dive := diveElem.Map()
		if dive != nil {
			for k, v := range dive {
				t.Dive[k] = NewEntryWithJson(v)
			}
		}
	}

	outputsElem := elem.Get("$outputs")
	if outputsElem.Exists() {
		outputsElem.ForEach(func(keyElem, ouputElem gjson.Result) bool {
			t.Outputs = append(t.Outputs, NewOutputWithJson(ouputElem))
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
		httpFetcher := fetchers.NewHttpFetcher()
		httpFetcher.SetUrl(&self.Url)
		httpFetcher.SetMethod(&self.Method)
		httpFetcher.SetQuery(&self.Query)
		httpFetcher.SetForm(&self.Form)
		httpFetcher.SetContentType(&self.ContentType)

		fetcher = httpFetcher
	}

	if fetcher == nil {
		self.err = errors.New("No supported fetcher")
	}

	self.result, self.err = fetcher.Fetch()
}
