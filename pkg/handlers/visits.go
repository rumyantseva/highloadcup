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

func (h *Handler) Visit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	fromCache := h.visit.Get(id)
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
	raw, err := txn.First("visit", "id", uid)
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

	var visit models.Visit
	visit = raw.(models.Visit)
	writeResponse(w, http.StatusOK, visit)
}

func (h *Handler) UserVisits(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	// Check user
	_, err = db.User(h.withdb.DB, uint(uid))
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

	country := r.URL.Query().Get("country")

	var toDistance *int
	sToDistance := r.URL.Query().Get("toDistance")
	if len(sToDistance) > 0 {
		iToDistance, err := strconv.Atoi(sToDistance)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}
		toDistance = &iToDistance
	}

	checker := NewVisitCheker(fromDate, toDate, toDistance, country)

	txn := h.withdb.DB.Txn(false)
	iter, err := txn.Get("visit", "user_id", uid)
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, nil)
		return
	}

	visits := make([]models.VisitExt, 0)
	for {
		raw := iter.Next()
		if raw == nil {
			break
		}
		visit := raw.(models.Visit)

		if checker.Check(h.withdb.DB, &visit) {
			loc, err := db.Location(h.withdb.DB, visit.Location)
			if err != nil {
				continue
			}

			visitExt := models.VisitExt{
				Mark:      visit.Mark,
				VisitedAt: visit.VisitedAt,
				Place:     loc.Place,
			}

			// order by visited at
			var i int
			for i = 0; i < len(visits); i++ {
				if visit.VisitedAt < visits[i].VisitedAt {
					break
				}
			}
			visits = append(visits[:i], append([]models.VisitExt{visitExt}, visits[i:]...)...)
		}
	}
	txn.Abort()

	data := struct {
		Visits []models.VisitExt `json:"visits"`
	}{
		Visits: visits,
	}

	writeResponse(w, http.StatusOK, data)
}

func (h *Handler) UpdateVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "new" {
		h.CreateVisit(w, r, ps)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	// Check if location exists
	visit, err := db.Visit(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	req := struct {
		User      *uint `json:"user,omitempty"`
		Location  *uint `json:"location,omitempty"`
		VisitedAt *int  `json:"visited_at,omitempty"`
		Mark      *int  `json:"mark,omitempty"`
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

	/*if req.User == nil || req.Location == nil || req.VisitedAt == nil ||
		req.Mark == nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	visit := models.Visit{
		User:      *req.User,
		Location:  *req.Location,
		VisitedAt: *req.VisitedAt,
		Mark:      *req.Mark,
		ID:        uint(uid),
	}()*/

	go func() {
		if req.User != nil {
			visit.User = *req.User
		}
		if req.Location != nil {
			visit.Location = *req.Location
		}
		if req.VisitedAt != nil {
			visit.VisitedAt = *req.VisitedAt
		}
		if req.Mark != nil {
			visit.Mark = *req.Mark
		}

		go h.visit.SetFrom(fmt.Sprint(visit.ID), visit)

		txn := h.withdb.DB.Txn(true)
		err = txn.Insert("visit", *visit)
		if err != nil {
			log.Print(err)
			writeResponse(w, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponse(w, http.StatusOK, struct{}{})
}

func (h *Handler) CreateVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := struct {
		User      *uint `json:"user,omitempty"`
		Location  *uint `json:"location,omitempty"`
		VisitedAt *int  `json:"visited_at,omitempty"`
		Mark      *int  `json:"mark,omitempty"`
		ID        *uint `json:"id,omitempty"`
	}{}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	if req.User == nil || req.Location == nil || req.VisitedAt == nil ||
		req.Mark == nil || req.ID == nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	go func() {
		visit := models.Visit{
			User:      *req.User,
			Location:  *req.Location,
			VisitedAt: *req.VisitedAt,
			Mark:      *req.Mark,
			ID:        *req.ID,
		}

		go h.visit.SetFrom(fmt.Sprint(visit.ID), visit)

		txn := h.withdb.DB.Txn(true)
		err = txn.Insert("visit", visit)
		if err != nil {
			log.Print(err)
			writeResponse(w, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponse(w, http.StatusOK, struct{}{})
}
