package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func (h *Handler) User(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	txn := h.withdb.DB.Txn(false)
	raw, err := txn.First("user", "id", uid)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	txn.Abort()

	if raw == nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	var user models.User
	user = raw.(models.User)
	writeResponse(w, http.StatusOK, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "new" {
		h.CreateUser(w, r, ps)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	// Check if user exists
	_, err = db.User(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	req := struct {
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		BirthDate *int    `json:"birth_date,omitempty"`
		Gender    *string `json:"gender,omitempty"`
		Email     *string `json:"email,omitempty"`
	}{}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	if req.FirstName == nil || req.LastName == nil || req.BirthDate == nil ||
		req.Gender == nil || req.Email == nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	user := models.User{
		FirstName: *req.FirstName,
		LastName:  *req.LastName,
		BirthDate: *req.BirthDate,
		Gender:    *req.Gender,
		Email:     *req.Email,
		ID:        uint(uid),
	}

	txn := h.withdb.DB.Txn(true)
	err = txn.Insert("user", user)
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, nil)
		return
	}
	txn.Commit()

	writeResponse(w, http.StatusOK, struct{}{})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	req := struct {
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		BirthDate *int    `json:"birth_date,omitempty"`
		Gender    *string `json:"gender,omitempty"`
		Email     *string `json:"email,omitempty"`
		ID        *uint   `json:"id,omitempty"`
	}{}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	if req.FirstName == nil || req.LastName == nil || req.BirthDate == nil ||
		req.Gender == nil || req.Email == nil || req.ID == nil {
		writeResponse(w, http.StatusBadRequest, nil)
		return
	}

	//h.withdb.MxUser.Lock()
	//defer h.withdb.MxUser.Unlock()

	//h.withdb.MaxUser++

	user := models.User{
		FirstName: *req.FirstName,
		LastName:  *req.LastName,
		BirthDate: *req.BirthDate,
		Gender:    *req.Gender,
		Email:     *req.Email,
		ID:        *req.ID,
	}

	txn := h.withdb.DB.Txn(true)
	err = txn.Insert("user", user)
	if err != nil {
		log.Print(err)
		writeResponse(w, http.StatusInternalServerError, nil)
		return
	}
	txn.Commit()

	writeResponse(w, http.StatusOK, struct{}{})
}
