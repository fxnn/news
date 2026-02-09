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
 
The application uses a unified configuration system with the following lower-to-higher precedence:
1. Defaults
2. Configuration File (`story-extractor.toml`)
3. Environment Variables
4. Command Line Flags
 
**1. Config File**
 
Create a `story-extractor.toml` file (default locations: `./story-extractor.toml`, `$HOME/story-extractor.toml`, or specify via `--config`):
 
```toml
# Global settings
maildir = "/path/to/maildir"
storydir = "/path/to/stories"
verbose = false

[llm]
provider = "openai"
model = "gpt-4o-mini"
api_key = "your-api-key"        # Optional, prefer env var
base_url = "https://api.openai.com/v1"
```
 
**2. Environment Variables**
 
Environment variables override config file values.
 
*   **Prefix**: `STORY_EXTRACTOR_`
*   **Mapping**: `.` and `-` are replaced with `_`
 
Examples:
*   `STORY_EXTRACTOR_LLM_API_KEY` overrides `[llm] api_key`
*   `STORY_EXTRACTOR_MAILDIR` overrides `maildir`
*   `STORY_EXTRACTOR_VERBOSE=true` sets verbose mode
 
**API Key Security (Recommended):**
```bash
export STORY_EXTRACTOR_LLM_API_KEY="your-api-key-here"
```
 
#### Build

```bash
make story-extractor
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
make ui-server
```

#### Usage
 
 ```bash
 ./ui-server --storydir ~/stories
 ```
 
 Or specify a custom port:
 ```bash
 ./ui-server --storydir ~/stories --port 3000
 ```
 
 **Environment Variables:**
 *   **Prefix**: `UI_SERVER_`
 *   Examples: `UI_SERVER_PORT=3000`, `UI_SERVER_STORYDIR=~/stories`
 *   **Config File**: `ui-server.toml` (defaults: `.`, `$HOME`)
 
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

# 2. Build and extract stories from newsletters
make story-extractor
./story-extractor \
  --maildir ~/Maildir/newsletters \
  --storydir ~/stories \
  --config story-extractor.toml

# 3. Start the UI server
make ui-server
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
  --config story-extractor.toml \
  --limit 5 \
  --verbose
```

### Debugging email parsing
```bash
./story-extractor \
  --maildir ~/Maildir/newsletters \
  --storydir ~/stories \
  --config story-extractor.toml \
  --limit 1 \
  --log-headers \
  --log-bodies
```

### Daily newsletter processing
Set up a cron job to run daily:
```bash
# crontab -e
0 8 * * * /usr/local/bin/mbsync -a && /path/to/story-extractor --maildir ~/Maildir/newsletters --storydir ~/stories --config ~/story-extractor.toml
```

## Development

This project is written in Go and every single line of code has been created with AI assistance.

Run `make help` for available targets. `make` on its own formats, vets, tests, and builds everything.

## License

[MIT License](LICENSE)

---

*An experiment in newsletter consumption and AI-assisted development.*
