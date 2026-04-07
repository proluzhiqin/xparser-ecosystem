---
name: xparser
description: Textin xParser document parsing CLI that converts PDFs, images, Office documents and 20+ file formats into Markdown and structured JSON via the Textin xParser API. Designed as Agent infrastructure — zero config, free API default, structured errors, stdout-friendly.
read_when:
  - Extracting text from PDF documents
  - Converting documents to Markdown
  - Converting PDF to JSON
  - Parsing Word or Excel or PowerPoint files
  - OCR on scanned documents or images
  - Extracting tables from documents
  - Extracting formulas from documents
  - Extracting charts from documents
  - Extracting images from documents
  - Batch document processing
  - Converting images (JPG, PNG, TIFF) to text
  - Reading or parsing PDF files
  - Converting Word documents to Markdown
  - Document parsing errors or troubleshooting
metadata: {"openclaw":{"emoji":"📄","requires":{"bins":["xparse-cli"]},"install":[{"id":"install-unix","kind":"download","os":["darwin","linux"],"bins":["xparse-cli"],"url":"https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh","label":"Install xparse-cli (Linux/macOS)"},{"id":"install-windows","kind":"download","os":["win32"],"bins":["xparse-cli"],"url":"https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.ps1","label":"Install xparse-cli (Windows)"}]}}
allowed-tools: Bash(xparse-cli:*)
---

# Document Parsing with xparse-cli

## Installation

### Linux / macOS

```bash
curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.ps1 | iex
```

### Verify installation

```bash
xparse-cli version
```

## Core concept: View

Agent only needs to understand one concept: **View** — different presentations of the same parse result.

| View | Description |
|------|-------------|
| `markdown` (default) | Markdown text, ready for LLM consumption |
| `json` | Full structured JSON with all parse details |

## Quick start — zero config

xparse-cli works out of the box with no API key — it defaults to the free API:

```bash
# Markdown to stdout (free API, zero config)
xparse-cli parse report.pdf

# JSON view
xparse-cli parse report.pdf --view json

# Save to directory
xparse-cli parse report.pdf --output ./result/

# Parse specific pages
xparse-cli parse report.pdf --page-range "1-5"

# Batch from file list
xparse-cli parse --list files.txt --output ./result/
```

No registration, no API key, no configuration needed for first use.

## Authentication (for paid API)

The paid API offers higher quotas and priority. Get credentials at <https://www.textin.com/console/dashboard/setting>.

| Env var | Textin credential |
|---------|-------------------|
| `XPARSE_APP_ID` | `x-ti-app-id` |
| `XPARSE_SECRET_CODE` | `x-ti-secret-code` |

```bash
# Interactive setup (saves to ~/.xparse-cli/config.yaml)
xparse-cli auth

# Or set environment variables
export XPARSE_APP_ID=your_app_id
export XPARSE_SECRET_CODE=your_secret_code

# Use paid API explicitly
xparse-cli parse report.pdf --api paid
```

**API selection logic:**
- No `--api` flag: if credentials exist, uses paid; otherwise free
- `--api free`: forces free API regardless of credentials
- `--api paid`: forces paid API (requires credentials)

## Commands

### parse — The main command

```bash
xparse-cli parse <file-or-url>
xparse-cli parse <file-or-url> --output <file_path>
xparse-cli parse --list files.txt --output ./result/
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--view` | | `markdown` | Output view: `markdown`, `json` |
| `--api` | | _(auto)_ | API mode: `free`, `paid` |
| `--page-range` | | | Page range: `"1-5"` or `"1-2,5-10"` |
| `--password` | | | Password for encrypted documents |
| `--include-char-details` | | `false` | Include character-level coordinates and confidence |
| `--list` | | | Read input list from file (one path per line); requires `--output` |
| `--output` | `-o` | _(stdout)_ | Output file path or directory |

#### API capabilities (always on by default)

These capabilities are automatically enabled — Agent gets the most complete result without configuration:

| Capability | Default |
|------------|---------|
| Heading hierarchy | On |
| Inline objects (images in text) | On |
| Image data | On |
| Table structure (HTML) | On |
| Per-page results | On |
| Title tree / catalog | On |
| Character details | **Off** (use `--include-char-details` to enable) |

### Supported input formats

| Category | Extensions |
|----------|-----------|
| PDF | `.pdf` |
| Images | `.jpg`, `.jpeg`, `.png`, `.bmp`, `.tiff`, `.gif`, `.webp` |
| Word | `.doc`, `.docx` |
| Excel | `.xls`, `.xlsx`, `.csv` |
| PowerPoint | `.ppt`, `.pptx` |
| Web | `.html`, `.mhtml` |
| Text | `.txt`, `.rtf` |
| Other | `.ofd` |
| URL | `http://…`, `https://…` |

### download — Fetch images

```bash
xparse-cli download --from result.json -o ./images/
xparse-cli download <image_id> -o ./images/
```

### config — Manage settings

```bash
xparse-cli config show
xparse-cli config set base_url https://my-server
xparse-cli config reset
xparse-cli config path
```

### update — Self-update

```bash
xparse-cli update
```

## Recipes

### Read document content (Scene 1)

