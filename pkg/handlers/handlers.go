package handlers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rumyantseva/highloadcup/pkg/db"
)

type Handler struct {
	withdb  *db.WithMax
	current int
}

func NewHandler(withmax *db.WithMax, current int) *Handler {
	return &Handler{
		withdb:  withmax,
		current: current,
	}
}

func writeResponse(w http.ResponseWriter, code int, resp interface{}) {
	if resp == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Couldn't encode response %+v to HTTP response body.", resp)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//	len, _ := w.Write(data)

	len := binary.Size(data)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len))
	w.WriteHeader(code)
	w.Write(data)

	/*	enc := json.NewEncoder(w)
		err = enc.Encode(resp)
		if err != nil {
			log.Printf("Couldn't encode response %+v to HTTP response body.", resp)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}*/
}
