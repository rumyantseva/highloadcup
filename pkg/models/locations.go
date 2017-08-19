package models

type Location struct {
	Distance int    `json:"distance"`
	City     string `json:"city"`
	Place    string `json:"place"`
	ID       uint   `json:"id"`
	Country  string `json:"country"`
}
