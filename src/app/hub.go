package app

import (
	"sync"

	"../common/schedule"
	"../crawler/job"
	"../crawler/sink"
	"../store"
	"../xlog"
)

var (
	hubSingleton *Hub
	once         sync.Once
)

type Hub struct {
	sink        *sink.Sink
	jobSchedule *schedule.Schedule
}

func GetHub() *Hub {
	once.Do(func() {
		hubSingleton = &Hub{
			sink:        sink.NewSink(),
			jobSchedule: schedule.NewSchedule(),
		}
	})

	return hubSingleton
}

func (h *Hub) GetAllJobs() map[uint64]*schedule.Task {
	return h.jobSchedule.AllTasks()
}

func (h *Hub) Init() error {
	errStore := store.GetStore().Init()
	if errStore != nil {
		return errStore
	}

	h.sink.Open()

	h.jobSchedule.Start()

	return nil
}

func (h *Hub) Destroy() {
	h.jobSchedule.Stop()

	store.GetStore().Destroy()
}

func (h *Hub) LoadJobs() {
	jobObjs, err := store.GetStore().QueryAllJobs()
	if err != nil {
		xlog.Errln("Query jobs failed:", err)
		return
	}

	jobsCount := 0
	for _, jobObj := range jobObjs {
		if jobObj["status"].(string) != "alive" {
			continue
		}

		job, err := job.NewJob(jobObj["title"].(string), jobObj["rule"].(string))
		if err != nil {
			xlog.Errln("Load job failed:", jobObj, err)
			continue
		}

		job.ConnectSink(h.sink)

		taskId := h.jobSchedule.CreateTask(job)
		if taskId > 0 {
			jobsCount++
		}
	}

	xlog.Outln("Loaded", jobsCount, "jobs")
}
