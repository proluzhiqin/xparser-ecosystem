---
name: xparser
description: Textin xParser document parsing CLI that converts PDFs, images, Office documents and 20+ file formats into Markdown and structured JSON via the Textin xParser API. Supports OCR, table extraction, formula recognition, and image extraction. Zero config, free API default, structured errors, stdout-friendly.
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

xparse-cli works out of the box with no API key — it defaults to the free API.
No registration, no API key, no configuration needed for first use.

```bash
xparse-cli parse report.pdf                        # Markdown to stdout
xparse-cli parse report.pdf --view json             # JSON view
xparse-cli parse report.pdf --output ./result/      # Save to directory
xparse-cli parse report.pdf --page-range "1-5"      # Specific pages
xparse-cli parse --list files.txt --output ./result/ # Batch mode
```

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
| `--verbose` | `-v` | `false` | Print HTTP request details for debugging |

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

### download — Download element images

```bash
xparse-cli download --from result.json -o ./images/
xparse-cli download <image_url> -o ./images/
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

```bash
# Encrypted PDF
xparse-cli parse secret.pdf --password mypassword

# Character-level details (significantly larger response)
xparse-cli parse report.pdf --view json --include-char-details

# Non-contiguous page ranges
xparse-cli parse book.pdf --page-range "1-2,5-10"

# Save to specific file
xparse-cli parse report.pdf --output report.md

# Piping — stdout is content only, stderr is status/logs
xparse-cli parse report.pdf | grep "revenue"
xparse-cli parse paper.pdf | llm "summarize this paper"
```

Batch mode requires `--output <directory>`. Progress is reported on stderr as `[n/total]`.

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
  "code": 40423,
  "message": "Password required or incorrect password",
  "suggestion": "Use --password to provide the correct document password",
  "retryable": false
}
```

Fields:
- `error_type`: category — `auth_error`, `quota_error`, `rate_limit_error`, `param_error`, `format_error`, `file_error`, `password_error`, `page_range_error`, `parse_error`, `server_error`, `partial_error`, `unknown_error`
- `code`: API error code
- `message`: human-readable description
- `suggestion`: recommended fix action
- `retryable`: whether Agent should retry

### API error code reference

| API code | error_type | Description |
|----------|-----------|-------------|
| 500 | server_error | Server internal error |
| 30203 | server_error | Base service failure |
| 40003 | quota_error | Insufficient balance |
| 40004 | param_error | Invalid parameter |
| 40007 | param_error | Service not found or not published |
| 40008 | param_error | Service not activated |
| 40101 | auth_error | Credentials empty |
| 40102 | auth_error | Authentication failed |
| 40103 | auth_error | IP not whitelisted |
| 40301 | format_error | Image type not supported |
| 40302 | file_error | File size exceeds limit (500MB) |
| 40303 | format_error | File type not supported |
| 40304 | file_error | Image dimensions out of range |
| 40305 | file_error | No file uploaded |
| 40306 | rate_limit_error | QPS limit exceeded |
| 40307 | quota_error | Daily free quota exhausted |
| 40400 | param_error | Invalid request URL |
| 40422 | file_error | The file is corrupted |
| 40423 | password_error | Password required or incorrect password |
| 40424 | page_range_error | Page number out of range |
| 40425 | format_error | The input file format is not supported |
| 40427 | param_error | DPI value is not in the allowed list |
| 40428 | parse_error | Office file conversion failed or timed out |
| 40429 | parse_error | Unsupported engine |
| 50207 | partial_error | Some pages failed to parse |

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
- If error 40423: "This PDF is encrypted. Use `--password` to provide the password."

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| No output, exit 1 | Check network; retry; add `--verbose` |
| Exit 2 | Check flag names and values |
| Exit 3 with quota_error (40003) | Balance depleted — top up or use `--api free` |
| Exit 3 with quota_error (40307) | Daily free quota used up — try tomorrow or `--api paid` |
| Exit 3 with rate_limit_error (40306) | Too many requests — wait and retry |
| Exit 3 with file_error (40302) | File exceeds 500MB — split or compress |
| Exit 3 with file_error (40422) | File is corrupted — re-download or use a different copy |
| Exit 3 with password_error (40423) | Add `--password` with the correct password |
| Exit 3 with page_range_error (40424) | Adjust `--page-range` to valid range |
| Exit 3 with format_error (40303/40425) | Check file format (see supported list) |
| Exit 3 with parse_error (40428) | Retry; or convert Office file to PDF first |
| Want paid API but no credentials | `xparse-cli auth` or set env vars |
| Batch partially failed | Check stderr for per-file errors |

## Notes

- All status/progress messages go to stderr; only document content goes to stdout
- Free API works with zero configuration — no registration needed
- Paid API requires `XPARSE_APP_ID` + `XPARSE_SECRET_CODE`
- Credentials stored in `~/.xparse-cli/config.yaml` after `xparse-cli auth`
- Token resolution: `--app-id`/`--secret-code` flags > env vars > config file
