package routers

import (
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"time"

	"../../xlog"
)

func UseWsRouter(e *echo.Echo) {

	e.GET("/ws", func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			for {
				err := websocket.Message.Send(ws, "Hello, Client!")
				if err != nil {
					xlog.Errln("WebSocket send error:", err)
				}

				t := time.NewTimer(5 * time.Second)
				<-t.C
			}
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	})

}
