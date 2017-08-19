package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func (h *Handler) Location(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	txn := h.db.Txn(false)
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
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	// Check location
	_, err = db.Location(h.db, uint(uid))
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

	checker := NewLocationCheker(fromDate, toDate, fromAge, toAge, gender, h.current)

	txn := h.db.Txn(false)
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

		if checker.Check(h.db, &visit) {
			totalMark += visit.Mark
			visitors++
		}
	}
	txn.Abort()

	var avg float32
	if visitors > 0 {
		avg = float32(totalMark) / float32(visitors)
		writeResponse(w, http.StatusOK, map[string]float32{"avg": avg})
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
	/*	id := ps.ByName("id")
		uid, err := strconv.ParseUint(id, 10, 32)

		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}*/

}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}
