package routers

import (
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"

	"../../app"
	"../../xlog"
)

func UseWsRouter(e *echo.Echo) {

	e.GET("/ws", func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			hub := app.GetHub()
			c := hub.GetTestrunOutput()

			for {
				output := <-c
				err := websocket.Message.Send(ws, output)
				if err != nil {
					xlog.Errln("WebSocket send error:", err)
				}
			}
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	})

}
