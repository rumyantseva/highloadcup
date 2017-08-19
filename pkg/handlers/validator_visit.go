package handlers

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

type VisitChecker struct {
	fromDate   *int
	toDate     *int
	country    string
	toDistance *int
}

func NewVisitCheker(fromDate, toDate, toDistance *int, country string) *VisitChecker {
	return &VisitChecker{
		fromDate:   fromDate,
		toDate:     toDate,
		country:    country,
		toDistance: toDistance,
	}
}

func (vc *VisitChecker) Check(mem *memdb.MemDB, visit *models.Visit) bool {
	if vc.fromDate != nil && *vc.fromDate > visit.VisitedAt {
		return false
	}

	if vc.toDate != nil && *vc.toDate < visit.VisitedAt {
		return false
	}

	if vc.toDistance != nil || len(vc.country) > 0 {
		loc, err := db.Location(mem, visit.Location)
		if err != nil {
			return false
		}

		if vc.toDistance != nil && *vc.toDistance < loc.Distance {
			return false
		}

		if len(vc.country) > 0 && vc.country != loc.Country {
			return false
		}
	}

	return true
}
