package rule

import (
	"time"

	"github.com/tidwall/gjson"

	"../../common/util"
)

type TargetTemplate struct {
	Age       time.Duration
	Priority  int64
	Timeout   time.Duration
	Retry     int64
	RetryWait time.Duration
	Mtag      string
	Client    string

	Dive map[string]*Entry

	ObjectOutputs []*ObjectOutput
	ListOutputs   []*ListOutput
}

func NewTargetTemplateWithJson(elem *gjson.Result) *TargetTemplate {
	t := &TargetTemplate{
		Age:       24 * time.Hour,
		Priority:  0,
		Timeout:   0,
		Retry:     3,
		RetryWait: time.Minute,

		Dive: map[string]*Entry{},

		ObjectOutputs: []*ObjectOutput{},
		ListOutputs:   []*ListOutput{},
	}

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

	timeoutElem := elem.Get("$timeout")
	if timeoutElem.Exists() {
		timeout, err := util.ParseDuration(timeoutElem.String())
		if err == nil {
			t.Timeout = timeout
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

	mtagElem := elem.Get("$mtag")
	if mtagElem.Exists() {
		t.Mtag = mtagElem.String()
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
