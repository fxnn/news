package main

import (
	"encoding/json" // Added for JSON marshalling
	"flag"
	"fmt"
	"log"
	"net/http" // Import net/http package
	"strings"  // Import strings package
)

// EmailFetcher defines the signature for a function that fetches emails.
// This allows for mocking in tests.
type EmailFetcher func(server string, port int, username, password, folder string, days int, tls bool) ([]Email, error)

// config holds all the application configuration values derived from flags.
type config struct {
	server         string
	port           int
	username       string
	password       string
	folder         string
	days           int
	summarizerType string
	mode           string // Application mode: "cli" or "server"
	httpPort       int    // Port for HTTP server mode
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
	flag.StringVar(&cfg.mode, "mode", "cli", "Application mode ('cli' or 'server')")
	flag.IntVar(&cfg.httpPort, "http-port", 8080, "Port for HTTP server (if mode is 'server')")
	flag.Parse()

	if cfg.mode != "cli" && cfg.mode != "server" {
		flag.Usage()
		log.Fatal("Invalid mode. Choose 'cli' or 'server'.")
	}

	if cfg.mode == "cli" && (cfg.server == "" || cfg.username == "" || cfg.password == "") {
		flag.Usage()
		log.Fatal("server, username, and password are required for cli mode")
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

// fetchAndSummarizeEmails fetches emails using the provided fetcher and then summarizes them.
// Individual summarization errors are stored within the Email structs.
// The top-level error is for critical issues like the fetcher failing.
func fetchAndSummarizeEmails(fetcher EmailFetcher, cfg config, summarizer Summarizer) ([]Email, error) {
	// Assuming TLS true based on previous hardcoding in processEmails.
	// This could be made configurable if needed.
	useTLS := true

	emails, err := fetcher(cfg.server, cfg.port, cfg.username, cfg.password, cfg.folder, cfg.days, useTLS)
	if err != nil {
		return nil, fmt.Errorf("error fetching emails: %w", err)
	}

	if len(emails) == 0 {
		return []Email{}, nil
	}

	for i := range emails {
		email := &emails[i] // Use pointer to modify the slice element
		if email.Body != "" {
			email.Stories, email.SummarizationError = summarizer.Summarize(email.Body)
		} else {
			email.Stories = []Story{} // Ensure Stories is not nil
			email.SummarizationError = nil
		}
	}
	return emails, nil
}

// processEmails fetches, summarizes, and prints emails for CLI mode.
func processEmails(cfg config, summarizer Summarizer) {
	emails, err := fetchAndSummarizeEmails(FetchEmails, cfg, summarizer) // Use the real FetchEmails
	if err != nil {
		// fetchAndSummarizeEmails already wraps the error from FetchEmails
		log.Fatalf("Error processing emails: %v\n", err)
	}

	if len(emails) == 0 {
		fmt.Println("No emails found matching the criteria.")
		return
	}

	fmt.Printf("Fetched %d emails:\n", len(emails))
	for i := range emails {
		email := &emails[i]

		// Log summarization errors if any occurred (already populated by fetchAndSummarizeEmails)
		if email.SummarizationError != nil {
			log.Printf("WARN: Failed to summarize email UID %d: %v", email.UID, email.SummarizationError)
		}

		formattedOutput := formatEmailDetails(email)
		fmt.Print(formattedOutput)
	}
}

func main() {
	cfg := parseAndValidateFlags()
	summarizer := initializeSummarizer(cfg.summarizerType) // Initialize summarizer once

	if cfg.mode == "cli" {
		processEmails(cfg, summarizer)
	} else if cfg.mode == "server" {
		startHttpServer(cfg, summarizer) // Pass summarizer to startHttpServer
	}
}

// newStoriesHandler creates an HTTP handler for the /stories endpoint.
// It uses the provided config, summarizer, and email fetcher.
func newStoriesHandler(cfg config, summarizer Summarizer, fetcher EmailFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emails, err := fetchAndSummarizeEmails(fetcher, cfg, summarizer)
		if err != nil {
			// Log the error server-side as well for more details if needed
			log.Printf("ERROR: Failed to fetch or summarize emails for /stories: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch or summarize emails: %v", err), http.StatusInternalServerError)
			return
		}

		allStories := []Story{}
		for _, email := range emails {
			// Only include stories from emails that were successfully summarized
			// and actually produced stories.
			if email.SummarizationError == nil && len(email.Stories) > 0 {
				allStories = append(allStories, email.Stories...)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		// Ensure an empty JSON array "[]" is sent if allStories is empty, not "null".
		if len(allStories) == 0 {
			w.Write([]byte("[]\n")) // Add newline for consistency with Encode
			return
		}

		if err := json.NewEncoder(w).Encode(allStories); err != nil {
			log.Printf("ERROR: Failed to marshal stories to JSON: %v", err)
			http.Error(w, fmt.Sprintf("Failed to marshal stories to JSON: %v", err), http.StatusInternalServerError)
		}
	}
}

// startHttpServer starts the HTTP server with configured routes.
// It now requires a Summarizer to pass to handlers.
func startHttpServer(cfg config, summarizer Summarizer) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "HTTP server is running")
	})

	// Setup /stories handler
	storiesHandler := newStoriesHandler(cfg, summarizer, FetchEmails) // Use the real FetchEmails
	http.HandleFunc("/stories", storiesHandler)

	addr := fmt.Sprintf(":%d", cfg.httpPort)
	log.Printf("Starting HTTP server on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
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