```bash
# Default: markdown text to stdout
xparse-cli parse report.pdf

# Full JSON for programmatic use
xparse-cli parse report.pdf --view json
```

### Parse specific pages (Scene 2)

```bash
xparse-cli parse book.pdf --page-range "1-5"
xparse-cli parse book.pdf --page-range "1-2,5-10"
```

### Encrypted PDF (Scene 3)

```bash
xparse-cli parse secret.pdf --password mypassword
```

### Character-level details (Scene 4)

```bash
# Warning: significantly larger response
xparse-cli parse report.pdf --view json --include-char-details
```

### Save output to file

```bash
# Save to directory (filename derived from input)
xparse-cli parse report.pdf --output ./result/

# Save to specific file
xparse-cli parse report.pdf --output report.md
```

### Batch processing

```bash
# From a file list
xparse-cli parse --list files.txt --output ./results/
```

Batch mode requires `--output <directory>`. Progress is reported on stderr as `[n/total]`.

### Piping into other tools

xParser outputs content to stdout, status/logs to stderr — clean piping:

```bash
# Parse and search
xparse-cli parse report.pdf | grep "revenue"

# Parse and feed to LLM
xparse-cli parse paper.pdf | llm "summarize this paper"
```

## Exit codes

| Code | Meaning | Recovery |
|------|---------|----------|
| 0 | Success | — |
| 1 | General API or unknown error | Check network connectivity; retry; use `--verbose` for details |
| 2 | Invalid parameters / usage error | Check command syntax and flag values |
| 3 | API returned error (structured JSON on stderr) | Parse stderr JSON for error_type, suggestion, retryable |

### Structured error output (exit code 3)

When exit code is 3, stderr contains a JSON object:

```json
{
  "error_type": "password_error",
  "code": 40422,
  "message": "Password required or incorrect",
  "suggestion": "Use --password to provide the correct document password",
  "retryable": false
}
```

Fields:
- `error_type`: category — `auth_error`, `password_error`, `page_range_error`, `format_error`, `file_error`, `param_error`, `parse_error`, `server_error`, `partial_error`, `unknown_error`
- `code`: API error code
- `message`: human-readable description
- `suggestion`: recommended fix action
- `retryable`: whether Agent should retry

### API error code reference

| API code | error_type | Description |
|----------|-----------|-------------|
| 40101 | auth_error | Missing credentials |
| 40102 | auth_error | Invalid credentials |
| 40103 | auth_error | Quota exceeded or account suspended |
| 40422 | password_error | Password required or incorrect |
| 40424 | page_range_error | Page number out of range |
| 40425 | format_error | File format not supported |
| 40426 | file_error | File is corrupted or unreadable |
| 40427 | param_error | Invalid DPI value |
| 40428 | parse_error | Office file conversion failed |
| 40429 | file_error | PDF content is empty |
| 500 | server_error | Server internal error |
| 50207 | partial_error | Some pages failed to parse |

## Output behavior

- **No `--output`**: result goes to stdout; status/progress on stderr
- **With `--output`**: result saved to file/directory; progress on stderr
- **Batch mode**: requires `--output <directory>`

## General rules

When using this skill on behalf of the user:

- **Quote file paths** with spaces: `xparse-cli parse "report 01.pdf"`
- **Don't run commands blindly on errors** — parse the structured error JSON and suggest a fix
- On exit code 3, read the `retryable` field to decide whether to retry or ask the user
- On exit code 1, add `--verbose` flag to diagnose network issues
- When the user asks to **upgrade**, run the install command first

### Default output directory

When saving output on behalf of the user and no `--output` is specified, generate:

```
~/xparse-cli/<name>_<hash>/
```

**Naming rules:**

- `<name>`: filename without extension, sanitized (replace spaces and shell-unsafe characters with `_`)
- `<hash>`: first 6 characters of MD5 of the full source path

```bash
echo -n "<full_source_path>" | md5sum | cut -c1-6
```

| Source | Output directory |
|--------|-----------------|
| `report.pdf` | `~/xparse-cli/report_f1a2b3/` |

When the user specifies `--output`, use their path as-is.

### Post-parse hints

After a successful parse, the agent MAY append ONE brief hint (don't repeat in the same session):

- If `--view json` not used: "Use `--view json` for full structured data."
- If parsing a large PDF: "Use `--page-range` to parse specific pages."
- If error 40422: "This PDF is encrypted. Use `--password` to provide the password."

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| No output, exit 1 | Check network; retry; add `--verbose` |
| Exit 2 | Check flag names and values |
| Exit 3 with password_error | Add `--password` |
| Exit 3 with page_range_error | Adjust `--page-range` |
| Exit 3 with format_error | Check file format (see supported list) |
| Want paid API but no credentials | `xparse-cli auth` or set env vars |
| Large response needed | Add `--include-char-details` only when needed |
| Batch partially failed | Check stderr for per-file errors |

## Notes

- All status/progress messages go to stderr; only document content goes to stdout
- Free API works with zero configuration — no registration needed
- Paid API requires `XPARSE_APP_ID` + `XPARSE_SECRET_CODE`
- Credentials stored in `~/.xparse-cli/config.yaml` after `xparse-cli auth`
- Token resolution: `--app-id`/`--secret-code` flags > env vars > config file
