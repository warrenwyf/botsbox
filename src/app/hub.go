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
		if jobObj["status"].(string) != "active" {
			continue
		}

		job, err := job.NewJob(jobObj["_id"].(string), jobObj["title"].(string), jobObj["rule"].(string))
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

func (h *Hub) ActiveJob(id string) bool {
	jobObj, err := store.GetStore().GetJob(id)
	if err != nil {
		xlog.Errln("Get job failed:", err)
		return false
	}

	ok := false

	if jobObj != nil {
		job, err := job.NewJob(jobObj["_id"].(string), jobObj["title"].(string), jobObj["rule"].(string))
		if err != nil {
			xlog.Errln("Load job failed:", jobObj, err)
			return false
		}

		job.ConnectSink(h.sink)

		taskId := h.jobSchedule.CreateTask(job)
		if taskId > 0 {
			xlog.Outln("Job", id, "actived")
			ok = true
		}

		// Update store
		store.GetStore().UpdateObject(store.JobDataset, id, []string{"status"}, []interface{}{"active"})
	}

	return ok
}

func (h *Hub) DeactiveJob(id string) bool {
	ok := false

	job := h.jobSchedule.GetTaskByRunnableId(id)
	if job != nil {
		ok = h.jobSchedule.DeleteTask(job.GetId())
		if ok {
			xlog.Outln("Job", id, "deactived")
		}
	}

	// Update store
	store.GetStore().UpdateObject(store.JobDataset, id, []string{"status"}, []interface{}{"deactive"})

	return ok
}
