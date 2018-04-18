package web

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
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
	e.Logger.SetOutput(ioutil.Discard)

	e.Use(middleware.Recover())
	e.Use(middleware.Static("src/web/static"))

	e.Renderer = &HtmlRenderer{
		templates: template.Must(template.ParseGlob("src/web/views/*.html")),
	}

	routers.UseRootRouter(e)
	routers.UseApiRouter(e)
	routers.UseWsRouter(e)

	err := e.Start(fmt.Sprintf(":%d", conf.HttpPort))
	if err != nil {
		xlog.Errln("Start web error", err)
		xlog.FlushAll()
		xlog.CloseAll()
		os.Exit(1)
	}
}

type HtmlRenderer struct {
	templates *template.Template
}

func (self *HtmlRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return self.templates.ExecuteTemplate(w, name, data)
}
