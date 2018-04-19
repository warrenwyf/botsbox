package routers

import (
	"github.com/labstack/echo"

	"../../app"
	"../../crawler/job"
	"../../runtime"
	"../../store"
	"../../xlog"
)

const ApiPrefix = "/api"

func UseApiRouter(e *echo.Echo) {

	e.GET(joinPath(ApiPrefix, "/info"), func(c echo.Context) error {
		result := map[string]interface{}{
			"version": runtime.GetVersion(),
			"dataDir": runtime.GetAbsDataDir(),
			"logDir":  runtime.GetAbsLogDir(),
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.GET(joinPath(ApiPrefix, "/jobs"), func(c echo.Context) error {
		jobs := []map[string]interface{}{}

		jobObjs, err := store.GetStore().QueryAllJobs()
		if err == nil {
			for _, jobObj := range jobObjs {
				jobs = append(jobs, map[string]interface{}{
					"id":     jobObj["_id"],
					"title":  jobObj["title"],
					"status": jobObj["status"],
				})
			}
		}

		hub := app.GetHub()
		startedAt := hub.GetStartedAt()

		result := map[string]interface{}{
			"jobs":      jobs,
			"startedAt": startedAt.UTC().Unix(),
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.GET(joinPath(ApiPrefix, "/activeJobs"), func(c echo.Context) error {
		hub := app.GetHub()

		tasks := hub.GetAllJobs()
		jobs := []map[string]interface{}{}
		for _, task := range tasks {
			job := task.GetRunnable().(*job.Job)
			jobs = append(jobs, map[string]interface{}{
				"id":       job.GetId(),
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

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/createJob"), func(c echo.Context) error {
		title := c.Request().PostFormValue("title")
		rule := c.Request().PostFormValue("rule")

		id, err := store.GetStore().InsertObject(store.JobDataset,
			[]string{"title", "rule", "status"},
			[]interface{}{title, rule, "deactive"})

		code := 0
		if err != nil {
			code = 5001

			xlog.Errln("Create job error:", err)
		}

		result := map[string]interface{}{
			"code": code,
			"id":   id,
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.GET(joinPath(ApiPrefix, "/job/:id"), func(c echo.Context) error {
		id := c.Param("id")

		jobObj, err := store.GetStore().GetJob(id)
		if err != nil {
			return err
		}

		result := map[string]interface{}{
			"id":     jobObj["_id"],
			"title":  jobObj["title"],
			"rule":   jobObj["rule"],
			"status": jobObj["status"],
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/job/:id/active"), func(c echo.Context) error {
		id := c.Param("id")

		hub := app.GetHub()
		ok := hub.ActiveJob(id)

		code := 0
		if !ok {
			code = 5001
		}

		result := map[string]interface{}{
			"code": code,
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/job/:id/deactive"), func(c echo.Context) error {
		id := c.Param("id")

		hub := app.GetHub()
		ok := hub.DeactiveJob(id)

		code := 0
		if !ok {
			code = 5001
		}

		result := map[string]interface{}{
			"code": code,
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/job/:id/update"), func(c echo.Context) error {
		id := c.Param("id")
		title := c.Request().PostFormValue("title")
		rule := c.Request().PostFormValue("rule")

		code := 0

		_, err := store.GetStore().UpdateObject(store.JobDataset, id,
			[]string{"title", "rule"},
			[]interface{}{title, rule})

		if err != nil {
			code = 5001

			xlog.Errln("Update job error:", err)
		} else { // Deactive job after updating
			hub := app.GetHub()
			hub.DeactiveJob(id)
		}

		result := map[string]interface{}{
			"code": code,
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/job/:id/delete"), func(c echo.Context) error {
		id := c.Param("id")

		code := 0

		_, err := store.GetStore().DeleteObjects(store.JobDataset, []string{id})
		if err != nil {
			code = 5001

			xlog.Errln("Delete job error:", err)
		}

		result := map[string]interface{}{
			"code": code,
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/testrun"), func(c echo.Context) error {
		rule := c.Request().PostFormValue("rule")

		hub := app.GetHub()
		id, err := hub.TestrunJob(rule)

		code := 0

		if err != nil {
			code = 5001

			xlog.Errln("Testrun job error:", err)
		}

		result := map[string]interface{}{
			"code": code,
			"id":   id,
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.POST(joinPath(ApiPrefix, "/testrun/cancel"), func(c echo.Context) error {
		hub := app.GetHub()
		hub.CancelTestrunJob()

		result := map[string]interface{}{
			"code": 0,
		}

		return writeJsonResponse(c.Response(), result)
	})

}
