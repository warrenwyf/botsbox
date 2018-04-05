package web

import (
	"fmt"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"../config"
	"../xlog"
	"./routers"
)

var e = echo.New()

func Start() {
	conf := config.GetConf()

	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Static("src/web/static"))

	routers.UseApiRouter(e)

	err := e.Start(fmt.Sprintf(":%d", conf.HttpPort))
	if err != nil {
		xlog.Errln("Start web error", err)
		xlog.FlushAll()
		xlog.CloseAll()
		os.Exit(1)
	}
}
