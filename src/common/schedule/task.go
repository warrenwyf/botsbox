package schedule

import (
	"github.com/petar/GoLLRB/llrb"
	"log"
	"time"
)

type task struct {
	id       uint64
	fn       func()
	interval time.Duration
	delay    time.Duration

	nextTime time.Time // The next time when task should be executed

	running     bool
	updatedChan chan<- *task // Write only
}

func (self *task) Less(item llrb.Item) bool { // llrb compare interface
	t := item.(*task)
	return self.nextTime.Before(t.nextTime)
}

func (self *task) exec() {
	if self.interval <= 0 || self.running {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Println("Recovering", err)
		}
	}()

	self.running = true

	self.fn()

	self.nextTime = self.nextTime.Add(self.interval)
	self.running = false

	self.updatedChan <- self
}

func (self *task) cancel() {
	self.interval = 0
}
