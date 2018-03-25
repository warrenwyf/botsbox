package job

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"../../xlog"
	"../analyzers"
	"../fetchers"
	"../rule"
	"../sink"
	"../target"
)

type Job struct {
	Title    string
	Interval time.Duration
	Delay    time.Duration

	rule *rule.Rule

	targets           map[uint64]*target.Target
	targetCrawledChan chan *target.Target

	sinkChan          chan<- *sink.SinkPack
	sinkChanConnected bool //
}

func NewJob(title string, ruleContent string) (*Job, error) {
	rule, err := rule.NewRuleWithContent(ruleContent)
	if err != nil {
		return nil, err
	}

	interval, errInterval := rule.GetInterval()
	if errInterval != nil {
		return nil, errInterval
	}

	delay, errDelay := rule.GetDelay()
	if errDelay != nil {
		return nil, errDelay
	}

	job := &Job{
		Title:    title,
		Interval: interval,
		Delay:    delay,

		rule: rule,

		targets:           map[uint64]*target.Target{},
		targetCrawledChan: make(chan *target.Target, 100),

		sinkChanConnected: false,
	}

	return job, nil
}

func NewJobWithFile(title string, rulePath string) (*Job, error) {
	b, err := ioutil.ReadFile(rulePath)
	if err != nil {
		return nil, err
	}

	return NewJob(title, string(b))
}

func (self *Job) Run() {
	entriesElem := self.rule.GetEntries()
	if !entriesElem.Exists() {
		return
	}

	xlog.Outf("Job \"%s\" start to crawl\n", self.Title)

	// Start to crawl targets
	entriesElem.ForEach(func(keyElem, entryElem gjson.Result) bool {
		entry := target.NewEntryWithJson(entryElem)
		targetTemplateElem := self.rule.GetTargetTemplate(entry.Name)
		if targetTemplateElem.Exists() {
			t := target.NewTargetWithJson(targetTemplateElem)
			if t != nil {
				t.Url = entry.Url
				self.startCrawlTarget(t)
			}
		}

		return true
	})

	// Check all targets crawled
	for {
		if len(self.targets) == 0 {
			break
		}

		select {
		case t := <-self.targetCrawledChan:
			self.analyze(t)
		}
	}

	xlog.Outf("Job \"%s\" finished crawling\n", self.Title)
}

func (self *Job) ConnectSink(sink *sink.Sink) {
	self.sinkChan = sink.C
	self.sinkChanConnected = true
}

func (self *Job) startCrawlTarget(t *target.Target) {
	if t != nil {
		self.targets[t.GetId()] = t

		go func() {
			t.Crawl()
			self.targetCrawledChan <- t
		}()
	}
}

func (self *Job) retryCrawlTarget(t *target.Target) {
	if t != nil {
		go func() {
			delay := t.RetryWait
			if delay > 0 {
				time.Sleep(delay)
			}

			t.Crawl()
			self.targetCrawledChan <- t
		}()
	}
}

func (self *Job) analyze(t *target.Target) {
	var err error = nil // Crawled target maybe have fetch error, parse error, etc.

	defer func() {
		if err != nil && t.CanRetry() {
			self.retryCrawlTarget(t)
		}
	}()

	err = t.GetErr() // Fetch error
	if err != nil {
		return
	}

	result := t.GetResult()
	if result == nil {
		err = errors.New("Got nothing from target")
		return
	}

	var analyzerResult *analyzers.Result

	contentType := strings.ToLower(t.ContentType)
	if contentType == "html" {
		htmlAnalyzer := analyzers.NewHtmlAnalyzer(self.rule)

		if result.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = htmlAnalyzer.ParseBytes(result.Content.([]byte), t)
			if err != nil {
				return
			}
		}

	} else if contentType == "json" {

		if result.Format == fetchers.ResultFormat_Bytes {
			//json := gjson.ParseBytes(result.Content.([]byte))
		}

	}

	if analyzerResult != nil {
		for _, t := range analyzerResult.Targets {
			self.startCrawlTarget(t)
		}

		if self.sinkChanConnected {
			for _, pack := range analyzerResult.SinkPacks {
				self.sinkChan <- pack
			}
		}
	}

	delete(self.targets, t.GetId())
}
