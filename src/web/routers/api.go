package routers

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"github.com/labstack/echo"
	"github.com/tidwall/gjson"

	"../../app"
	"../../crawler/job"
	"../../crawler/rule"
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

		// Parse outputs names
		outputNames := []string{}
		if r, err := rule.NewRuleWithContent(jobObj["rule"].(string)); err == nil {
			for _, tt := range r.TargetTemplates {
				for _, output := range tt.ObjectOutputs {
					outputNames = append(outputNames, output.Name)
				}

				for _, output := range tt.ListOutputs {
					outputNames = append(outputNames, output.Name)
				}
			}
		}

		result := map[string]interface{}{
			"id":      jobObj["_id"],
			"title":   jobObj["title"],
			"rule":    jobObj["rule"],
			"status":  jobObj["status"],
			"outputs": outputNames,
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

	e.GET(joinPath(ApiPrefix, "/data/:dataset/export.:format"), func(c echo.Context) error {
		dataset := c.Param("dataset")
		format := strings.ToLower(c.Param("format"))

		dataObjs, err := store.GetStore().QueryAllDataObjects(dataset)
		if err != nil {
			return err
		}

		fieldsMap := map[string]struct{}{}
		for _, dataObj := range dataObjs {
			obj, ok := gjson.Parse(dataObj["data"].(string)).Value().(map[string]interface{})
			if !ok {
				continue
			}

			id, ok := dataObj["id"]
			if ok {
				idStr := id.(string)
				if len(idStr) > 0 {
					obj["id"] = idStr
				}
			}

			for k, _ := range obj {
				_, ok := fieldsMap[k]
				if !ok {
					fieldsMap[k] = struct{}{}
				}
			}
		}

		fields := []string{}
		for field, _ := range fieldsMap {
			fields = append(fields, field)
		}
		sort.Strings(fields)

		if format == "csv" {
			w := csv.NewWriter(c.Response())
			if err := w.Write(fields); err != nil {
				return err
			}

			for _, dataObj := range dataObjs {
				obj, ok := gjson.Parse(dataObj["data"].(string)).Value().(map[string]interface{})
				if !ok {
					continue
				}

				id, ok := dataObj["id"]
				if ok {
					obj["id"] = id
				}

				values := []string{}
				for _, field := range fields {
					value, ok := obj[field]
					if ok {
						values = append(values, fmt.Sprintf("%v", value))
					} else {
						values = append(values, "")
					}
				}

				if err := w.Write(values); err != nil {
					return err
				}
			}

			w.Flush()
		}

		return nil
	})

	e.POST(joinPath(ApiPrefix, "/data/:dataset/empty"), func(c echo.Context) error {
		dataset := c.Param("dataset")

		code := 0

		if err := store.GetStore().EmptyDataset(dataset); err != nil {
			code = 5001

			xlog.Errln("Empty dataset error:", err)
		}

		result := map[string]interface{}{
			"code": code,
		}

		return writeJsonResponse(c.Response(), result)
	})

}
