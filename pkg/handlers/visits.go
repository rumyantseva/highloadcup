package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func (h *Handler) Visit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	txn := h.db.Txn(false)
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
	_, err = db.User(h.db, uint(uid))
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

	txn := h.db.Txn(false)
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

		if checker.Check(h.db, &visit) {
			loc, err := db.Location(h.db, visit.Location)
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

	/*	id := ps.ByName("id")
		uid, err := strconv.ParseUint(id, 10, 32)

		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}
	*/
}

func (h *Handler) CreateVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}
