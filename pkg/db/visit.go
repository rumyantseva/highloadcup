package db

import (
	"fmt"
	"log"
	"net/http"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func Visit(db *memdb.MemDB, id uint) (*models.Visit, error) {
	txn := db.Txn(false)
	raw, err := txn.First("visit", "id", id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("%d", http.StatusInternalServerError)
	}
	txn.Abort()

	if raw == nil {
		return nil, fmt.Errorf("%d", http.StatusNotFound)
	}

	var visit models.Visit
	visit = raw.(models.Visit)

	return &visit, nil
}
