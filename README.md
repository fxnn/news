# news

A different way to read email newsletters.

## About

`news` is a specialized email client designed for newsletter consumption. Following the UNIX philosophy, it consists of three orthogonal tools:

1. **Mail Downloader**: Downloads emails to a Maildir directory (using existing tools)
2. **Story Extractor**: Extracts stories from newsletter emails using AI
3. **UI Server**: Serves a web interface to browse and read stories

Currently in **early development stage**.

See [SKETCH.md](SKETCH.md) for detailed vision and project plans.

## Setup

### 1. Mail Downloader

The system uses standard email synchronization tools to download newsletters into a Maildir directory. You can use any of the following tools:

#### Option A: mbsync (recommended)

Install via your package manager:
```bash
# macOS
brew install isync

# Debian/Ubuntu
apt install isync
```

Create `~/.mbsyncrc`:
```
IMAPAccount newsletter
Host imap.example.com
User your-email@example.com
PassCmd "security find-generic-password -s mbsync-newsletter -w"
SSLType IMAPS

IMAPStore newsletter-remote
Account newsletter

MaildirStore newsletter-local
Path ~/Maildir/newsletters/
Inbox ~/Maildir/newsletters/INBOX

Channel newsletter
Far :newsletter-remote:
Near :newsletter-local:
Patterns *
Create Near
Sync All
```

Sync emails:
```bash
mbsync -a
```

#### Option B: offlineimap

Install:
```bash
# macOS
brew install offlineimap

# Debian/Ubuntu
apt install offlineimap
```

Create `~/.offlineimaprc`:
```ini
[general]
accounts = newsletter

[Account newsletter]
localrepository = newsletter-local
remoterepository = newsletter-remote

[Repository newsletter-local]
type = Maildir
localfolders = ~/Maildir/newsletters

[Repository newsletter-remote]
type = IMAP
remotehost = imap.example.com
remoteuser = your-email@example.com
remotepass = your-password
ssl = yes
```

Sync emails:
```bash
offlineimap
```

#### Option C: fetchmail

Install:
```bash
# macOS
brew install fetchmail

# Debian/Ubuntu
apt install fetchmail
```

Create `~/.fetchmailrc`:
```
poll imap.example.com
  protocol IMAP
  username "your-email@example.com"
  password "your-password"
  ssl
  mda "/usr/bin/procmail -d %s"
```

### 2. Story Extractor

Build and run:
```bash
go build ./cmd/story-extractor
./story-extractor --maildir ~/Maildir/newsletters --storydir ~/stories --config config.toml
```

### 3. UI Server

Build and run:
```bash
go build ./cmd/ui-server
./ui-server --storydir ~/stories
```

## Development

This project is written in Go and every single line of code has been created with AI assistance.

## License

[MIT License](LICENSE)

---

*An experiment in newsletter consumption and AI-assisted development.*
