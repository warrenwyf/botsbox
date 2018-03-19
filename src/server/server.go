package server

import (
	"fmt"
	"log"
	"net/http"

	"../config"
)

func Start() bool {
	s := &server{
		hub: newHub(),
	}

	return s.start()
}

type server struct {
	hub *hub
}

func (srv *server) start() bool {
	conf := config.GetConf()

	hub := srv.hub
	if hub == nil {
		log.Fatalln("Server initalized error")
		return false
	}

	errInit := hub.init()
	if errInit != nil {
		log.Fatalln("Hub initalized error: ", errInit.Error())
		return false
	}

	go hub.loadJobs() // Load exsiting jobs from store

	http.HandleFunc("/", hub.httpHandler)

	errHttp := http.ListenAndServe(fmt.Sprintf(":%d", conf.HttpPort), nil)
	if errHttp != nil {
		log.Fatalln("Listen http error: ", errHttp.Error())
		return false
	}

	return true
}
