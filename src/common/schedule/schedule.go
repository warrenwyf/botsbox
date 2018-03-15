package schedule

import (
	"github.com/petar/GoLLRB/llrb"
	"sync/atomic"
	"time"
)

type Schedule struct {
	startedAt time.Time

	tick time.Duration

	taskIdSeq       uint64
	taskMap         map[uint64]*task
	taskTree        *llrb.LLRB
	taskUpdatedChan chan *task

	paused     bool
	pauseChan  chan struct{}
	resumeChan chan struct{}
	stopChan   chan struct{}
}

var signal = struct{}{}

func New() *Schedule {
	return &Schedule{
		tick: time.Second,

		taskIdSeq:       0,
		taskMap:         make(map[uint64]*task),
		taskTree:        llrb.New(),
		taskUpdatedChan: make(chan *task),

		paused:     false,
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		stopChan:   make(chan struct{}),
	}
}

func (self *Schedule) CreateTask(fn func(), interval time.Duration, delay time.Duration) uint64 {
	now := time.Now()

	atomic.AddUint64(&self.taskIdSeq, 1)

	t := &task{
		id:       self.taskIdSeq,
		fn:       fn,
		interval: interval,
		delay:    delay,

		nextTime: now.Add(delay),

		updatedChan: self.taskUpdatedChan,
	}

	self.taskTree.InsertNoReplace(t)
	if self.taskTree.Has(t) {
		self.taskMap[t.id] = t
		return t.id
	}

	return 0
}

func (self *Schedule) DeleteTask(id uint64) bool {
	t, ok := self.taskMap[id]
	if ok {
		item := self.taskTree.Delete(t)
		if item != nil {
			delete(self.taskMap, id)
			return true
		} else {
			return false
		}
	}

	return false
}

func (self *Schedule) Start() {
	self.startedAt = time.Now()

	go self.loop()

	self.resumeChan <- signal // Skip the first pause in loop
}

func (self *Schedule) Pause() {
	self.pauseChan <- signal
}

func (self *Schedule) Resume() {
	self.resumeChan <- signal
}

func (self *Schedule) Stop() {
	if self.paused {
		self.resumeChan <- signal
	}

	self.stopChan <- signal
}

func (self *Schedule) loop() {
	var (
		fakeTask = &task{} // For quering in taskTree
		ticker   = time.NewTicker(self.tick)
	)

	defer ticker.Stop()

pause:
	self.paused = true
	<-self.resumeChan
	self.paused = false

	for {
		self.readTaskChan()

		select {
		case <-ticker.C:
			fakeTask.nextTime = time.Now()
			self.taskTree.DescendLessOrEqual(fakeTask, self.execTaskIter)

		case <-self.pauseChan:
			goto pause

		case <-self.stopChan:
			goto end
		}
	}

end:
}

func (self *Schedule) readTaskChan() {
	var updatedTask *task

	for {
		select {
		case updatedTask = <-self.taskUpdatedChan:
			self.taskTree.ReplaceOrInsert(updatedTask)

		default:
			goto end
		}
	}

end:
}

func (self *Schedule) execTaskIter(item llrb.Item) bool {
	t, ok := item.(*task)
	if !ok {
		return false
	}

	self.taskTree.Delete(t)
	go t.exec()

	return true
}
