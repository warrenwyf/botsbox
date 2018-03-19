package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"../common/schedule"
	"../common/store"
	"../crawler/job"
	"../runtime"
)

type hub struct {
	store    store.Store
	schedule *schedule.Schedule
}

func newHub() *hub {
	h := &hub{
		store:    store.NewStore(),
		schedule: schedule.NewSchedule(),
	}

	return h
}

func (h *hub) init() error {
	if h.store == nil {
		return errors.New("Store is null")
	}

	errStore := h.store.Init()
	if errStore != nil {
		return errStore
	}

	return nil
}

func (h *hub) loadJobs() {
	jobObjs, err := h.store.QueryAllJobs()
	if err != nil {
		log.Println("Query jobs failed:", err)
		return
	}

	for _, jobObj := range jobObjs {
		job, err := job.NewJob(jobObj["title"].(string), jobObj["rule"].(string))
		if err != nil {
			log.Println("Load job failed:", jobObj, err)
			continue
		}

		h.schedule.CreateTask(job.Title, job.Fn, job.Interval, job.Delay)
	}
}

func (h *hub) allJobTasks() map[uint64]*schedule.Task {
	return h.schedule.AllTasks()
}

func (h *hub) httpHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI

	if uri == "/info" {
		result := map[string]interface{}{
			"version": fmt.Sprintf("%d.%d.%d", runtime.VersionMajor, runtime.VersionMinor, runtime.VersionPatch),
		}

		h.writeJsonResponse(w, result)
		return

	} else if uri == "/jobs" {
		jobs := []map[string]interface{}{}

		tasks := h.allJobTasks()
		for _, task := range tasks {
			jobs = append(jobs, map[string]interface{}{
				"title":     task.GetTitle(),
				"interval":  task.GetInterval().Seconds(),
				"next":      task.GetNextTime().UTC().Unix(),
				"executing": task.IsExecuting(),
			})
		}

		result := map[string]interface{}{
			"jobs": jobs,
		}

		h.writeJsonResponse(w, result)
		return

	}
}

func (h *hub) writeJsonResponse(w http.ResponseWriter, v interface{}) error {
	b, errMarshal := json.Marshal(v)
	if errMarshal != nil {
		return errMarshal
	}

	_, err := io.WriteString(w, string(b))
	return err
}
