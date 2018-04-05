package schedule

import (
	"github.com/petar/GoLLRB/llrb"
	"time"
)

type Runnable interface {
	GetTitle() string
	GetFn() func()
	GetInterval() time.Duration
	GetDelay() time.Duration
}

type Task struct {
	id uint64

	runnable Runnable

	interval    time.Duration
	nextTime    time.Time // The next time when task should run
	running     bool
	updatedChan chan<- *Task // Write only
}

func (self *Task) GetId() uint64 {
	return self.id
}

func (self *Task) GetRunnable() Runnable {
	return self.runnable
}

func (self *Task) GetInterval() time.Duration {
	return self.interval
}

func (self *Task) GetNextTime() time.Time {
	return self.nextTime
}

func (self *Task) IsRunning() bool {
	return self.running
}

func (self *Task) Less(item llrb.Item) bool { // llrb compare interface
	t := item.(*Task)
	return self.nextTime.Before(t.nextTime)
}

func (self *Task) run() {
	if self.interval <= 0 || self.running {
		return
	}

	go func() {
		defer func() {
			recover()
		}()

		self.running = true

		fn := self.runnable.GetFn()
		if fn != nil {
			fn()
		}

		// Calculate next time and make sure the time is after now
		now := time.Now()
		next := self.nextTime.Add(self.interval)
		if now.After(next) {
			next = now
		}
		self.nextTime = next

		self.running = false

		self.updatedChan <- self
	}()
}

func (self *Task) cancel() {
	self.interval = 0
}
