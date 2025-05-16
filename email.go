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
	// Summary string // Add Summary field - Replaced by Stories and SummarizationError
	Stories            []Story // Stores all stories from the summarizer
	SummarizationError error   // Stores any error from the summarization process
}
