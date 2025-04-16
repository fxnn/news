package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/server"
)

// TestMessage holds the data needed to create a test email message.
type TestMessage struct {
	Date    time.Time
	Subject string
	From    string // e.g., "Sender <sender@example.com>"
	To      string // e.g., "Recipient <recipient@example.com>"
	Body    string
}

// setupMockIMAPServer creates and starts a mock IMAP server with the given test messages.
// It returns the server host, port, username, password, and a cleanup function.
func setupMockIMAPServer(t *testing.T, messages []TestMessage) (host string, port int, username, password string, cleanup func()) {
	// Create a memory backend
	be := memory.New()

	// Set credentials
	username = "username"
	password = "password"

	// Create a user
	user := &memory.User{}

	// Login to get the user
	connInfo := &imap.ConnInfo{}
	u, err := be.Login(connInfo, username, password)
	if err != nil {
		// First login creates the user
		t.Logf("First login creates the user: %v", err)
	} else {
		user = u.(*memory.User)
	}

	// Create INBOX
	err = user.CreateMailbox("INBOX")
	if err != nil {
		t.Logf("INBOX already exists: %v", err)
	}

	// Get the INBOX
	mboxInterface, err := user.GetMailbox("INBOX")
	if err != nil {
		t.Fatalf("Failed to get INBOX: %v", err)
	}
	mbox := mboxInterface.(*memory.Mailbox)

	// Clear any existing messages in the mailbox
	mbox.Messages = []*memory.Message{}

	// Add all test messages to the mailbox
	for i, msg := range messages {
		// Format the email with headers and body
		var fullMsg strings.Builder
		fmt.Fprintf(&fullMsg, "From: %s\r\n", msg.From)
		fmt.Fprintf(&fullMsg, "To: %s\r\n", msg.To)
		fmt.Fprintf(&fullMsg, "Subject: %s\r\n", msg.Subject)
		fmt.Fprintf(&fullMsg, "Date: %s\r\n", msg.Date.Format(time.RFC1123Z))
		fmt.Fprintf(&fullMsg, "\r\n%s", msg.Body)
		
		// Create the message in the mailbox
		err = mbox.CreateMessage([]string{"\\Seen"}, msg.Date, strings.NewReader(fullMsg.String()))
		if err != nil {
			t.Fatalf("Failed to create message #%d (%s): %v", i+1, msg.Subject, err)
		}
		t.Logf("Added message #%d: %s", i+1, msg.Subject)
	}

	// Create a new server
	s := server.New(be)
	s.AllowInsecureAuth = true

	// Listen on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// Start the server
	go s.Serve(listener)

	// Get the chosen port
	listenerAddr := listener.Addr().String()
	host, portStr, _ := net.SplitHostPort(listenerAddr)
	portNum := 0
	fmt.Sscanf(portStr, "%d", &portNum)
	port = portNum

	// Return a cleanup function
	cleanup = func() {
		listener.Close()
		s.Close()
	}

	return host, port, username, password, cleanup
}

func TestFetchEmails(t *testing.T) {
	// Define test messages
	recentDate := time.Now().AddDate(0, 0, -3)
	oldDate := time.Now().AddDate(0, 0, -10)
	testMessages := []TestMessage{
		{
			Date:    recentDate,
			Subject: "Recent Test Email",
			From:    "Sender <sender@example.com>",
			To:      "Recipient <recipient@example.com>",
			Body:    "This is a recent test email.",
		},
		{
			Date:    oldDate,
			Subject: "Old Test Email",
			From:    "Sender <sender@example.com>",
			To:      "Recipient <recipient@example.com>",
			Body:    "This is an old test email.",
		},
	}

	// Setup mock server with our test messages
	host, port, username, password, cleanup := setupMockIMAPServer(t, testMessages)
	defer cleanup()

	// Use the FetchEmails function with tls=false for testing
	emails, err := FetchEmails(host, port, username, password, "INBOX", 7, false)
	if err != nil {
		t.Fatalf("FetchEmails failed: %v", err)
	}

	// Verify results
	// We should have exactly 1 email (the recent one, not the old one)
	if len(emails) != 1 {
		t.Errorf("Expected exactly 1 email (the recent one), got %d", len(emails))
	}

	// Check the date filtering
	dateThreshold := time.Now().AddDate(0, 0, -7)
	for i, email := range emails {
		if email.Date.Before(dateThreshold) {
			t.Errorf("Email %d has date %v, which is before the threshold %v",
				i, email.Date, dateThreshold)
		}
	}

	// Check the content of the email
	// Check subject
	expectedSubject := "Recent Test Email"
	if emails[0].Subject != expectedSubject {
		t.Errorf("Expected subject '%s', got '%s'", expectedSubject, emails[0].Subject)
	}

	// Check sender
	if !strings.Contains(emails[0].From, "sender@example.com") {
		t.Errorf("Expected From to contain 'sender@example.com', got '%s'", emails[0].From)
	}

	// Check body
	expectedBody := "This is a recent test email."
	// Note: IMAP servers might add extra CRLF, so we trim space for comparison
	if strings.TrimSpace(emails[0].Body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, strings.TrimSpace(emails[0].Body))
	}
}
