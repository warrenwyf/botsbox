package testrun

import (
	"strconv"
	"sync/atomic"

	"../job"
)

type Runner struct {
	idSeq uint64

	runningJob *job.Job

	outputChan chan string
}

func NewRunner() *Runner {
	return &Runner{
		outputChan: make(chan string, 1000),
	}
}

func (self *Runner) Run(rule string) (uint64, error) {
	defer func() {
		recover()
	}()

	if self.runningJob != nil { // Interrupt running job
		self.runningJob.CancelTestrun()
	}

	id := atomic.AddUint64(&self.idSeq, 1)

	j, err := job.NewJob(strconv.FormatUint(id, 10), "Test Run", rule)
	if err != nil {
		return id, err
	}

	self.runningJob = j
	j.ConnectTestrunOutput(self.outputChan)

	go j.Testrun()

	return id, nil
}

func (self *Runner) CancelRunning() {
	if self.runningJob != nil { // Interrupt running job
		self.runningJob.CancelTestrun()
	}
}

func (self *Runner) GetOutputChan() chan string {
	return self.outputChan
}
