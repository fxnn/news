package main

import (
	"flag"
	"fmt"
	"log"
	"strings" // Import strings package
)

// config holds all the application configuration values derived from flags.
type config struct {
	server         string
	port           int
	username       string
	password       string
	folder         string
	days           int
	summarizerType string
}

// parseAndValidateFlags parses command line flags and validates required ones.
// It terminates the program if validation fails.
func parseAndValidateFlags() config {
	cfg := config{}
	flag.StringVar(&cfg.server, "server", "", "IMAP server address (required)")
	flag.IntVar(&cfg.port, "port", 993, "IMAP server port")
	flag.StringVar(&cfg.username, "username", "", "Email username (required)")
	flag.StringVar(&cfg.password, "password", "", "Email password (required)")
	flag.StringVar(&cfg.folder, "folder", "INBOX", "Email folder to search")
	flag.IntVar(&cfg.days, "days", 7, "Number of days to look back")
	flag.StringVar(&cfg.summarizerType, "summarizer", "stub", "Summarizer type ('stub' or 'langchain')")
	flag.Parse()

	if cfg.server == "" || cfg.username == "" || cfg.password == "" {
		flag.Usage()
		log.Fatal("server, username, and password are required")
	}
	return cfg
}

// initializeSummarizer creates and returns a summarizer based on the provided type.
// It terminates the program if the type is invalid or initialization fails.
func initializeSummarizer(summarizerType string) Summarizer {
	var summarizer Summarizer
	var err error
	switch summarizerType {
	case "stub":
		summarizer = NewStubSummarizer()
		log.Println("Using stub summarizer.")
	case "langchain":
		summarizer, err = NewLangChainSummarizer()
		if err != nil {
			log.Fatalf("Failed to initialize LangChain summarizer: %v", err)
		}
		log.Println("Using LangChain summarizer.")
	default:
		log.Fatalf("Invalid summarizer type: %s. Choose 'stub' or 'langchain'.", summarizerType)
	}
	return summarizer
}

// processEmails fetches, summarizes, and prints emails.
func processEmails(cfg config, summarizer Summarizer) {
	emails, err := FetchEmails(cfg.server, cfg.port, cfg.username, cfg.password, cfg.folder, cfg.days, true)
	if err != nil {
		log.Fatalf("Error fetching emails: %v\n", err)
	}

	if len(emails) == 0 {
		fmt.Println("No emails found matching the criteria.")
		return
	}

	fmt.Printf("Fetched %d emails:\n", len(emails))
	for i := range emails {
		email := &emails[i]

		email.Stories, email.SummarizationError = summarizer.Summarize(email.Body)
		if email.SummarizationError != nil {
			log.Printf("WARN: Failed to summarize email UID %d: %v", email.UID, email.SummarizationError)
		}

		formattedOutput := formatEmailDetails(email)
		fmt.Print(formattedOutput)
	}
}

func main() {
	cfg := parseAndValidateFlags()
	summarizer := initializeSummarizer(cfg.summarizerType)
	processEmails(cfg, summarizer)
}

// formatEmailDetails creates a string representation of an email, including its basic info,
// body preview, and all summarized stories or any summarization error.
func formatEmailDetails(email *Email) string {
	var sb strings.Builder

	sb.WriteString("\n=== Message ===\n")
	sb.WriteString(fmt.Sprintf("UID: %d\n", email.UID))
	sb.WriteString(fmt.Sprintf("Date: %v\n", email.Date))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", email.Subject))
	sb.WriteString(fmt.Sprintf("From: %v\n", email.From))
	sb.WriteString(fmt.Sprintf("To: %v\n", email.To))

	preview := createBodyPreview(email.Body)
	sb.WriteString(fmt.Sprintf("Body Preview: %s\n", preview))

	if email.SummarizationError != nil {
		sb.WriteString(fmt.Sprintf("Summarization Error: %v\n", email.SummarizationError))
	} else if len(email.Stories) == 0 {
		sb.WriteString("[No summary generated]\n")
	} else {
		for i, story := range email.Stories {
			sb.WriteString(fmt.Sprintf("--- Story %d ---\n", i+1))
			sb.WriteString(fmt.Sprintf("Headline: %s\n", story.Headline))
			sb.WriteString(fmt.Sprintf("Teaser: %s\n", story.Teaser))
			sb.WriteString(fmt.Sprintf("URL: %s\n", story.URL))
		}
	}
	return sb.String()
}

// createBodyPreview generates a short, single-line preview of an email body.
// It replaces CRLF, newline, and carriage return characters with spaces,
// then truncates to 100 characters, adding ellipsis if needed.
func createBodyPreview(body string) string {
	// Replace CRLF first, then standalone CR and LF to handle all line endings correctly
	preview := strings.ReplaceAll(body, "\r\n", " ")
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "\r", " ") // Replace standalone CR with space
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	return preview
}
