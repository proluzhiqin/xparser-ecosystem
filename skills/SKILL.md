---
name: xparse-cli
description: Parse documents (PDF, images, Office, HTML, 20+ formats) into Markdown or structured JSON via the Textin xParse API. Use this skill whenever the user needs to extract text, tables, formulas, or images from any document file.
read_when:
  - Parsing or reading PDF files
  - Converting documents to Markdown or JSON
  - Parsing Word, Excel, or PowerPoint files
  - OCR on scanned documents or images
  - Extracting tables, formulas, or charts from documents
  - Extracting images from documents
  - Batch document processing
  - Converting images (JPG, PNG, TIFF) to text
  - Document parsing errors or troubleshooting
compatibility: Requires the `xparse-cli` binary. Free API works with zero config; paid API optionally requires `XPARSE_APP_ID` and `XPARSE_SECRET_CODE` environment variables.
metadata: {"openclaw":{"emoji":"📄","requires":{"bins":["xparse-cli"]},"install":[{"id":"install-unix","kind":"download","os":["darwin","linux"],"bins":["xparse-cli"],"url":"https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh","label":"Install xparse-cli (Linux/macOS)"},{"id":"install-windows","kind":"download","os":["win32"],"bins":["xparse-cli"],"url":"https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.ps1","label":"Install xparse-cli (Windows)"}]}}
allowed-tools: Bash(xparse-cli:*)
---

# Document Parsing with xparse-cli

## Installation

Check if installed: `xparse-cli version`. If not found, install:

```bash
# Linux / macOS — runs in current shell, PATH takes effect immediately
source <(curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh)
```

```powershell
# Windows (PowerShell)
irm https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.ps1 | iex
```

Verify: `xparse-cli version`

## Quick start

Zero config — free API, no registration needed. Free API supports **PDF and images only**.

```bash
xparse-cli parse report.pdf                          # Markdown → stdout
xparse-cli parse photo.jpg                           # Image OCR → stdout
xparse-cli parse report.pdf --view json              # JSON → stdout
xparse-cli parse report.pdf --output ./result/       # Save to directory
xparse-cli parse report.pdf --page-range "1-5"       # Specific pages
xparse-cli parse secret.pdf --password mypassword    # Encrypted PDF
xparse-cli parse --list files.txt --output ./result/ # Batch mode
```

## Paid API (optional, higher quota)

```bash
xparse-cli auth                                     # Interactive credential setup
```

