package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func (h *Handler) Location(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	fromCache := h.location.Get(id)
	if fromCache != nil {
		writeResponseFromBytes(w, http.StatusOK, fromCache)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	txn := h.withdb.DB.Txn(false)
	raw, err := txn.First("location", "id", uid)
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, nil)
		return
	}
	txn.Abort()

	if raw == nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	var loc models.Location
	loc = raw.(models.Location)
	writeResponse(w, http.StatusOK, loc)
}

func (h *Handler) LocationMark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		log.Printf("Couldn't parse id: %s", id)
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	// Check location
	_, err = db.Location(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	var fromDate *int
	sFromDate := r.URL.Query().Get("fromDate")
	if len(sFromDate) > 0 {
		iFromDate, err := strconv.Atoi(sFromDate)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}
		fromDate = &iFromDate
	}

	var toDate *int
	sToDate := r.URL.Query().Get("toDate")
	if len(sToDate) > 0 {
		iToDate, err := strconv.Atoi(sToDate)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}
		toDate = &iToDate
	}

	var fromAge *int
	sFromAge := r.URL.Query().Get("fromAge")
	if len(sFromAge) > 0 {
		iFromAge, err := strconv.Atoi(sFromAge)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}
		fromAge = &iFromAge
	}

	var toAge *int
	sToAge := r.URL.Query().Get("toAge")
	if len(sToAge) > 0 {
		iToAge, err := strconv.Atoi(sToAge)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}
		toAge = &iToAge
	}

	gender := r.URL.Query().Get("gender")
	if len(gender) > 0 && gender != "f" && gender != "m" {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	checker := NewLocationCheker(fromDate, toDate, fromAge, toAge, gender, h.current)

	txn := h.withdb.DB.Txn(false)
	iter, err := txn.Get("visit", "location_id", uid)
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, nil)
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
		writeResponse(w, http.StatusOK, avs)
		return
	}

	writeResponse(w, http.StatusOK, map[string]int{"avg": 0})
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "new" {
		h.CreateLocation(w, r, ps)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	// Check if location exists
	location, err := db.Location(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	req := struct {
		Distance *int    `json:"distance,omitempty"`
		City     *string `json:"city,omitempty"`
		Place    *string `json:"place,omitempty"`
		Country  *string `json:"country,omitempty"`
	}{}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	bodyString := string(body)
	//log.Print(bodyString)

	// if body contains null, ignore it
	if strings.Contains(bodyString, "null") {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	err = json.Unmarshal(body, &req)
	//err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
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
			writeResponse(w, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponseFromBytes(w, http.StatusOK, []byte("{}"))
}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := struct {
		Distance *int    `json:"distance,omitempty"`
		City     *string `json:"city,omitempty"`
		Place    *string `json:"place,omitempty"`
		Country  *string `json:"country,omitempty"`
		ID       *uint   `json:"id,omitempty"`
	}{}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	if req.Distance == nil || req.City == nil || req.Place == nil ||
		req.Country == nil || req.ID == nil {
		writeResponse(w, http.StatusBadRequest, nil)
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
			writeResponse(w, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponseFromBytes(w, http.StatusOK, []byte("{}"))
}
