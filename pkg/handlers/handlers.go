package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	memdb "github.com/hashicorp/go-memdb"
)

type Handler struct {
	db      *memdb.MemDB
	current int
}

func NewHandler(db *memdb.MemDB, current int) *Handler {
	return &Handler{
		db:      db,
		current: current,
	}
}

func writeResponse(w http.ResponseWriter, code int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if resp == nil {
		return
	}

	enc := json.NewEncoder(w)
	err := enc.Encode(resp)
	if err != nil {
		log.Printf("Couldn't encode response %+v to HTTP response body.", resp)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
