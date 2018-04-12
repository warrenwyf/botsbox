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

}
