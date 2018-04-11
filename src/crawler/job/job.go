package job

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"../../common/queue"
	"../../store"
	"../../xlog"
	"../analyzers"
	"../fetchers"
	"../rule"
	"../sink"
	"../target"
)

type Job struct {
	title    string
	interval time.Duration
	delay    time.Duration

	timeout     time.Duration
	concurrency int

	rule *rule.Rule

	targetsQueue           *queue.PriorityQueue      // Targets queue waiting for crawling
	targetsInCrawling      map[uint64]*target.Target // Targets crawling
	targetNotifyChan       chan *target.Target       // Notify target's status has changed
	targetNotifyChanClosed bool

	sinkChan          chan<- *sink.SinkPack
	sinkChanConnected bool

	runAt               time.Time
	crawledTargetsCount uint64
}

func NewJob(title string, ruleContent string) (*Job, error) {
	rule, err := rule.NewRuleWithContent(ruleContent)
	if err != nil {
		return nil, err
	}

	job := &Job{
		title:    title,
		interval: rule.Interval,
		delay:    rule.Delay,

		timeout:     rule.Timeout,
		concurrency: rule.Concurrency,

		rule: rule,

		targetsQueue:      queue.NewPriorityQueue(),
		targetsInCrawling: map[uint64]*target.Target{},

		sinkChanConnected: false,

		crawledTargetsCount: 0,
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

func (self *Job) GetTitle() string {
	return self.title
}

func (self *Job) GetFn() func() {
	return self.fn
}

func (self *Job) GetInterval() time.Duration {
	return self.interval
}

func (self *Job) GetDelay() time.Duration {
	return self.delay
}

func (self *Job) fn() {
	xlog.Outf("Job \"%s\" start to crawl\n", self.title)

	self.runAt = time.Now()

	self.targetNotifyChan = make(chan *target.Target)
	self.targetNotifyChanClosed = false

	defer func() {
		self.targetNotifyChanClosed = true
		close(self.targetNotifyChan)
	}()

	// Start to crawl targets
	for _, entry := range self.rule.Entries {
		targetTemplate, ok := self.rule.TargetTemplates[entry.Name]
		if ok {
			targets := target.MakeTargetsWithRule(entry, targetTemplate)
			for _, t := range targets {
				self.targetsQueue.Push(t)
			}
		}
	}

	// Check all targets crawled
	for {
		runningCount := len(self.targetsInCrawling)
		waitingCount := self.targetsQueue.Len()

		if runningCount == 0 && waitingCount == 0 {
			break
		}

		if runningCount < self.concurrency || self.concurrency <= 0 {
			if waitingCount > 0 {
				t := self.targetsQueue.Pop().(*target.Target)
				self.targetsInCrawling[t.GetId()] = t

				go self.crawl(t)
			}
		}

		// Notice: runningCount or waitingCount may changed
		if len(self.targetsInCrawling) > 0 {
			left := self.runAt.Add(self.timeout).Sub(time.Now())
			if left > 0 {
				var t *target.Target

				tm := time.NewTimer(left)
				select {
				case <-tm.C:
				case t = <-self.targetNotifyChan:
					tm.Stop()
				}

				if t != nil {

					if err := t.GetFetchErr(); err != nil { // fetch error
						xlog.Errln("Fetch target error:", err)

						if t.CanRetry() {
							go self.recrawl(t)
						}

					} else {
						if t.Analyzed {
							if t.AnalyzeErr != nil {
								xlog.Errln("Analyze target error:", t.AnalyzeErr)

							}

							self.crawledTargetsCount++
							delete(self.targetsInCrawling, t.GetId())
						} else {
							go self.analyze(t)

						}

					}

				}

			}
		}

		elapse := time.Now().Sub(self.runAt)
		if self.timeout > 0 && elapse > self.timeout { // Check timeout
			xlog.Outf("Job \"%s\" did not finish crawling, timeout\n", self.title)
			break
		} else if elapse > self.interval && self.crawledTargetsCount == 0 { // Check crawled count
			xlog.Outf("Job \"%s\" did not crawl anything during interval time\n", self.title)
			break
		}
	}

	xlog.Outf("Job \"%s\" finished crawling\n", self.title)
}

func (self *Job) GetRunAt() time.Time {
	return self.runAt
}

func (self *Job) GetCrawledTargetsCount() uint64 {
	return self.crawledTargetsCount
}

func (self *Job) ConnectSink(sink *sink.Sink) {
	self.sinkChan = sink.C
	self.sinkChanConnected = true
}

func (self *Job) crawl(t *target.Target) {
	t.Crawl()

	if !self.targetNotifyChanClosed {
		self.targetNotifyChan <- t
	}
}

func (self *Job) recrawl(t *target.Target) {
	delay := t.RetryWait
	if delay > 0 {
		time.Sleep(delay)
	}

	self.crawl(t)
}

func (self *Job) analyze(t *target.Target) {
	defer func() {
		t.Analyzed = true

		if !self.targetNotifyChanClosed {
			self.targetNotifyChan <- t
		}
	}()

	fetchResult := t.GetFetchResult()
	if fetchResult == nil {
		t.AnalyzeErr = errors.New(fmt.Sprintf("Got nothing from target %s", t.Url))
		return
	}

	var analyzerResult *analyzers.Result
	var err error

	resultType := strings.ToLower(t.ResultType)
	if resultType == "html" {
		htmlAnalyzer := analyzers.NewHtmlAnalyzer(self.rule)

		if fetchResult.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = htmlAnalyzer.ParseBytes(
				fetchResult.Content.([]byte),
				fetchResult.ContentType, t)
			if err != nil {
				t.AnalyzeErr = err
				return
			}
		}

	} else if resultType == "json" {
		jsonAnalyzer := analyzers.NewJsonAnalyzer(self.rule)

		if fetchResult.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = jsonAnalyzer.ParseBytes(fetchResult.Content.([]byte), t)
			if err != nil {
				t.AnalyzeErr = err
				return
			}
		}

	} else if resultType == "xml" {
		xmlAnalyzer := analyzers.NewXmlAnalyzer(self.rule)

		if fetchResult.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = xmlAnalyzer.ParseBytes(fetchResult.Content.([]byte),
				fetchResult.ContentType, t)
			if err != nil {
				t.AnalyzeErr = err
				return
			}
		}

	} else if resultType == "webp" ||
		resultType == "jpg" ||
		resultType == "jpeg" ||
		resultType == "png" ||
		resultType == "bmp" ||
		resultType == "gif" {
		binaryAnalyzer := analyzers.NewBinaryAnalyzer(self.rule)

		if fetchResult.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = binaryAnalyzer.ParseBytes(fetchResult.Content.([]byte), t)
			if err != nil {
				t.AnalyzeErr = err
				return
			}
		}

	}

	if analyzerResult != nil {

		// Check mtag
		mtag := analyzerResult.Mtag
		if len(mtag) > 0 {
			// Check exists mtag
			hash := t.GetHash()
			storedTarget, _ := store.GetStore().GetLatestTarget(hash)
			if storedTarget != nil {
				storedMtag := storedTarget["mtag"].(string)
				if mtag == storedMtag {
					return
				}
			}

			// Save mtag
			store.GetStore().InsertObject(store.TargetDataset,
				[]string{"hash", "mtag"},
				[]interface{}{hash, mtag})
		}

		// Crawl deeper target
		for _, t := range analyzerResult.Targets {
			self.targetsQueue.Push(t)
		}

		// Save crawled data to sink
		if self.sinkChanConnected {
			for _, pack := range analyzerResult.SinkPacks {
				self.sinkChan <- pack
			}
		}
	}

}
