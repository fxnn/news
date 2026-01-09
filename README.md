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

The story extractor processes newsletter emails using AI to extract individual news stories.

#### Configuration

Create a `config.toml` file (copy from `config.example.toml`):

```toml
[llm]
provider = "openai"
model = "gpt-4o-mini"
api_key = "your-api-key-here"
base_url = "https://api.openai.com/v1"
```

Supported models:
- `gpt-4o-mini` (recommended, fast and cost-effective)
- `gpt-4o` (higher quality, more expensive)

#### Build

```bash
go build ./cmd/story-extractor
```

#### Usage

Basic usage:
```bash
./story-extractor \
  --maildir ~/Maildir/newsletters \
  --storydir ~/stories \
  --config config.toml
```

#### CLI Flags

Required:
- `--maildir`: Path to the Maildir directory containing newsletters
- `--storydir`: Path to the directory where stories will be saved as JSON files
- `--config`: Path to the TOML configuration file with LLM settings

Optional:
- `--limit N`: Process maximum N emails (useful for testing)
- `--verbose`: Enable verbose logging
- `--log-headers`: Log email headers (for debugging)
- `--log-bodies`: Log email bodies (for debugging)
- `--log-stories`: Log extracted stories

#### How It Works

1. Reads emails from the Maildir directory (recursively scans `cur/` and `new/` subdirectories)
2. Parses email headers, body (plain text, HTML, multipart MIME)
3. Sends each email to the configured LLM with a prompt to extract news stories
4. Saves each story as a JSON file: `<date>_<message-id>_<index>.json`
5. Skips emails that have already been processed (incremental processing)

Example story file (`2006-01-02_test@example.com_1.json`):
```json
{
  "headline": "Example News Headline",
  "teaser": "Brief summary of the article in 1-2 sentences.",
  "url": "https://example.com/article",
  "from_email": "newsletter@example.com",
  "from_name": "Example Newsletter",
  "date": "2006-01-02T15:04:05Z"
}
```

### 3. UI Server

The UI server provides a web interface to browse and read extracted stories.

#### Build

```bash
go build ./cmd/ui-server
```

#### Usage

```bash
./ui-server --storydir ~/stories
```

Or specify a custom port:
```bash
./ui-server --storydir ~/stories --port 3000
```

#### CLI Flags

Required:
- `--storydir`: Path to the directory containing story JSON files

Optional:
- `--port`: Port to listen on (default: 8080)

#### Access

Open your browser and navigate to:
```
http://localhost:8080
```

The UI displays:
- All extracted stories sorted by date (newest first)
- Story headline (clickable link to original article)
- Brief teaser text
- Source newsletter (sender name/email)
- Publication date (shown as relative time: "Today", "2 days ago", etc.)

## Quick Start

Complete workflow from setup to reading stories:

```bash
# 1. Download newsletters using mbsync
mbsync -a

# 2. Extract stories from newsletters
go build ./cmd/story-extractor
./story-extractor \
  --maildir ~/Maildir/newsletters \
  --storydir ~/stories \
  --config config.toml

# 3. Start the UI server
go build ./cmd/ui-server
./ui-server --storydir ~/stories

# 4. Open in browser
open http://localhost:8080
```

## Usage Examples

### Testing with a small batch
```bash
./story-extractor \
  --maildir ~/Maildir/newsletters \
  --storydir ~/stories \
  --config config.toml \
  --limit 5 \
  --verbose
```

### Debugging email parsing
```bash
./story-extractor \
  --maildir ~/Maildir/newsletters \
  --storydir ~/stories \
  --config config.toml \
  --limit 1 \
  --log-headers \
  --log-bodies
```

### Daily newsletter processing
Set up a cron job to run daily:
```bash
# crontab -e
0 8 * * * /usr/local/bin/mbsync -a && /path/to/story-extractor --maildir ~/Maildir/newsletters --storydir ~/stories --config ~/config.toml
```

## Development

This project is written in Go and every single line of code has been created with AI assistance.

## License

[MIT License](LICENSE)

---

*An experiment in newsletter consumption and AI-assisted development.*
