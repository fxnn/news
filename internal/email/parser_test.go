package email

import (
	"strings"
	"testing"
	"time"
)

func TestParse_PlainText(t *testing.T) {
	rawEmail := `From: John Doe <john@example.com>
To: jane@example.com
Subject: Test Email
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <123456@example.com>

This is a plain text email body.
`

	email, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if email.Subject != "Test Email" {
		t.Errorf("Subject = %v, want Test Email", email.Subject)
	}

	if email.FromEmail != "john@example.com" {
		t.Errorf("FromEmail = %v, want john@example.com", email.FromEmail)
	}

	if email.FromName != "John Doe" {
		t.Errorf("FromName = %v, want John Doe", email.FromName)
	}

	if email.MessageID != "<123456@example.com>" {
		t.Errorf("MessageID = %v, want <123456@example.com>", email.MessageID)
	}

	if !strings.Contains(email.Body, "plain text email body") {
		t.Errorf("Body = %v, should contain 'plain text email body'", email.Body)
	}

	expectedDate := time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", -7*3600))
	if !email.Date.Equal(expectedDate) {
		t.Errorf("Date = %v, want %v", email.Date, expectedDate)
	}
}

func TestParse_HTML(t *testing.T) {
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: HTML Email
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <html123@example.com>
Content-Type: text/html; charset="UTF-8"

<html>
<body>
<p>This is an HTML email.</p>
</body>
</html>
`

	email, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if email.Subject != "HTML Email" {
		t.Errorf("Subject = %v, want HTML Email", email.Subject)
	}

	if email.FromEmail != "sender@example.com" {
		t.Errorf("FromEmail = %v, want sender@example.com", email.FromEmail)
	}

	if email.FromName != "" {
		t.Errorf("FromName = %v, want empty string", email.FromName)
	}

	if !strings.Contains(email.Body, "HTML email") {
		t.Errorf("Body = %v, should contain 'HTML email'", email.Body)
	}
}

func TestParse_Multipart(t *testing.T) {
	rawEmail := `From: Alice <alice@example.com>
To: bob@example.com
Subject: Multipart Email
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <multi123@example.com>
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain; charset="UTF-8"

This is the plain text version.

--boundary123
Content-Type: text/html; charset="UTF-8"

<html><body><p>This is the HTML version.</p></body></html>

--boundary123--
`

	email, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if email.Subject != "Multipart Email" {
		t.Errorf("Subject = %v, want Multipart Email", email.Subject)
	}

	if email.FromEmail != "alice@example.com" {
		t.Errorf("FromEmail = %v, want alice@example.com", email.FromEmail)
	}

	if email.FromName != "Alice" {
		t.Errorf("FromName = %v, want Alice", email.FromName)
	}

	// Should prefer plain text for multipart emails
	if !strings.Contains(email.Body, "plain text version") {
		t.Errorf("Body = %v, should contain 'plain text version'", email.Body)
	}
}

func TestParse_MissingHeaders(t *testing.T) {
	rawEmail := `From: test@example.com
Subject: Minimal Email

Body only.
`

	email, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if email.Subject != "Minimal Email" {
		t.Errorf("Subject = %v, want Minimal Email", email.Subject)
	}

	if email.FromEmail != "test@example.com" {
		t.Errorf("FromEmail = %v, want test@example.com", email.FromEmail)
	}

	// Should have generated fallback Message-ID
	if !strings.HasPrefix(email.MessageID, "<generated-") {
		t.Errorf("MessageID = %v, want generated fallback starting with <generated-", email.MessageID)
	}
	if !strings.HasSuffix(email.MessageID, "@fallback>") {
		t.Errorf("MessageID = %v, want generated fallback ending with @fallback>", email.MessageID)
	}

	if email.Date.IsZero() {
		t.Error("Date should not be zero")
	}
}

func TestParse_InvalidEmail(t *testing.T) {
	rawEmail := `This is not a valid email`

	_, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		// It's okay if parsing fails for invalid input
		// But the function should handle it gracefully
		t.Logf("Parse() returned error for invalid email: %v", err)
	}
}

func TestParse_EncodedSubject(t *testing.T) {
	rawEmail := `From: test@example.com
Subject: =?utf-8?B?R2VsZCBhbmxlZ2VuIGbDvHIgS2luZGU=?= =?utf-8?B?cg==?=
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <encoded@example.com>

Test body.
`

	email, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// The decoded subject should be "Geld anlegen für Kinder"
	expectedSubject := "Geld anlegen für Kinder"
	if email.Subject != expectedSubject {
		t.Errorf("Subject = %v, want %v", email.Subject, expectedSubject)
	}
}

func TestParse_EncodedFromName(t *testing.T) {
	rawEmail := `From: =?utf-8?Q?J=C3=B6rg_M=C3=BCller?= <jorg@example.com>
Subject: Test
Date: Mon, 02 Jan 2006 15:04:05 -0700

Test body.
`

	email, err := Parse(strings.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// The decoded name should be "Jörg Müller"
	expectedName := "Jörg Müller"
	if email.FromName != expectedName {
		t.Errorf("FromName = %v, want %v", email.FromName, expectedName)
	}

	if email.FromEmail != "jorg@example.com" {
		t.Errorf("FromEmail = %v, want jorg@example.com", email.FromEmail)
	}
}
