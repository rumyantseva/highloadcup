package handlers

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func (h *Handler) User(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(w, http.StatusNotFound, nil)
		return
	}

	txn := h.db.Txn(false)
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
	/*	id := ps.ByName("id")
		uid, err := strconv.ParseUint(id, 10, 32)

		if err != nil {
			writeResponse(w, http.StatusBadRequest, nil)
			return
		}

		// Check if user exists

		type request struct {
			FirstName *string `json:"first_name,omitempty"`
			LastName  *string `json:"last_name,omitempty"`
			BirthDate *int    `json:"birth_date,omitempty"`
			Gender    *string `json:"gender,omitempty"`
			Email     *string `json:"email,omitempty"`
		}

		txn := h.db.Txn(true)
		for _, user := range data.Users {
			if err := txn.("user", user); err != nil {
				return err
			}
		}
		txn.Commit()*/

}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	/*	txn := db.Txn(true)
		for _, user := range data.Users {
			if err := txn.Insert("user", user); err != nil {
				return err
			}
		}
		txn.Commit()*/
}
