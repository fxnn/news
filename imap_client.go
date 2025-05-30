package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	// (Dropped html2text import; we’ll just return raw HTML when no text/plain part is present.)
)

// FetchEmails connects to the IMAP server, selects the folder, and fetches emails within the specified date range,
// optionally limiting the number of messages.
func FetchEmails(server string, port int, username, password, folder string, days int, tls bool, limit int) ([]Email, error) {
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

	if len(uids) == 0 {
		log.Println("No messages found matching criteria.")
		return []Email{}, nil // No messages found
	}
	log.Printf("Found %d messages matching date criteria\n", len(uids))

	// Apply limit
	if limit == 0 {
		log.Println("Limit is 0, no messages will be fetched.")
		uids = []uint32{} // Ensure uids is empty if limit is 0
	} else if limit > 0 && len(uids) > limit {
		// We take the last 'limit' UIDs, assuming higher UIDs are generally newer.
		// This is a common behavior but not strictly guaranteed by IMAP for "latest".
		originalCount := len(uids)
		uids = uids[len(uids)-limit:]
		log.Printf("Applied limit: selected %d newest messages from %d found\n", len(uids), originalCount)
	}
	// If limit < 0 (e.g., -1), uids remains unchanged, meaning no limit.

	if len(uids) == 0 {
		log.Println("No messages to fetch after applying date criteria and limit.")
		return []Email{}, nil
	}

	// Create sequence set for fetching
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	// Define what to fetch: Envelope, UID, Date, and the full body structure (BODY[])
	section := &imap.BodySectionName{} // Empty section name means BODY[]
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchInternalDate, section.FetchItem()}

	// Fetch messages
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	var fetchedEmails []Email
	count := 0

	// As each msg comes back, time the work and bump a counter
	for msg := range messages {
		count++

		bodyContent := "" // Default to empty body

		// Get the BODY[] literal reader
		section := &imap.BodySectionName{}
		r := msg.GetBody(section)
		if r == nil {
			log.Printf("Server didn't return BODY[] for UID %d", msg.Uid)
		} else {
			// Parse the MIME message body
			mr, err := mail.CreateReader(r)
			if err == nil {
				// Extract plain text or convert HTML
				bodyContent = extractPlainText(mr, msg.Uid)
				mr.Close() // Close the reader
			} else {
				log.Printf("UID %d: error creating mail.Reader: %v", msg.Uid, err)
			}
		}

		fetchedEmails = append(fetchedEmails, Email{
			UID:     msg.Uid,
			Date:    msg.InternalDate,
			Subject: msg.Envelope.Subject,
			From:    formatAddresses(msg.Envelope.From),
			To:      formatAddresses(msg.Envelope.To),
			Body:    bodyContent, // Populate the Body field
			// Summary will be populated later
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

// extractPlainText walks through the MIME parts of an email body and extracts
// the plain text content. It prefers "text/plain"; if that’s missing, it
// returns the raw HTML body.
func extractPlainText(mr *mail.Reader, uid uint32) string {
	var plainBody string
	var htmlBody string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading next part for UID %d: %v", uid, err)
			continue
		}

		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			ctype, _, err := h.ContentType()
			if err != nil {
				log.Printf("Error parsing content type for UID %d: %v", uid, err)
				continue
			}
			bodyBytes, err := io.ReadAll(part.Body)
			if err != nil {
				log.Printf("Error reading body for UID %d: %v", uid, err)
				continue
			}
			if ctype == "text/plain" && plainBody == "" {
				plainBody = string(bodyBytes)
			} else if ctype == "text/html" && htmlBody == "" {
				htmlBody = string(bodyBytes)
			}
		case *mail.AttachmentHeader:
			// Skip attachments
			continue
		default:
			// Other parts are ignored
		}
	}

	if plainBody != "" {
		return plainBody
	}
	if htmlBody != "" {
		// No plain text found, return the raw HTML instead of converting
		return htmlBody
	}
	return "" // No suitable body found
}
