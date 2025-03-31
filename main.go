package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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

	// Connect to server
	addr := fmt.Sprintf("%s:%d", *server, *port)
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to %s\n", addr)

	// Cleanup
	defer c.Logout()

	// Login
	if err := c.Login(*username, *password); err != nil {
		log.Fatal(err)
	}
	log.Printf("Logged in as %s\n", *username)

	// Select folder
	_, err = c.Select(*folder, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Selected folder: %s\n", *folder)

	// Calculate the date threshold
	since := time.Now().AddDate(0, 0, -*days)

	// Search criteria
	criteria := imap.NewSearchCriteria()
	criteria.Since = since

	// Search for messages
	uids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d messages\n", len(uids))

	if len(uids) == 0 {
		return
	}

	// Create sequence set for fetching
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	// Define what to fetch
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchInternalDate}

	// Fetch messages
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	// Print messages info
	for msg := range messages {
		fmt.Printf("\n=== Message ===\n")
		fmt.Printf("UID: %d\n", msg.Uid)
		fmt.Printf("Date: %v\n", msg.InternalDate)
		fmt.Printf("Subject: %s\n", msg.Envelope.Subject)
		fmt.Printf("From: %v\n", formatAddresses(msg.Envelope.From))
		fmt.Printf("To: %v\n", formatAddresses(msg.Envelope.To))
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}

func formatAddresses(addresses []*imap.Address) string {
	if len(addresses) == 0 {
		return ""
	}
	addr := addresses[0]
	if addr.PersonalName != "" {
		return fmt.Sprintf("%s <%s@%s>", addr.PersonalName, addr.MailboxName, addr.HostName)
	}
	return fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName)
}
