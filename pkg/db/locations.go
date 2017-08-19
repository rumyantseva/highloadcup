package db

import (
	"fmt"
	"log"
	"net/http"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func Location(db *memdb.MemDB, id uint) (*models.Location, error) {
	txn := db.Txn(false)
	raw, err := txn.First("location", "id", id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("%d", http.StatusInternalServerError)
	}
	txn.Abort()

	if raw == nil {
		return nil, fmt.Errorf("%d", http.StatusNotFound)
	}

	var loc models.Location
	loc = raw.(models.Location)

	return &loc, nil
}
