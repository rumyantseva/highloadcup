package handlers

import (
	"time"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

type LocationChecker struct {
	fromDate *int
	toDate   *int
	fromAge  *int
	toAge    *int
	gender   string

	current int
}

func NewLocationCheker(fromDate, toDate, fromAge, toAge *int, gender string, current int) *LocationChecker {
	return &LocationChecker{
		fromDate: fromDate,
		toDate:   toDate,
		fromAge:  fromAge,
		toAge:    toAge,
		gender:   gender,
		current:  current,
	}
}

func (lc *LocationChecker) Check(mem *memdb.MemDB, visit *models.Visit) bool {
	if lc.fromDate != nil && *lc.fromDate > visit.VisitedAt {
		return false
	}

	if lc.toDate != nil && *lc.toDate < visit.VisitedAt {
		return false
	}

	if lc.fromAge != nil || lc.toAge != nil || len(lc.gender) > 0 {
		user, err := db.User(mem, visit.User)
		if err != nil {
			return false
		}

		born := time.Unix(int64(user.BirthDate), 0)
		if lc.fromAge != nil && *lc.fromAge > age(born) {
			return false
		}

		if lc.toAge != nil && *lc.toAge <= age(born) {
			return false
		}

		if len(lc.gender) > 0 && lc.gender != user.Gender {
			return false
		}
	}

	return true
}

func age(birthday time.Time) int {
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		years--
	}
	return years
}
