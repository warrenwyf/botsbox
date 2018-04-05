package job

import (
	"errors"
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

	targetsQueue      *queue.PriorityQueue
	targetsInCrawling map[uint64]*target.Target
	targetCrawledChan chan *target.Target

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
		targetCrawledChan: make(chan *target.Target, 100),

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
				self.startCrawlTarget(t)
			}
		}

		// Notice: runningCount or waitingCount may changed
		if len(self.targetsInCrawling) > 0 {
			t := <-self.targetCrawledChan

			self.crawledTargetsCount++
			self.analyze(t)
		}

		elapse := time.Now().Sub(self.runAt)
		if elapse > self.timeout {
			xlog.Outf("Job \"%s\" did not finish crawling, timeout\n", self.title)
			break
		} else if elapse > self.interval && self.crawledTargetsCount == 0 {
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

func (self *Job) startCrawlTarget(t *target.Target) {
	if t != nil {
		self.targetsInCrawling[t.GetId()] = t

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
		if err != nil {
			xlog.Errln("Crawl target error:", err)

			if t.CanRetry() {
				self.retryCrawlTarget(t)
			}
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

	resultType := strings.ToLower(t.ResultType)
	if resultType == "html" {
		htmlAnalyzer := analyzers.NewHtmlAnalyzer(self.rule)

		if result.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = htmlAnalyzer.ParseBytes(result.Content.([]byte), result.ContentType, t)
			if err != nil {
				return
			}
		}

	} else if resultType == "json" {
		jsonAnalyzer := analyzers.NewJsonAnalyzer(self.rule)

		if result.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = jsonAnalyzer.ParseBytes(result.Content.([]byte), t)
			if err != nil {
				return
			}
		}

	} else if resultType == "xml" {
		xmlAnalyzer := analyzers.NewXmlAnalyzer(self.rule)

		if result.Format == fetchers.ResultFormat_Bytes {
			analyzerResult, err = xmlAnalyzer.ParseBytes(result.Content.([]byte), result.ContentType, t)
			if err != nil {
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

		analyzerResult, err = binaryAnalyzer.Parse(result.Content.([]byte), t)
		if err != nil {
			return
		}

	}

	if analyzerResult != nil {
		mtag := analyzerResult.Mtag
		if len(mtag) > 0 {
			// Check exists mtag
			hash := t.GetHash()
			storedTarget, _ := store.GetStore().GetLatestTarget(hash)
			if storedTarget != nil {
				storedMtag := storedTarget["mtag"].(string)
				if mtag == storedMtag {
					goto end
				}
			}

			// Save mtag
			store.GetStore().InsertObject(store.TargetDataset,
				[]string{"hash", "mtag"},
				[]interface{}{hash, mtag})
		}

		// Crawl deeper target
		for _, t := range analyzerResult.Targets {
			self.startCrawlTarget(t)
		}

		// Save crawled data to sink
		if self.sinkChanConnected {
			for _, pack := range analyzerResult.SinkPacks {
				self.sinkChan <- pack
			}
		}
	}

end:
	delete(self.targetsInCrawling, t.GetId())
}
