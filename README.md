# news

A different way to read email newsletters.

## About

`news` is a specialized email client designed for newsletter consumption. It connects to your email via IMAP, extracts newsletter content, and presents it in a clean, focused interface.

Currently in **early development stage**.

See [SKETCH.md](SKETCH.md) for detailed vision and project plans.

## Quick Start

```bash
# Build the project
go build

# Run the application
./news -server imap.example.com -port 993 -username your_username -password your_password -folder INBOX -days 7
```

## Development

This project is written in Go and every single line of code has been created with AI assistance.

## License

[MIT License](LICENSE)

---

*An experiment in newsletter consumption and AI-assisted development.*
