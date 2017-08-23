package main

import (
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/rumyantseva/highloadcup/pkg/cache"
	"github.com/rumyantseva/highloadcup/pkg/handlers"
	"github.com/valyala/fasthttp"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/data"
	"github.com/rumyantseva/highloadcup/pkg/db"
)

func main() {

	mem, err := memdb.NewMemDB(db.Schema())
	if err != nil {
		log.Fatal(err)
	}

	withdb := db.NewWithMax(mem)
	userCache := cache.NewStorage()
	locationCache := cache.NewStorage()
	visitCache := cache.NewStorage()

	imp := data.NewStorage(withdb, userCache, locationCache, visitCache)

	go func() {
		err = imp.Import()
		if err != nil {
			log.Fatal(err)
		}
	}()

	stamp, err := data.LocalTime()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("`Current` time is: %d", stamp)

	r := fasthttprouter.New()
	h := handlers.NewHandler(withdb, userCache, locationCache, visitCache, stamp)

	r.GET("/users/:id", h.User)
	r.GET("/locations/:id", h.Location)
	r.GET("/visits/:id", h.Visit)

	r.GET("/users/:id/visits", h.UserVisits)
	r.GET("/locations/:id/avg", h.LocationMark)

	r.POST("/users/:id", h.UpdateUser)
	r.POST("/locations/:id", h.UpdateLocation)
	r.POST("/visits/:id", h.UpdateVisit)

	log.Fatal(fasthttp.ListenAndServe(":80", r.Handler))
}
