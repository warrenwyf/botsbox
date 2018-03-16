package server

import (
	"fmt"
	"log"
	"net/http"

	"../config"
)

func Start() bool {
	conf := config.GetConf()
	conf.SyncFromFile("./config.json")

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

	http.HandleFunc("/", hub.HttpHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.HttpPort), nil)
	if err != nil {
		log.Fatalln("Listen http error: ", err.Error())
		return false
	}

	return true
}
