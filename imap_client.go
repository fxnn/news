package main

import (
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// Email represents the essential metadata of an email.
type Email struct {
	UID     uint32
	Date    time.Time
	Subject string
	From    string
	To      string
}

// FetchEmails connects to the IMAP server, selects the folder, and fetches emails within the specified date range.
func FetchEmails(server string, port int, username, password, folder string, days int, tls bool) ([]Email, error) {
	// Connect to server
	addr := fmt.Sprintf("%s:%d", server, port)
	
	var c *client.Client
	var err error
	
	// Use TLS or non-TLS connection based on the tls parameter
	if tls {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	log.Printf("Connected to %s\n", addr)

	// Cleanup
	defer c.Logout()

	// Login
	if err := c.Login(username, password); err != nil {
		return nil, fmt.Errorf("failed to login as %s: %w", username, err)
	}
	log.Printf("Logged in as %s\n", username)

	// Select folder
	_, err = c.Select(folder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select folder %s: %w", folder, err)
	}
	log.Printf("Selected folder: %s\n", folder)

	// Calculate the date threshold
	since := time.Now().AddDate(0, 0, -days)

	// Search criteria
	criteria := imap.NewSearchCriteria()
	criteria.Since = since

	// Search for messages
	uids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search for messages: %w", err)
	}
	log.Printf("Found %d messages\n", len(uids))

	if len(uids) == 0 {
		return []Email{}, nil // No messages found
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

	var fetchedEmails []Email
	for msg := range messages {
		fetchedEmails = append(fetchedEmails, Email{
			UID:     msg.Uid,
			Date:    msg.InternalDate,
			Subject: msg.Envelope.Subject,
			From:    formatAddresses(msg.Envelope.From),
			To:      formatAddresses(msg.Envelope.To),
		})
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return fetchedEmails, nil
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
