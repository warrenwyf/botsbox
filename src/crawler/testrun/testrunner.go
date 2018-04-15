package testrun

import (
	"strconv"
	"sync/atomic"

	"../job"
)

type Runner struct {
	idSeq uint64

	runningJob *job.Job

	OutputChan chan string
}

func NewRunner() *Runner {
	return &Runner{
		OutputChan: make(chan string),
	}
}

func (self *Runner) Run(rule string) (uint64, error) {
	if self.runningJob != nil { // Interrupt running job
		self.runningJob.CancelTestrun()
	}

	id := atomic.AddUint64(&self.idSeq, 1)

	j, err := job.NewJob(strconv.FormatUint(id, 10), "Test Run", rule)
	if err != nil {
		return id, err
	}

	self.runningJob = j
	j.ConnectTestrunOutput(self.OutputChan)

	go j.Testrun()

	return id, nil
}

func (self *Runner) CancelRunning() {
	if self.runningJob != nil { // Interrupt running job
		self.runningJob.CancelTestrun()
	}
}
