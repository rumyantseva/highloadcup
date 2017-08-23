package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
	"github.com/valyala/fasthttp"
)

func (h *Handler) Location(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)

	fromCache := h.location.Get(id)
	if fromCache != nil {
		writeResponseFromBytes(ctx, http.StatusOK, fromCache)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	txn := h.withdb.DB.Txn(false)
	raw, err := txn.First("location", "id", uid)
	if err != nil {
		log.Print(err)
		writeResponse(ctx, http.StatusInternalServerError, nil)
		return
	}
	txn.Abort()

	if raw == nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	var loc models.Location
	loc = raw.(models.Location)
	writeResponse(ctx, http.StatusOK, loc)
}

func (h *Handler) LocationMark(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		log.Printf("Couldn't parse id: %s", id)
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	// Check location
	_, err = db.Location(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	var fromDate *int
	sFromDate := string(ctx.URI().QueryArgs().Peek("fromDate"))
	if len(sFromDate) > 0 {
		iFromDate, err := strconv.Atoi(sFromDate)
		if err != nil {
			writeResponse(ctx, http.StatusBadRequest, nil)
			return
		}
		fromDate = &iFromDate
	}

	var toDate *int
	sToDate := string(ctx.URI().QueryArgs().Peek("toDate"))
	if len(sToDate) > 0 {
		iToDate, err := strconv.Atoi(sToDate)
		if err != nil {
			writeResponse(ctx, http.StatusBadRequest, nil)
			return
		}
		toDate = &iToDate
	}

	var fromAge *int
	sFromAge := string(ctx.URI().QueryArgs().Peek("fromAge"))
	if len(sFromAge) > 0 {
		iFromAge, err := strconv.Atoi(sFromAge)
		if err != nil {
			writeResponse(ctx, http.StatusBadRequest, nil)
			return
		}
		fromAge = &iFromAge
	}

	var toAge *int
	sToAge := string(ctx.URI().QueryArgs().Peek("toAge"))
	if len(sToAge) > 0 {
		iToAge, err := strconv.Atoi(sToAge)
		if err != nil {
			writeResponse(ctx, http.StatusBadRequest, nil)
			return
		}
		toAge = &iToAge
	}

	gender := string(ctx.URI().QueryArgs().Peek("gender"))
	if len(gender) > 0 && gender != "f" && gender != "m" {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	checker := NewLocationCheker(fromDate, toDate, fromAge, toAge, gender, h.current)

	txn := h.withdb.DB.Txn(false)
	iter, err := txn.Get("visit", "location_id", uid)
	if err != nil {
		log.Print(err)
		writeResponse(ctx, http.StatusInternalServerError, nil)
		return
	}

	visitors := 0
	totalMark := 0
	for {
		raw := iter.Next()
		if raw == nil {
			break
		}
		visit := raw.(models.Visit)

		if checker.Check(h.withdb.DB, &visit) {
			totalMark += visit.Mark
			visitors++
		}
	}
	txn.Abort()

	var avg float32
	if visitors > 0 {
		avg = float32(totalMark) / float32(visitors)
		avs := &Avg{Avg: avg}
		writeResponse(ctx, http.StatusOK, avs)
		return
	}

	writeResponse(ctx, http.StatusOK, map[string]int{"avg": 0})
}

func (h *Handler) UpdateLocation(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	if id == "new" {
		h.CreateLocation(ctx)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	// Check if location exists
	location, err := db.Location(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	req := struct {
		Distance *int    `json:"distance,omitempty"`
		City     *string `json:"city,omitempty"`
		Place    *string `json:"place,omitempty"`
		Country  *string `json:"country,omitempty"`
	}{}

	//defer ctx.PostBody().Close
	bodyString := string(ctx.PostBody())
	//log.Print(bodyString)

	// if body contains null, ignore it
	if strings.Contains(bodyString, "null") {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	err = json.Unmarshal(ctx.PostBody(), &req)
	//err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	/*if req.Distance == nil || req.City == nil || req.Place == nil ||
		req.Country == nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	location := models.Location{
		Distance: *req.Distance,
		City:     *req.City,
		Place:    *req.Place,
		Country:  *req.Country,
		ID:       uint(uid),
	}*/

	go func() {

		if req.Distance != nil {
			location.Distance = *req.Distance
		}
		if req.City != nil {
			location.City = *req.City
		}
		if req.Place != nil {
			location.Place = *req.Place
		}
		if req.Country != nil {
			location.Country = *req.Country
		}

		go h.location.SetFrom(fmt.Sprint(location.ID), location)

		txn := h.withdb.DB.Txn(true)
		err = txn.Insert("location", *location)
		if err != nil {
			log.Print(err)
			writeResponse(ctx, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponseFromBytes(ctx, http.StatusOK, []byte("{}"))
}

func (h *Handler) CreateLocation(ctx *fasthttp.RequestCtx) {
	req := struct {
		Distance *int    `json:"distance,omitempty"`
		City     *string `json:"city,omitempty"`
		Place    *string `json:"place,omitempty"`
		Country  *string `json:"country,omitempty"`
		ID       *uint   `json:"id,omitempty"`
	}{}

	//defer r.Body.Close()
	//err := json.NewDecoder(r.Body).Decode(&req)
	err := json.Unmarshal(ctx.PostBody(), &req)

	if err != nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	if req.Distance == nil || req.City == nil || req.Place == nil ||
		req.Country == nil || req.ID == nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	go func() {
		location := models.Location{
			Distance: *req.Distance,
			City:     *req.City,
			Place:    *req.Place,
			Country:  *req.Country,
			ID:       *req.ID,
		}

		go h.location.SetFrom(fmt.Sprint(location.ID), location)

		txn := h.withdb.DB.Txn(true)
		err = txn.Insert("location", location)
		if err != nil {
			log.Print(err)
			writeResponse(ctx, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponseFromBytes(ctx, http.StatusOK, []byte("{}"))
}
