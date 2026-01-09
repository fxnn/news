package story

import "time"

type Story struct {
	Headline  string    `json:"headline"`
	Teaser    string    `json:"teaser"`
	URL       string    `json:"url"`
	FromEmail string    `json:"from_email"`
	FromName  string    `json:"from_name"`
	Date      time.Time `json:"date"`
}
