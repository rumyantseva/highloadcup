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

		age := age(lc.current, user.BirthDate)
		if lc.fromAge != nil && *lc.fromAge > age {
			return false
		}

		if lc.toAge != nil && *lc.toAge <= age {
			return false
		}

		if len(lc.gender) > 0 && lc.gender != user.Gender {
			return false
		}
	}

	return true
}

func age(current int, birthday int) int {
	now := time.Unix(int64(current), 0)
	born := time.Unix(int64(birthday), 0)
	years := now.Year() - born.Year()
	if now.YearDay() < born.YearDay() {
		years--
	}
	return years
}
