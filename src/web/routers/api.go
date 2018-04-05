package routers

import (
	"fmt"

	"github.com/labstack/echo"

	"../../app"
	"../../crawler/job"
	"../../runtime"
)

const ApiPrefix = "/api"

func UseApiRouter(e *echo.Echo) {

	e.GET(joinPath(ApiPrefix, "/info"), func(c echo.Context) error {
		result := map[string]interface{}{
			"version": fmt.Sprintf("%d.%d.%d", runtime.VersionMajor, runtime.VersionMinor, runtime.VersionPatch),
			"dataDir": runtime.GetAbsDataDir(),
			"logDir":  runtime.GetAbsLogDir(),
		}

		return writeJsonResponse(c.Response(), result)
	})

	e.GET(joinPath(ApiPrefix, "/jobs"), func(c echo.Context) error {
		hub := app.GetHub()

		tasks := hub.GetAllJobs()
		jobs := []map[string]interface{}{}
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

		return writeJsonResponse(c.Response(), result)
	})

}
