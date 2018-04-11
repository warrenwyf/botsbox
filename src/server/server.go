package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"../app"
	"../config"
	"../web"
	"../xlog"
)

var s *server

func Start() error {
	if s != nil {
		s.destroy()
	}

	s = &server{
		sigChan: make(chan os.Signal),
		hub:     app.GetHub(),
	}

	return s.start()
}

type server struct {
	sigChan chan os.Signal
	hub     *app.Hub
}

func (self *server) start() error {
	h := self.hub
	if h == nil {
		return errors.New("App initalized error")
	}

	errInit := h.Init()
	if errInit != nil {
		return errInit
	}

	go self.startDebug()

	go web.Start()

	go h.LoadJobs() // Load exsiting jobs from store

	xlog.Outln("Server started")

	// Wait for system signal
	signal.Notify(self.sigChan, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGABRT, syscall.SIGSEGV, syscall.SIGBUS, syscall.SIGILL)
	for {
		sig := <-self.sigChan

		switch sig {
		case syscall.SIGINT:
			goto end
		case syscall.SIGTERM:
			goto end
		default:
			xlog.Errln(sig)
		}
	}

end:
	self.destroy()

	xlog.Outln("Server stopped")
	xlog.FlushAll()
	xlog.CloseAll()

	return nil
}

func (self *server) destroy() {
	h := self.hub
	if h != nil {
		h.Destroy()
	}

	close(self.sigChan)
}

func (self *server) startDebug() {
	conf := config.GetConf()

	if conf.DebugPort > 0 {
		err := http.ListenAndServe(fmt.Sprintf(":%d", conf.DebugPort), nil)
		if err != nil {
			xlog.Errln("Start debug error", err)
			xlog.FlushAll()
			xlog.CloseAll()
			os.Exit(1)
		}
	}
}
