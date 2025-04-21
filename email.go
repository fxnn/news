package main

import "time"

// Email represents the structure of a fetched email.
type Email struct {
	UID     uint32
	Date    time.Time
	Subject string
	From    string
	To      string
	Body    string // Add Body field
	Summary string // Add Summary field
}
