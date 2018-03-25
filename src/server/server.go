package server

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"../xlog"
)

func Start() error {
	s := &server{
		sigChan: make(chan os.Signal),
		hub:     newHub(),
	}

	return s.start()
}

type server struct {
	sigChan chan os.Signal
	hub     *hub
}

func (self *server) start() error {
	h := self.hub
	if h == nil {
		return errors.New("Server initalized error")
	}

	errInit := h.init()
	if errInit != nil {
		return errInit
	}

	go h.loadJobs() // Load exsiting jobs from store

	go h.listenHttp()

	xlog.Outln("Server started")

	// Wait for system signal
	signal.Notify(self.sigChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-self.sigChan:
			switch sig {
			case syscall.SIGINT:
				goto end
			case syscall.SIGTERM:
				goto end
			}
		}
	}

end:
	h.destroy()

	xlog.Outln("Server stopped")
	xlog.FlushAll()
	xlog.CloseAll()

	return nil
}
