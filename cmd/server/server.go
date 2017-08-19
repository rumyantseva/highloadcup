package main

import (
	"log"
	"net/http"

	"github.com/rumyantseva/highloadcup/pkg/handlers"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/julienschmidt/httprouter"
	"github.com/rumyantseva/highloadcup/pkg/data"
	"github.com/rumyantseva/highloadcup/pkg/db"
)

func main() {

	mem, err := memdb.NewMemDB(db.Schema())
	if err != nil {
		log.Fatal(err)
	}

	var stamp int
	err = data.Import(mem, &stamp)
	if err != nil {
		log.Fatal(err)
	}

	r := httprouter.New()
	h := handlers.NewHandler(mem)

	r.GET("/users/:id", h.User)
	r.GET("/locations/:id", h.Location)
	r.GET("/visits/:id", h.Visit)

	r.GET("/users/:id/visits", h.UserVisits)
	r.GET("/locations/:id/avg", h.LocationMark)

	r.POST("/users/:id", h.UpdateUser)
	r.POST("/locations/:id", h.UpdateLocation)
	r.POST("/visits/:id", h.UpdateVisit)

	log.Fatal(http.ListenAndServe(":80", r))
}
