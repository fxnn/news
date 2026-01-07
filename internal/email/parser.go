package email

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Email struct {
	Subject   string
	Body      string
	FromEmail string
	FromName  string
	Date      time.Time
	MessageID string
}

// Parse reads and parses an email from the given reader.
func Parse(r io.Reader) (*Email, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read email: %w", err)
	}

	email := &Email{}

	// Parse Subject
	email.Subject = msg.Header.Get("Subject")

	// Parse From
	from, err := mail.ParseAddress(msg.Header.Get("From"))
	if err == nil {
		email.FromEmail = from.Address
		email.FromName = from.Name
	} else {
		// Fallback if parsing fails
		email.FromEmail = msg.Header.Get("From")
	}

	// Parse Date
	dateStr := msg.Header.Get("Date")
	if dateStr != "" {
		email.Date, err = mail.ParseDate(dateStr)
		if err != nil {
			// Use current time as fallback
			email.Date = time.Now()
		}
	} else {
		email.Date = time.Now()
	}

	// Parse Message-ID
	email.MessageID = msg.Header.Get("Message-ID")

	// Parse Body
	contentType := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Default to plain text if Content-Type can't be parsed
		body, _ := io.ReadAll(msg.Body)
		email.Body = string(body)
		return email, nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		body, err := parseMultipart(msg.Body, params["boundary"])
		if err != nil {
			return nil, fmt.Errorf("failed to parse multipart: %w", err)
		}
		email.Body = body
	} else if mediaType == "text/html" {
		body, _ := io.ReadAll(msg.Body)
		email.Body = extractTextFromHTML(string(body))
	} else {
		body, _ := io.ReadAll(msg.Body)
		email.Body = string(body)
	}

	return email, nil
}

func parseMultipart(body io.Reader, boundary string) (string, error) {
	mr := multipart.NewReader(body, boundary)

	var plainText string
	var htmlText string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		contentType := part.Header.Get("Content-Type")
		mediaType, _, _ := mime.ParseMediaType(contentType)

		partBody, err := io.ReadAll(part)
		if err != nil {
			continue
		}

		switch mediaType {
		case "text/plain":
			plainText = string(partBody)
		case "text/html":
			htmlText = string(partBody)
		}
	}

	// Prefer plain text over HTML
	if plainText != "" {
		return plainText, nil
	}
	if htmlText != "" {
		return extractTextFromHTML(htmlText), nil
	}

	return "", nil
}

func extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var text strings.Builder
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(doc)

	return text.String()
}
