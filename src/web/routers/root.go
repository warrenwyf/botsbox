package routers

import (
	"net/http"

	"github.com/labstack/echo"

	"../../runtime"
)

func UseRootRouter(e *echo.Echo) {

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"version": runtime.GetVersion(),
		})
	})

	e.GET("/create-job", func(c echo.Context) error {
		return c.Render(http.StatusOK, "job-editor.html", map[string]interface{}{
			"version": runtime.GetVersion(),
		})
	})

	e.GET("/job/:id", func(c echo.Context) error {
		jobId := c.Param("id")

		return c.Render(http.StatusOK, "job-editor.html", map[string]interface{}{
			"version": runtime.GetVersion(),
			"jobId":   jobId,
		})
	})

	e.GET("/job/:id/outputs", func(c echo.Context) error {
		jobId := c.Param("id")

		return c.Render(http.StatusOK, "job-outputs.html", map[string]interface{}{
			"version": runtime.GetVersion(),
			"jobId":   jobId,
		})
	})

}
