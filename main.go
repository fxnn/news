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
		email.Stories, email.SummarizationError = summarizer.Summarize(email.Body)
		if email.SummarizationError != nil {
			// Log the summarization error but continue processing other emails
			// The error is stored in email.SummarizationError and will be handled by formatEmailDetails
			log.Printf("WARN: Failed to summarize email UID %d: %v", email.UID, email.SummarizationError)
		}

		// Format and print the email details including all stories or errors
		formattedOutput := formatEmailDetails(email)
		fmt.Print(formattedOutput)
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
