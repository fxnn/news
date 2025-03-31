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

func TestFetchEmails(t *testing.T) {
	// Create a memory backend
	be := memory.New()

	// Create a user
	user := &memory.User{}
	
	// Login to get the user
	connInfo := &imap.ConnInfo{}
	u, err := be.Login(connInfo, "username", "password")
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

	// Create a message within the date range (3 days ago)
	recentDate := time.Now().AddDate(0, 0, -3)
	recentBody := strings.NewReader("From: Sender <sender@example.com>\r\n" +
		"To: Recipient <recipient@example.com>\r\n" +
		"Subject: Recent Test Email\r\n" +
		"Date: " + recentDate.Format(time.RFC1123Z) + "\r\n" +
		"\r\n" +
		"This is a recent test email.")
	
	err = mbox.CreateMessage([]string{"\\Seen"}, recentDate, recentBody)
	if err != nil {
		t.Fatalf("Failed to create recent message: %v", err)
	}

	// Create a message outside the date range (10 days ago)
	oldDate := time.Now().AddDate(0, 0, -10)
	oldBody := strings.NewReader("From: Sender <sender@example.com>\r\n" +
		"To: Recipient <recipient@example.com>\r\n" +
		"Subject: Old Test Email\r\n" +
		"Date: " + oldDate.Format(time.RFC1123Z) + "\r\n" +
		"\r\n" +
		"This is an old test email.")
	
	err = mbox.CreateMessage([]string{"\\Seen"}, oldDate, oldBody)
	if err != nil {
		t.Fatalf("Failed to create old message: %v", err)
	}

	// Create a new server
	s := server.New(be)
	s.AllowInsecureAuth = true

	// Listen on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	// Start the server
	go s.Serve(listener)
	defer s.Close()

	// Get the chosen port
	listenerAddr := listener.Addr().String()
	host, port, _ := net.SplitHostPort(listenerAddr)
	portNum := 0
	fmt.Sscanf(port, "%d", &portNum)

	// Use the FetchEmails function with tls=false for testing
	emails, err := FetchEmails(host, portNum, "username", "password", "INBOX", 7, false)
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
}
