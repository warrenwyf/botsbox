package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"../common/schedule"
	"../config"
	"../crawler/job"
	"../crawler/sink"
	"../runtime"
	"../store"
	"../xlog"
)

type hub struct {
	sink        *sink.Sink
	jobSchedule *schedule.Schedule
}

func newHub() *hub {
	h := &hub{
		sink:        sink.NewSink(),
		jobSchedule: schedule.NewSchedule(),
	}

	return h
}

func (h *hub) init() error {
	errStore := store.GetStore().Init()
	if errStore != nil {
		return errStore
	}

	h.sink.Open()

	h.jobSchedule.Start()

	return nil
}

func (h *hub) destroy() {
	h.jobSchedule.Stop()

	store.GetStore().Destroy()
}

func (h *hub) listenHttp() {
	conf := config.GetConf()

	http.HandleFunc("/", h.httpHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.HttpPort), nil)
	if err != nil {
		xlog.Errln("Listern HTTP error", err)
		xlog.FlushAll()
		xlog.CloseAll()
		os.Exit(1)
	}
}

func (h *hub) loadJobs() {
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

func (h *hub) allJobTasks() map[uint64]*schedule.Task {
	return h.jobSchedule.AllTasks()
}

func (h *hub) httpHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI

	if uri == "/info" {
		result := map[string]interface{}{
			"version": fmt.Sprintf("%d.%d.%d", runtime.VersionMajor, runtime.VersionMinor, runtime.VersionPatch),
			"dataDir": runtime.GetAbsDataDir(),
			"logDir":  runtime.GetAbsLogDir(),
		}

		h.writeJsonResponse(w, result)
		return

	} else if uri == "/jobs" {
		jobs := []map[string]interface{}{}

		tasks := h.allJobTasks()
		for _, task := range tasks {
			job := task.GetRunnable().(*job.Job)
			jobs = append(jobs, map[string]interface{}{
				"title":    job.GetTitle(),
				"interval": job.GetInterval().Seconds(),
				"runAt":    job.GetRunAt().UTC().Unix(),
				"crawled":  job.GetCrawledTargetsCount(),

				"next":    task.GetNextTime().UTC().Unix(),
				"running": task.IsRunning(),
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
