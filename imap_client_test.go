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
	Date      time.Time
	Subject   string
	From      string // e.g., "Sender <sender@example.com>"
	To        string // e.g., "Recipient <recipient@example.com>"
	BodyPlain string
	BodyHTML  string
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
	for i, msgData := range messages {
		// Construct a multipart/alternative message
		var fullMsg strings.Builder
		boundary := "testboundary123"

		// Headers
		fmt.Fprintf(&fullMsg, "From: %s\r\n", msgData.From)
		fmt.Fprintf(&fullMsg, "To: %s\r\n", msgData.To)
		fmt.Fprintf(&fullMsg, "Subject: %s\r\n", msgData.Subject)
		fmt.Fprintf(&fullMsg, "Date: %s\r\n", msgData.Date.Format(time.RFC1123Z))
		fmt.Fprintf(&fullMsg, "MIME-Version: 1.0\r\n")
		fmt.Fprintf(&fullMsg, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
		fmt.Fprintf(&fullMsg, "\r\n") // End of headers

		// Plain text part
		fmt.Fprintf(&fullMsg, "--%s\r\n", boundary)
		fmt.Fprintf(&fullMsg, "Content-Type: text/plain; charset=utf-8\r\n")
		fmt.Fprintf(&fullMsg, "Content-Transfer-Encoding: quoted-printable\r\n") // Use quoted-printable for safety
		fmt.Fprintf(&fullMsg, "\r\n")
		// Simple quoted-printable encoding (replace '=' with '=3D', can be improved)
		encodedPlain := strings.ReplaceAll(msgData.BodyPlain, "=", "=3D")
		fmt.Fprintf(&fullMsg, "%s\r\n", encodedPlain)

		// HTML part
		fmt.Fprintf(&fullMsg, "--%s\r\n", boundary)
		fmt.Fprintf(&fullMsg, "Content-Type: text/html; charset=utf-8\r\n")
		fmt.Fprintf(&fullMsg, "Content-Transfer-Encoding: quoted-printable\r\n")
		fmt.Fprintf(&fullMsg, "\r\n")
		encodedHTML := strings.ReplaceAll(msgData.BodyHTML, "=", "=3D")
		fmt.Fprintf(&fullMsg, "%s\r\n", encodedHTML)

		// End boundary
		fmt.Fprintf(&fullMsg, "--%s--\r\n", boundary)

		// Create the message in the mailbox
		err = mbox.CreateMessage([]string{"\\Seen"}, msgData.Date, strings.NewReader(fullMsg.String()))
		if err != nil {
			t.Fatalf("Failed to create message #%d (%s): %v", i+1, msgData.Subject, err)
		}
		t.Logf("Added message #%d: %s", i+1, msgData.Subject)
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
	recentDate := time.Now().AddDate(0, 0, -3) // 3 days ago
	oldDate := time.Now().AddDate(0, 0, -10)   // 10 days ago
	testMessages := []TestMessage{
		{
			Date:      recentDate,
			Subject:   "Recent Test Email",
			From:      "Sender <sender@example.com>",
			To:        "Recipient <recipient@example.com>",
			BodyPlain: "This is the plain text version.",
			BodyHTML:  "<html><body><p>This is the <b>HTML</b> version.</p></body></html>",
		},
		{
			Date:      oldDate,
			Subject:   "Old Test Email",
			From:      "Another Sender <sender2@example.com>",
			To:        "Another Recipient <recipient2@example.com>",
			BodyPlain: "This is an old plain text email.",
			BodyHTML:  "<html><body><p>Old HTML content.</p></body></html>",
		},
	}

	// Setup mock server with our test messages
	host, port, username, password, cleanup := setupMockIMAPServer(t, testMessages)
	defer cleanup()

	// --- Original Test Case: Fetch recent emails, no limit ---
	t.Run("FetchRecentEmailsNoLimit", func(t *testing.T) {
		emails, err := FetchEmails(host, port, username, password, "INBOX", 7, false, -1)
		if err != nil {
			t.Fatalf("FetchEmails failed: %v", err)
		}
		if len(emails) != 1 {
			t.Errorf("Expected 1 recent email, got %d", len(emails))
		}
		if emails[0].Subject != "Recent Test Email" {
			t.Errorf("Expected subject 'Recent Test Email', got '%s'", emails[0].Subject)
		}
		// Check body (should be the plain text part)
		expectedBody := "This is the plain text version."
		if strings.TrimSpace(emails[0].Body) != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, strings.TrimSpace(emails[0].Body))
		}
	})

	// --- Test Case: Apply limit ---
	// Add one more recent message to test limiting
	moreTestMessages := append(testMessages, TestMessage{
		Date:      time.Now().AddDate(0, 0, -1), // 1 day ago
		Subject:   "Very Recent Test Email",
		From:      "Sender3 <sender3@example.com>",
		To:        "Recipient3 <recipient3@example.com>",
		BodyPlain: "This is very recent.",
		BodyHTML:  "<html><body><p>Very recent HTML.</p></body></html>",
	})
	// Need to restart server with new messages for this sub-test
	hostLimited, portLimited, userLimited, passLimited, cleanupLimited := setupMockIMAPServer(t, moreTestMessages)
	defer cleanupLimited()

	t.Run("FetchWithLimitApplied", func(t *testing.T) {
		// We have 2 messages within 7 days: "Recent Test Email" (3 days ago) and "Very Recent Test Email" (1 day ago)
		// The mock server adds messages and they get UIDs. FetchEmails with limit takes the *last* UIDs.
		// Assuming "Very Recent Test Email" was added last, it should have a higher UID.
		emails, err := FetchEmails(hostLimited, portLimited, userLimited, passLimited, "INBOX", 7, false, 1)
		if err != nil {
			t.Fatalf("FetchEmails with limit failed: %v", err)
		}
		if len(emails) != 1 {
			t.Fatalf("Expected 1 email due to limit, got %d", len(emails))
		}
		// The one email should be "Very Recent Test Email" as it's assumed to be the "newest" by UID by FetchEmails logic
		if emails[0].Subject != "Very Recent Test Email" {
			t.Errorf("Expected subject 'Very Recent Test Email' with limit 1, got '%s'", emails[0].Subject)
		}
	})

	t.Run("FetchWithLimitGreaterThanAvailable", func(t *testing.T) {
		// Still using moreTestMessages (2 recent emails)
		emails, err := FetchEmails(hostLimited, portLimited, userLimited, passLimited, "INBOX", 7, false, 5)
		if err != nil {
			t.Fatalf("FetchEmails with limit > available failed: %v", err)
		}
		if len(emails) != 2 { // Should get both recent emails
			t.Errorf("Expected 2 emails when limit is greater than available, got %d", len(emails))
		}
	})

	t.Run("FetchWithLimitZero", func(t *testing.T) {
		emails, err := FetchEmails(hostLimited, portLimited, userLimited, passLimited, "INBOX", 7, false, 0)
		if err != nil {
			t.Fatalf("FetchEmails with limit 0 failed: %v", err)
		}
		if len(emails) != 0 {
			t.Errorf("Expected 0 emails when limit is 0, got %d", len(emails))
		}
	})

}
