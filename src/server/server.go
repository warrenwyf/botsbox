package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"../config"
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
	hub := self.hub
	if hub == nil {
		return errors.New("Server initalized error")
	}

	errInit := hub.init()
	if errInit != nil {
		return errInit
	}

	go hub.loadJobs() // Load exsiting jobs from store

	http.HandleFunc("/", hub.httpHandler)
	go self.listenHttp()

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
	xlog.Outln("Server stopped")
	xlog.FlushAll()
	xlog.CloseAll()

	return nil
}

func (self *server) listenHttp() {
	conf := config.GetConf()

	errHttp := http.ListenAndServe(fmt.Sprintf(":%d", conf.HttpPort), nil)
	if errHttp != nil {
		xlog.Errln("Listern HTTP error", errHttp)
		os.Exit(1)
	}
}
