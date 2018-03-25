package schedule

import (
	"github.com/petar/GoLLRB/llrb"
	"time"
)

type Task struct {
	id uint64

	title    string
	fn       func()
	interval time.Duration
	delay    time.Duration

	nextTime time.Time // The next time when task should be executed

	executing   bool
	updatedChan chan<- *Task // Write only
}

func (self *Task) GetTitle() string {
	return self.title
}

func (self *Task) GetInterval() time.Duration {
	return self.interval
}

func (self *Task) GetNextTime() time.Time {
	return self.nextTime
}

func (self *Task) IsExecuting() bool {
	return self.executing
}

func (self *Task) Less(item llrb.Item) bool { // llrb compare interface
	t := item.(*Task)
	return self.nextTime.Before(t.nextTime)
}

func (self *Task) exec() {
	if self.interval <= 0 || self.executing {
		return
	}

	go func() {
		defer func() {
			recover()
		}()

		self.executing = true

		self.fn()

		// Calculate next time and make sure the time is after now
		now := time.Now()
		next := self.nextTime.Add(self.interval)
		if now.After(next) {
			next = now
		}
		self.nextTime = next

		self.executing = false

		self.updatedChan <- self
	}()
}

func (self *Task) cancel() {
	self.interval = 0
}
