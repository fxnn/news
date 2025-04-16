package main

import (
	"fmt"
	"io"
	"log"
	"mime" // Import mime package for header parsing
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message" // Import go-message for MIME parsing
	"github.com/emersion/go-message/mail"
	"github.com/jaytaylor/html2text" // Import html2text for HTML conversion
)

// Email represents the essential metadata of an email.
type Email struct {
	UID     uint32
	Date    time.Time
	Subject string
	From    string
	To      string
	Body    string // Add Body field
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
	for msg := range messages {
		bodyContent := "" // Default to empty body

		// Get the BODY[] literal reader
		section := &imap.BodySectionName{}
		r := msg.GetBody(section)
		if r == nil {
			log.Printf("Server didn't return BODY[] for UID %d", msg.Uid)
		} else {
			// Parse the MIME message body
			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Printf("Error creating mail reader for UID %d: %v", msg.Uid, err)
			} else {
				// Extract plain text or convert HTML
				bodyContent = extractPlainText(mr, msg.Uid)
				mr.Close() // Close the reader
			}
		}

		fetchedEmails = append(fetchedEmails, Email{
			UID:     msg.Uid,
			Date:    msg.InternalDate,
			Subject: msg.Envelope.Subject,
			From:    formatAddresses(msg.Envelope.From),
			To:      formatAddresses(msg.Envelope.To),
			Body:    bodyContent, // Populate the Body field
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
// the plain text content. It prefers text/plain, but falls back to converting
// text/html if necessary.
func extractPlainText(mr *mail.Reader, uid uint32) string {
	plainBody := ""
	htmlBody := ""

	// Recursively walk through the MIME parts
	var walkPart func(p *message.Entity)
	walkPart = func(p *message.Entity) {
		mediaType, params, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
		if err != nil {
			// Ignore parts with invalid Content-Type
			return
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			// This is a multipart entity, walk its children
			pr := p.MultipartReader()
			if pr == nil {
				return // Should not happen for multipart/*
			}
			for {
				subPart, err := pr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading sub-part for UID %d: %v", uid, err)
					continue
				}
				walkPart(subPart) // Recurse
			}
		} else if mediaType == "text/plain" && plainBody == "" { // Prefer plain text, take the first one found
			bodyBytes, err := io.ReadAll(p.Body)
			if err != nil {
				log.Printf("Error reading text/plain part for UID %d: %v", uid, err)
			} else {
				// Consider charset if specified, default to UTF-8
				charset := params["charset"]
				if charset != "" && !strings.EqualFold(charset, "utf-8") {
					// TODO: Add charset conversion if needed, for now assume UTF-8 or compatible
					log.Printf("UID %d: Found text/plain with charset %s, attempting UTF-8 read", uid, charset)
				}
				plainBody = string(bodyBytes)
			}
		} else if mediaType == "text/html" && htmlBody == "" { // Store the first HTML part found as fallback
			bodyBytes, err := io.ReadAll(p.Body)
			if err != nil {
				log.Printf("Error reading text/html part for UID %d: %v", uid, err)
			} else {
				htmlBody = string(bodyBytes)
			}
		}
	}

	// Start walking from the main message entity
	walkPart(mr.Entity)

	// Return plain text if found, otherwise convert HTML
	if plainBody != "" {
		return plainBody
	} else if htmlBody != "" {
		convertedText, err := html2text.FromString(htmlBody, html2text.Options{PrettyTables: true})
		if err != nil {
			log.Printf("Error converting HTML to text for UID %d: %v", uid, err)
			return htmlBody // Return raw HTML on conversion error
		}
		return convertedText
	}

	// If neither plain nor HTML found (or errors occurred), return empty
	return ""
}
