package schedule

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/petar/GoLLRB/llrb"
)

type Schedule struct {
	startedAt time.Time

	tick time.Duration

	taskIdSeq       uint64
	taskMap         map[uint64]*Task
	taskMapMutex    sync.Mutex
	taskTree        *llrb.LLRB
	taskTreeMutex   sync.Mutex
	taskUpdatedChan chan *Task

	paused     bool
	pauseChan  chan struct{}
	resumeChan chan struct{}
	stopChan   chan struct{}
}

var signal = struct{}{}

func NewSchedule() *Schedule {
	return &Schedule{
		tick: time.Second,

		taskIdSeq:       0,
		taskMap:         map[uint64]*Task{},
		taskTree:        llrb.New(),
		taskUpdatedChan: make(chan *Task, 1000),

		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		stopChan:   make(chan struct{}),
	}
}

func (self *Schedule) CreateTask(runnable Runnable) uint64 {
	atomic.AddUint64(&self.taskIdSeq, 1)

	task := &Task{
		id: self.taskIdSeq,

		runnable: runnable,

		interval:    runnable.GetInterval(),
		nextTime:    time.Now().Add(runnable.GetDelay()),
		running:     false,
		updatedChan: self.taskUpdatedChan,
	}

	self.taskTreeMutex.Lock()
	self.taskTree.InsertNoReplace(task)
	ok := self.taskTree.Has(task)
	self.taskTreeMutex.Unlock()

	if ok {
		self.taskMapMutex.Lock()
		self.taskMap[task.id] = task
		self.taskMapMutex.Unlock()

		return task.id
	}

	return 0
}

func (self *Schedule) DeleteTask(id uint64) bool {
	task, ok := self.taskMap[id]
	if ok {
		self.taskTreeMutex.Lock()
		item := self.taskTree.Delete(task)
		self.taskTreeMutex.Unlock()

		if item != nil {
			self.taskMapMutex.Lock()
			delete(self.taskMap, id)
			self.taskMapMutex.Unlock()

			return true
		} else {
			return false
		}
	}

	return false
}

func (self *Schedule) Clear() {
	self.taskMapMutex.Lock()
	self.taskMap = map[uint64]*Task{}
	self.taskMapMutex.Unlock()

	self.taskTreeMutex.Lock()
	self.taskTree = llrb.New()
	self.taskTreeMutex.Unlock()
}

func (self *Schedule) AllTasks() map[uint64]*Task {
	return self.taskMap
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

/**
 * goroutine loop
 */
func (self *Schedule) loop() {
	var (
		fakeTask = &Task{} // For quering in taskTree
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

			tasks := []*Task{}
			self.taskTree.DescendLessOrEqual(fakeTask, func(item llrb.Item) bool {
				task, ok := item.(*Task)
				tasks = append(tasks, task)
				return ok
			})

			for _, task := range tasks {
				self.taskTreeMutex.Lock()
				self.taskTree.Delete(task)
				self.taskTreeMutex.Unlock()

				go task.run()
			}

		case <-self.pauseChan:
			goto pause

		case <-self.stopChan:
			goto end
		}
	}

end:
}

func (self *Schedule) readTaskChan() {
	for {
		select {
		case updatedTask := <-self.taskUpdatedChan:
			self.taskTreeMutex.Lock()
			self.taskTree.ReplaceOrInsert(updatedTask)
			self.taskTreeMutex.Unlock()

		default:
			goto end
		}
	}

end:
}
