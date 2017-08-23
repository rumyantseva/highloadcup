package handlers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rumyantseva/highloadcup/pkg/cache"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	withdb   *db.WithMax
	current  int
	user     *cache.Storage
	location *cache.Storage
	visit    *cache.Storage
}

func NewHandler(withmax *db.WithMax, user, location, visit *cache.Storage, current int) *Handler {
	return &Handler{
		withdb:   withmax,
		user:     user,
		location: location,
		visit:    visit,
		current:  current,
	}
}

func writeResponseFromBytes(ctx *fasthttp.RequestCtx, code int, data []byte) {
	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
	len := binary.Size(data)
	ctx.Response.Header.Set("Content-Length", fmt.Sprintf("%d", len))
	ctx.SetStatusCode(code)
	ctx.SetBody(data)
}

func writeResponse(ctx *fasthttp.RequestCtx, code int, resp interface{}) {
	if resp == nil {
		ctx.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
		ctx.SetStatusCode(code)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Couldn't encode response %+v to HTTP response body.", resp)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
	//	len, _ := w.Write(data)

	len := binary.Size(data)
	ctx.Response.Header.Set("Content-Length", fmt.Sprintf("%d", len))
	ctx.SetStatusCode(code)
	ctx.SetBody(data)

	/*	enc := json.NewEncoder(w)
		err = enc.Encode(resp)
		if err != nil {
			log.Printf("Couldn't encode response %+v to HTTP response body.", resp)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}*/
}
