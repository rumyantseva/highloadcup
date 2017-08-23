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

func (h *Handler) User(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)

	fromCache := h.user.Get(id)
	if fromCache != nil {
		writeResponseFromBytes(ctx, http.StatusOK, fromCache)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	txn := h.withdb.DB.Txn(false)
	raw, err := txn.First("user", "id", uid)
	if err != nil {
		writeResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	txn.Abort()

	if raw == nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	var user models.User
	user = raw.(models.User)
	writeResponse(ctx, http.StatusOK, user)
}

func (h *Handler) UpdateUser(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)

	if id == "new" {
		h.CreateUser(ctx)
		return
	}

	uid, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	// Check if user exists
	user, err := db.User(h.withdb.DB, uint(uid))
	if err != nil {
		writeResponse(ctx, http.StatusNotFound, nil)
		return
	}

	req := struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		BirthDate *int    `json:"birth_date"`
		Gender    *string `json:"gender"`
		Email     *string `json:"email"`
	}{}

	//	defer r.Body.Close()
	//	body, err := ioutil.ReadAll(r.Body)
	bodyString := string(ctx.PostBody())
	//log.Print(bodyString)

	// if body contains null, ignore it
	if strings.Contains(bodyString, ": null") {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	err = json.Unmarshal(ctx.PostBody(), &req)
	//err = json.NewDecoder(r.Body).Decode(&req)
	//log.Printf("%+v", req)

	if err != nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	/*if req.FirstName == nil || req.LastName == nil || req.BirthDate == nil ||
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
	}*/

	go func() {
		if req.FirstName != nil {
			user.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			user.LastName = *req.LastName
		}
		if req.BirthDate != nil {
			user.BirthDate = *req.BirthDate
		}
		if req.Gender != nil {
			user.Gender = *req.Gender
		}
		if req.Email != nil {
			user.Email = *req.Email
		}

		go h.user.SetFrom(fmt.Sprint(user.ID), user)

		txn := h.withdb.DB.Txn(true)
		err = txn.Insert("user", *user)
		if err != nil {
			log.Print(err)
			writeResponse(ctx, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponseFromBytes(ctx, http.StatusOK, []byte("{}"))
}

func (h *Handler) CreateUser(ctx *fasthttp.RequestCtx) {

	req := struct {
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		BirthDate *int    `json:"birth_date,omitempty"`
		Gender    *string `json:"gender,omitempty"`
		Email     *string `json:"email,omitempty"`
		ID        *uint   `json:"id,omitempty"`
	}{}

	//	defer r.Body.Close()
	//	err := json.NewDecoder(r.Body).Decode(&req)
	err := json.Unmarshal(ctx.PostBody(), &req)

	if err != nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	if req.FirstName == nil || req.LastName == nil || req.BirthDate == nil ||
		req.Gender == nil || req.Email == nil || req.ID == nil {
		writeResponse(ctx, http.StatusBadRequest, nil)
		return
	}

	//h.withdb.MxUser.Lock()
	//defer h.withdb.MxUser.Unlock()

	//h.withdb.MaxUser++

	go func() {
		user := models.User{
			FirstName: *req.FirstName,
			LastName:  *req.LastName,
			BirthDate: *req.BirthDate,
			Gender:    *req.Gender,
			Email:     *req.Email,
			ID:        *req.ID,
		}

		go h.user.SetFrom(fmt.Sprint(user.ID), user)

		txn := h.withdb.DB.Txn(true)
		err = txn.Insert("user", user)
		if err != nil {
			log.Print(err)
			writeResponse(ctx, http.StatusInternalServerError, nil)
			return
		}
		txn.Commit()
	}()

	writeResponseFromBytes(ctx, http.StatusOK, []byte("{}"))
}
