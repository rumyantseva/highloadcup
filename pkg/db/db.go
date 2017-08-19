package db

import (
	"sync"

	memdb "github.com/hashicorp/go-memdb"
)

type WithMax struct {
	DB *memdb.MemDB

	MaxLocation uint
	MxLocation  *sync.RWMutex

	MaxUser uint
	MxUser  *sync.RWMutex

	MaxVisit uint
	MxVisit  *sync.RWMutex
}

func NewWithMax(memdb *memdb.MemDB) *WithMax {
	return &WithMax{
		DB: memdb,

		MxLocation: &sync.RWMutex{},
		MxUser:     &sync.RWMutex{},
		MxVisit:    &sync.RWMutex{},
	}
}
