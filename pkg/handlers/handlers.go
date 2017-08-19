package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	memdb "github.com/hashicorp/go-memdb"
)

type Handler struct {
	db *memdb.MemDB
}

func NewHandler(db *memdb.MemDB) *Handler {
	return &Handler{
		db: db,
	}
}

func writeResponse(w http.ResponseWriter, code int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if resp == nil {
		return
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(resp)
	if err != nil {
		log.Printf("Couldn't encode response %+v to HTTP response body.", resp)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
