package db

import (
	"fmt"
	"log"
	"net/http"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

func User(db *memdb.MemDB, id uint) (*models.User, error) {
	txn := db.Txn(false)
	raw, err := txn.First("user", "id", id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("%d", http.StatusInternalServerError)
	}
	txn.Abort()

	if raw == nil {
		return nil, fmt.Errorf("%d", http.StatusNotFound)
	}

	var user models.User
	user = raw.(models.User)

	return &user, nil
}
