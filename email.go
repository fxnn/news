package main

import "time"

// Story represents a single news item or section extracted from an email.
type Story struct {
	Headline      string
	Teaser        string
	URL           string
	Source        string    // Source of the news story (e.g., email From field)
	PublishedDate time.Time // Date the story was published (e.g., email Date field)
}

// Email represents the structure of a fetched email.
type Email struct {
	UID     uint32
	Date    time.Time
	Subject string
	From    string
	To      string
	Body    string // Add Body field
	// Summary string // Add Summary field - Replaced by Stories and SummarizationError
	Stories []Story // Stores all stories from the summarizer
}
