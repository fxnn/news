package main

import (
	"flag"
	"fmt"
	"log"
	"strings" // Import strings package
)

func main() {
	// Define command line flags
	server := flag.String("server", "", "IMAP server address (required)")
	port := flag.Int("port", 993, "IMAP server port")
	username := flag.String("username", "", "Email username (required)")
	password := flag.String("password", "", "Email password (required)")
	folder := flag.String("folder", "INBOX", "Email folder to search")
	days := flag.Int("days", 7, "Number of days to look back")
	summarizerType := flag.String("summarizer", "stub", "Summarizer type ('stub' or 'langchain')")
	flag.Parse()

	// Validate required flags
	if *server == "" || *username == "" || *password == "" {
		flag.Usage()
		log.Fatal("server, username, and password are required")
	}

	// --- Initialize Summarizer ---
	var summarizer Summarizer // Declare the summarizer variable
	var err error             // Declare err for NewLangChainSummarizer
	switch *summarizerType {
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
		log.Fatalf("Invalid summarizer type: %s. Choose 'stub' or 'langchain'.", *summarizerType)
	}
	// --- End Initialize Summarizer ---

	// Fetch emails using the IMAP client (with TLS)
	emails, err := FetchEmails(*server, *port, *username, *password, *folder, *days, true)
	if err != nil {
		log.Fatalf("Error fetching emails: %v\n", err)
	}

	// Print messages info
	if len(emails) == 0 {
		fmt.Println("No emails found matching the criteria.")
		return
	}

	fmt.Printf("Fetched %d emails:\n", len(emails))
	for i := range emails { // Use index to modify the slice element directly
		email := &emails[i] // Get a pointer to the email for modification

		// Attempt to summarize the body using the chosen summarizer instance
		summary, err := summarizer.Summarize(email.Body) // Use the summarizer instance
		if err != nil {
			// Log the summarization error but continue processing other emails
			log.Printf("WARN: Failed to summarize email UID %d: %v", email.UID, err)
			email.Summary = "[Summarization failed]" // Assign placeholder on error
		} else {
			email.Summary = summary // Assign the generated summary
		}

		fmt.Printf("\n=== Message ===\n")
		fmt.Printf("UID: %d\n", email.UID)
		fmt.Printf("Date: %v\n", email.Date)
		fmt.Printf("Subject: %s\n", email.Subject)
		fmt.Printf("From: %v\n", email.From)
		fmt.Printf("To: %v\n", email.To)

		// Create and print body preview
		preview := createBodyPreview(email.Body)
		fmt.Printf("Body Preview: %s\n", preview)
		fmt.Printf("Summary: %s\n", email.Summary) // Print the summary
	}
}

// createBodyPreview generates a short, single-line preview of an email body.
// It replaces CRLF, newline, and carriage return characters with spaces,
// then truncates to 20 characters, adding ellipsis if needed.
func createBodyPreview(body string) string {
	// Replace CRLF first, then standalone CR and LF to handle all line endings correctly
	preview := strings.ReplaceAll(body, "\r\n", " ")
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "\r", " ") // Replace standalone CR with space
	if len(preview) > 20 {
		preview = preview[:20] + "..."
	}
	return preview
}