When running `xparse-cli auth`, the user will be prompted for:
- **App ID** (`x-ti-app-id`) — get from [Textin console](https://www.textin.com/user/login?redirect=%252Fconsole%252Fdashboard%252Fsetting&from=xparse-parse-skill)
- **Secret Code** (`x-ti-secret-code`) — same page

Or set environment variables directly (useful for CI/CD):

```bash
export XPARSE_APP_ID=your_app_id
export XPARSE_SECRET_CODE=your_secret_code
```

| `--api` value | Behavior |
|---------------|----------|
| _(omitted)_ | Paid if credentials exist, else free |
| `free` | Force free API |
| `paid` | Force paid API |

Credential priority: CLI flags → env vars → `~/.xparse-cli/config.yaml`

## Commands

| Command | Description |
|---------|-------------|
| `parse` | Parse a document to Markdown or JSON |
| `auth` | Configure API credentials (interactive) |
| `download` | Download images from parse results |
| `config` | Manage configuration (show / set / reset / path) |
| `update` | Self-update to latest version |
| `version` | Show version information |
| `help` | Show help for any command (e.g. `xparse-cli help parse`) |

## Flags

**parse flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--view` | `markdown` | Output view: `markdown`, `json` |
| `--api` | _(auto)_ | API mode: `free`, `paid` |
| `--page-range` | | Page range: `"1-5"` or `"1-2,5-10"` |
| `--password` | | Password for encrypted documents |
| `--include-char-details` | `false` | Include character-level coordinates and confidence |
| `--list` | | Read input list from file; requires `--output` |
| `--output` | _(stdout)_ | Output file path or directory |

**Global flags (apply to all commands):**

| Flag | Description |
|------|-------------|
| `--app-id` | Textin App ID (overrides env and config) |
| `--secret-code` | Textin Secret Code (overrides env and config) |
| `--base-url` | API base URL (for private deployments) |
| `--verbose` | Print HTTP request/response details (add when retrying errors) |

## Supported formats

| API mode | Max file size | Supported formats |
|----------|---------------|-------------------|
| **free** | 10 MB | PDF, JPG, JPEG, PNG, BMP, TIFF, GIF, WebP |
| **paid** | 500 MB | PDF, JPG, JPEG, PNG, BMP, TIFF, GIF, WebP, DOC, DOCX, XLS, XLSX, CSV, PPT, PPTX, HTML, MHTML, TXT, RTF, OFD |

Both modes support URL input (`http://…`).

If the file is Office/HTML/TXT format and no paid credentials exist, ask human to run `xparse-cli auth` or set credentials; do not silently use free API (it will fail with `40303：文件类型不支持`).

If the file exceeds the size limit, suggest compressing/splitting it or switching to paid API with `--api paid`.

## Download images

```bash
xparse-cli download --from result.json --output ./images/   # From parse result JSON
xparse-cli download <url> --output ./images/                # Direct URL
xparse-cli download <url> --output ./photo.jpg              # Save as specific file
```

> `--output` directory must exist. If not, ask human to run `mkdir -p <dir>` first.

---

## Error handling

CLI output contract:
- **stdout** → document content only (markdown or json)
- **stderr** → error lines (see format below)
- **$?** → exit code: 0 (success), 1 (general error), 2 (bad command), 3 (API error)

### stderr format

Every error prints at least two lines to stderr. The second line tells the agent exactly what to do — follow it directly. Some API errors (when Textin support is needed) print an optional third line with a request ID.

```
<error message>
> <suggestion>
  (request_id: <id>, contact Textin support if unresolved)   ← optional, exit code 3 only
```

| Suggestion tag | Action |
|----------------|--------|
| `[fix]` | Agent fixes the command and re-runs |
| `[retry]` | Agent retries automatically (with backoff) |
| `[fallback]` | Agent tries alternative approach; **check file format first** — free API only supports PDF and images, do not fall back to `--api free` for files |
| `[ask human]` | Requires human intervention — escalate with the suggestion text |

### Capture pattern

```bash
RESULT=$(xparse-cli parse "$FILE" 2>stderr.tmp)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  echo "$RESULT"
else
  # Read the suggestion line and follow it
  SUGGESTION=$(sed -n '2p' stderr.tmp)
  # $SUGGESTION starts with "> " — parse the [tag] to decide action

  # For exit code 3: optional line 3 contains request_id for Textin support
  # Format: "  (request_id: <id>, contact Textin support if unresolved)"
  REQUEST_ID_LINE=$(sed -n '3p' stderr.tmp)
fi
rm -f stderr.tmp
```

### Exit code summary

| Exit code | Meaning | Typical action |
|-----------|---------|----------------|
| 0 | Success | Use stdout |
| 1 | General error (network, filesystem, credentials) | Follow `> suggestion` — usually retry or ask human |
| 2 | Bad command (wrong flag, missing input) | Follow `> [fix]` suggestion — fix and re-run, never retry same command |
| 3 | API error (`api_code：message` on line 1) | Follow `> suggestion` — retry, fallback, fix, or ask human; if line 3 present, include request_id when escalating to support |

## Rules

- **stdout is sacred** — always separate stderr capture; never mix into stdout pipeline
- **Free API = PDF + images only** — for Office/HTML/TXT files, paid API credentials are required
- **Follow the `> suggestion` line** — every error provides a context-specific actionable suggestion on the second line of stderr; this is the primary error-handling mechanism
- **Quote file paths** with spaces: `xparse-cli parse "report 01.pdf"`
- **--output directories must exist** — validated before any API call; ask human to `mkdir -p <dir>` if missing
- **--verbose on retry** — add `--verbose` when retrying network errors
- **auth is human-only** — `xparse-cli auth` is interactive; ask human to run it
- **Upgrade:** `source <(curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh)`

### Default output directory

When user doesn't specify `--output`, suggest saving to:

```
~/xparse-cli/<name>_<hash>/
```

- `<name>`: filename without extension, sanitized
- `<hash>`: first 6 chars of md5 of the full file path
