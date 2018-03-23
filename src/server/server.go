package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"../config"
	"../xlog"
)

var signal = struct{}{}

func Start() error {
	s := &server{
		stopChan: make(chan struct{}),
		hub:      newHub(),
	}

	return s.start()
}

type server struct {
	stopChan chan struct{}
	hub      *hub
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
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		if uri == "/stop" {
			self.stop()

			io.WriteString(w, "stopping")
			return
		}
	})
	go self.listenHttp()

	// Wait for stop signal
	for {
		select {
		case <-self.stopChan:
			goto end
		}
	}

end:

	return nil
}

func (self *server) stop() {
	xlog.FlushAll()
	xlog.CloseAll()

	self.stopChan <- signal
}

func (self *server) listenHttp() {
	conf := config.GetConf()

	errHttp := http.ListenAndServe(fmt.Sprintf(":%d", conf.HttpPort), nil)
	if errHttp != nil {
		xlog.Errln("Listern HTTP error", errHttp)
		os.Exit(1)
	}
}
