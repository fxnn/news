package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	// Define command line flags
	server := flag.String("server", "", "IMAP server address (required)")
	port := flag.Int("port", 993, "IMAP server port")
	username := flag.String("username", "", "Email username (required)")
	password := flag.String("password", "", "Email password (required)")
	folder := flag.String("folder", "INBOX", "Email folder to search")
	days := flag.Int("days", 7, "Number of days to look back")
	flag.Parse()

	// Validate required flags
	if *server == "" || *username == "" || *password == "" {
		flag.Usage()
		log.Fatal("server, username, and password are required")
	}

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
	for _, email := range emails {
		fmt.Printf("\n=== Message ===\n")
		fmt.Printf("UID: %d\n", email.UID)
		fmt.Printf("Date: %v\n", email.Date)
		fmt.Printf("Subject: %s\n", email.Subject)
		fmt.Printf("From: %v\n", email.From)
		fmt.Printf("To: %v\n", email.To)
	}
}
