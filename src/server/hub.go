package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"../config"
)

type hub struct {
}

func newHub() *hub {
	return &hub{}
}

func (h *hub) HttpHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI

	if uri == "/info" {
		result := map[string]interface{}{
			"version": fmt.Sprintf("%d.%d.%d", config.VersionMajor, config.VersionMinor, config.VersionPatch),
		}

		h.writeJsonResponse(w, result)
		return
	} else if uri == "/job/create" {
	}
}

func (h *hub) writeJsonResponse(w http.ResponseWriter, v interface{}) error {
	b, errMarshal := json.Marshal(v)
	if errMarshal != nil {
		return errMarshal
	}

	_, err := io.WriteString(w, string(b))
	return err
}
