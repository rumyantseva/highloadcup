package models

type Visit struct {
	User      uint `json:"user"`
	Location  uint `json:"location"`
	VisitedAt int  `json:"visited_at"`
	ID        uint `json:"id"`
	Mark      int  `json:"mark"`
}

type VisitExt struct {
	Mark      int    `json:"mark"`
	VisitedAt int    `json:"visited_at"`
	Place     string `json:"place"`
}
