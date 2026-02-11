package email

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Email represents a parsed email message with extracted metadata.
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

	// Parse Subject (decode MIME-encoded words)
	decoder := &mime.WordDecoder{}
	subject := msg.Header.Get("Subject")
	decodedSubject, err := decoder.DecodeHeader(subject)
	if err == nil {
		email.Subject = decodedSubject
	} else {
		// Fallback to raw subject if decoding fails
		email.Subject = subject
	}

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
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			log.Warn("failed to parse email date, using current time",
				"date_header", dateStr,
				"error", err)
		}
	} else {
		email.Date = time.Now()
		log := slog.New(slog.NewTextHandler(os.Stderr, nil))
		log.Warn("email missing Date header, using current time")
	}

	// Parse Message-ID (generate fallback if missing to avoid filename conflicts)
	email.MessageID = msg.Header.Get("Message-ID")
	if email.MessageID == "" {
		// Generate fallback Message-ID using hash of email content
		// This ensures unique IDs even for emails without Message-ID headers
		hash := sha256.Sum256([]byte(email.Subject + email.FromEmail + email.Date.String()))
		email.MessageID = "<generated-" + hex.EncodeToString(hash[:16]) + "@fallback>"
		log := slog.New(slog.NewTextHandler(os.Stderr, nil))
		log.Warn("email missing Message-ID header, generated fallback",
			"generated_id", email.MessageID)
	}

	// Parse Body
	contentType := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Default to plain text if Content-Type can't be parsed
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		email.Body = string(body)
		return email, nil
	}

	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		body, err := parseMultipart(msg.Body, params["boundary"])
		if err != nil {
			return nil, fmt.Errorf("failed to parse multipart: %w", err)
		}
		email.Body = body
	case mediaType == "text/html":
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read HTML body: %w", err)
		}
		email.Body = extractTextFromHTML(string(body))
	default:
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
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
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}

		func() {
			defer part.Close()

			contentType := part.Header.Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				// Skip parts with invalid content type
				return
			}

			partBody, err := io.ReadAll(part)
			if err != nil {
				return
			}

			switch mediaType {
			case "text/plain":
				plainText = string(partBody)
			case "text/html":
				htmlText = string(partBody)
			}
		}()
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
			// Trim and skip empty text nodes
			trimmed := strings.TrimSpace(n.Data)
			if trimmed != "" {
				// Add space before text if builder has content and doesn't end with whitespace
				if text.Len() > 0 {
					lastChar := text.String()[text.Len()-1]
					if lastChar != '\n' && lastChar != ' ' {
						text.WriteString(" ")
					}
				}
				text.WriteString(trimmed)
			}
		}
		if n.Type == html.ElementNode {
			// Add newline after block-level elements for better structure
			isBlockElement := n.Data == "p" || n.Data == "div" || n.Data == "br" ||
				n.Data == "h1" || n.Data == "h2" || n.Data == "h3" ||
				n.Data == "h4" || n.Data == "h5" || n.Data == "h6" ||
				n.Data == "li" || n.Data == "tr"

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extract(c)
			}

			if isBlockElement && text.Len() > 0 {
				// Add double newline for better paragraph separation
				text.WriteString("\n\n")
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extract(c)
			}
		}
	}
	extract(doc)

	return text.String()
}
