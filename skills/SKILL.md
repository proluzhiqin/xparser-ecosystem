---
name: xparser
description: Textin xParser document parsing CLI that converts PDFs, images, Office documents and 20+ file formats into Markdown and structured JSON via the Textin xParser API. Supports multiple parse modes (auto, scan, lite, parse, vlm), table/formula/chart recognition, image extraction, batch processing, and piped workflows.
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
  - VLM document parsing
metadata: {"openclaw":{"emoji":"📄","requires":{"bins":["xparser"]},"install":[{"id":"install-unix","kind":"download","os":["darwin","linux"],"bins":["xparser"],"url":"https://dllf.intsig.net/download/2026/Solution/xparser/install.sh","label":"Install xparser CLI (Linux/macOS)"},{"id":"install-windows","kind":"download","os":["win32"],"bins":["xparser"],"url":"https://dllf.intsig.net/download/2026/Solution/xparser/install.ps1","label":"Install xparser CLI (Windows)"}]}}
allowed-tools: Bash(xparser:*)
---

# Document Parsing with xparser

## Installation

### Linux / macOS

```bash
curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparser/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://dllf.intsig.net/download/2026/Solution/xparser/install.ps1 | iex
```

### Verify installation

```bash
xparser version
```

## Five parse modes

| Mode | Speed | Best for |
|------|-------|----------|
| `scan` (**default**) | Medium | Scanned docs, images, mixed documents — treats everything as images |
| `parse` | Fastest | Born-digital PDFs only (not scanned), no formula/chart support |
| `vlm` | Slow | Complex layouts, academic papers, mixed content — highest accuracy |
| `lite` | Fast | Lightweight — tables and text only, no full layout analysis |
| `auto` | Auto | Unsure about document type — engine selects automatically |

### Mode comparison

| | `scan` | `parse` | `vlm` | `lite` | `auto` |
|---|---|---|---|---|---|
| Scanned docs / images | Yes | No | Yes | Yes | Yes |
| Electronic PDFs | Yes | **Best** | Yes | Yes | Yes |
| Table recognition | Yes | Yes | Yes | Yes | Yes |
| Formula recognition | Yes | No | Yes | No | Yes |
| Hallucination risk | None | None | **Yes** | None | None |

## Core workflow

1. **Install**: run the install command above
2. **Authenticate**: `xparser auth` (requires Textin API credentials from <https://www.textin.com/console/dashboard/setting>)
3. **Parse**: `xparser parse report.pdf` for Markdown output
4. **Save output**: add `-o ./out/` to save to directory
5. **Extract images**: parse with `--get-image objects -f json`, then `xparser download --from result.json`

## Authentication

Every parse call needs Textin API credentials. Get them at <https://www.textin.com/console/dashboard/setting>.

| CLI parameter / env var | Textin credential |
|-------------------------|-------------------|
| `--app-id` / `XPARSER_APP_ID` | `x-ti-app-id` |
| `--secret-code` / `XPARSER_SECRET_CODE` | `x-ti-secret-code` |

When prompting the user for credentials, mention the Textin names (`x-ti-app-id`, `x-ti-secret-code`) so they know which values to copy from the console.

```bash
xparser auth                          # interactive setup → saves to ~/.xparser/config.yaml
xparser auth --show                   # show current credentials (masked)
```

Token resolution order: `--app-id`/`--secret-code` flags > `XPARSER_APP_ID`/`XPARSER_SECRET_CODE` env > `~/.xparser/config.yaml`.

For private deployments, set `--base-url https://your-server.com` or `xparser config set base_url https://your-server.com`.

## Supported input formats

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

**Limits:** files ≤ 500 MB, PDFs ≤ 1000 pages, Excel ≤ 2000 rows / 100 cols per sheet, TXT ≤ 100 KB, images 20–20000 px (aspect ratio < 2) or 20–10000 px (others).

## Commands

### parse — The main command

```bash
xparser parse report.pdf                              # Markdown to stdout
xparser parse report.pdf -o report.md                 # Save to file
xparser parse report.pdf -o ./out/                    # Save to directory
xparser parse report.pdf -f json                      # Full JSON output
xparser parse report.pdf -f md,json -o ./out/         # Both formats (needs -o dir)
xparser parse report.pdf --parse-mode vlm             # VLM mode
xparser parse *.pdf -o ./results/                     # Batch convert
xparser parse https://example.com/doc.pdf             # Parse a URL
cat doc.pdf | xparser parse --stdin -o result.md      # Pipe in
```

#### parse flags

**Output:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | _(stdout)_ | Output path (file or directory) |
| `--format` | `-f` | `md` | Output formats: `md`, `json` (comma-separated) |

**Parse mode & pages:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--parse-mode` | | `scan` | Parse mode: `auto`, `scan`, `lite`, `parse`, `vlm` |
| `--pdf-pwd` | | | Password for encrypted PDFs |
| `--page-start` | | `0` | Start page (0-indexed) |
| `--page-count` | | `1000` | Number of pages to parse (max 1000) |
| `--dpi` | | `144` | PDF coordinate DPI: `72`, `144`, `216` |

**Tables & structure:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--table-flavor` | | `html` | Table format in Markdown: `html`, `md`, `none` |
| `--apply-document-tree` | | `1` | Generate heading hierarchy: `0` off, `1` on |
| `--paratext-mode` | | `annotation` | Non-body text display: `none`, `annotation`, `body` |
| `--apply-merge` | | `1` | Merge paragraphs and tables: `0` off, `1` on |

**Images:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--get-image` | | `none` | Image extraction: `none`, `page`, `objects`, `both` |
| `--image-output-type` | | `default` | Image output type: `default` (URL/ID), `base64str` |

**Recognition:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--formula-level` | | `0` | Formula recognition: `0` all, `1` display only, `2` off |
| `--apply-chart` | | `0` | Chart recognition (output as table): `0` off, `1` on |
| `--underline-level` | | `0` | Underline detection (scan mode only): `0` off, `1` empty lines, `2` all |
| `--apply-image-analysis` | | `0` | LLM-based image analysis: `0` off, `1` on |
| `--crop-dewarp` | | `0` | Crop and deskew: `0` off, `1` on |
| `--remove-watermark` | | `0` | Watermark removal: `0` off, `1` on |

**Output detail:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--markdown-details` | | `1` | Return `detail` field with element info: `0` off, `1` on |
| `--page-details` | | `1` | Return `pages` field with per-page results: `0` off, `1` on |
| `--raw-ocr` | | `0` | Return full OCR results with char coordinates: `0` off, `1` on |
| `--char-details` | | `0` | Return `char_pos` field with character positions: `0` off, `1` on |
| `--catalog-details` | | `0` | Return `catalog` field with TOC info: `0` off, `1` on |
| `--get-excel` | | `0` | Return Excel base64 result: `0` off, `1` on |

**Input modes:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--list` | | | Read input list from file (one path per line) |
| `--stdin-list` | | `false` | Read file list from stdin |
| `--stdin` | | `false` | Read file content from stdin |
| `--stdin-name` | | `stdin.pdf` | Filename hint for stdin mode |
| `--timeout` | | `600` | Timeout in seconds |

### download — Fetch images

```bash
xparser download --from result.json -o ./images/  # From parse result JSON (recommended)
xparser download <image_id> -o ./images/           # By image ID
xparser download id1 id2 id3 -o ./images/          # Batch by IDs
xparser download --from result.json extra_id -o .  # Mixed: JSON + manual IDs
```

`--from` automatically extracts image IDs from `metrics[].image_id`, `result.pages[].image_id`, `result.detail[].image_url`, deduplicates, and downloads each one.

Image IDs are valid for **30 days** after parsing.

### config — Manage settings

```bash
xparser config show                              # Display all settings
xparser config set base_url https://my-server    # Set a value
xparser config set app_id YOUR_ID                # Set app ID
xparser config reset                             # Reset to defaults
xparser config path                              # Show config file location
```

Keys: `app_id`, `secret_code`, `base_url`.

### update — Self-update

```bash
xparser update              # Download and install latest version
```

### version

```bash
xparser version             # Show version, commit, Go version, OS/arch
```

## Recipes

### Batch processing

```bash
# All PDFs in current directory
xparser parse *.pdf -o ./results/

# From a file list
xparser parse --list files.txt -o ./results/

# From find/pipe
find ./docs -name "*.pdf" | xparser parse --stdin-list -o ./results/
```

Batch mode always requires `-o <directory>`. Progress is reported on stderr as `[n/total]`.

### Image extraction

Two steps: parse with `--get-image` to get JSON result, then use `--from` to batch download.

```bash
# Step 1: parse with image extraction, save JSON result
xparser parse report.pdf --get-image objects -f json -o ./out/

# Step 2: download all images directly from the JSON result
xparser download --from ./out/report.json -o ./out/images/
```

For inline base64 images (larger response, but no second step):
```bash
xparser parse report.pdf --get-image objects --image-output-type base64str -o ./out/
```

### Tables

Tables default to HTML. For Markdown tables:
```bash
xparser parse report.pdf --table-flavor md
```

Use `--table-flavor none` to skip table recognition entirely.

### Formulas and charts

```bash
xparser parse paper.pdf --formula-level 0              # All formulas (default)
xparser parse paper.pdf --formula-level 1              # Display formulas only
xparser parse paper.pdf --apply-chart 1                # Charts → tables
```

### PDF page ranges

```bash
xparser parse book.pdf --page-start 10 --page-count 10   # Pages 10–19
xparser parse secret.pdf --pdf-pwd mypassword             # Encrypted PDF
```

### Piping into other tools

xParser outputs Markdown to stdout by default (status goes to stderr), so it pipes cleanly:

```bash
# Parse and search
xparser parse report.pdf | grep "revenue"

# Parse and feed to another model
xparser parse paper.pdf --parse-mode vlm | llm "summarize this paper"

# Parse from stdin
cat report.pdf | xparser parse --stdin
```

### JSON output for programmatic use

```bash
# Full structured JSON
xparser parse report.pdf -f json

# Minimal JSON (no detail/pages arrays — smaller payload)
xparser parse report.pdf -f json --markdown-details 0 --page-details 0
```

**Note:** `lite` and `vlm` modes return an `elements` array instead of the `detail`/`pages` structure used by other modes.

## Output behavior

- **No `-o` flag**: result goes to stdout; status/progress messages go to stderr
- **With `-o` flag**: result saved to file/directory; progress on stderr
- **`-f md,json`**: requires `-o <directory>`
- **`--get-excel 1`**: Excel file saved alongside Markdown
- **Batch mode**: requires `-o <directory>`

## General rules

When using this skill on behalf of the user:

- **Quote file paths** that contain spaces or special characters with double quotes. Example: `xparser parse "report 01.pdf"`, NOT `xparser parse report 01.pdf`.
- **Don't run commands blindly on errors** — explain the exit code and suggest a fix.
- On failure, use `-v` to show HTTP details if the cause is unclear.
- When the user asks to **upgrade** or **update**, run the install command to get the latest binary before using new features.

### Choosing between modes

The agent MUST follow this decision logic:

1. **Default to `scan`** when:
   - User says nothing special about mode
   - Document is a scanned PDF or image
   - Mixed document types (scanned + electronic)

2. **Use `parse`** when:
   - User explicitly mentions "electronic PDF" or wants fastest speed
   - Document is confirmed to be a born-digital PDF (not scanned)
   - No formula or chart recognition needed

3. **Use `vlm`** when:
   - User mentions "VLM" or wants highest accuracy
   - Document has complex layouts, academic papers, intricate tables
   - User prioritizes accuracy over speed

4. **Use `lite`** when:
   - User only needs tables and text, mentions "lightweight"
   - No formula, chart, or full layout analysis needed
   - Speed is important but document may include scanned content

5. **Use `auto`** when:
   - User explicitly says "auto" or is unsure about document type
   - Processing a batch of mixed document types

**Note:** `vlm` mode may produce hallucinated text in rare cases. When the user prioritizes reliability and no-hallucination guarantee, suggest `scan` (default) instead. When accuracy on complex layouts matters most, suggest `vlm` with this caveat.

### Default output directory

When the agent saves output on behalf of the user and no `-o` is specified, generate:

```
~/xParser-Skill/<name>_<hash>/
```

**Naming rules:**

- `<name>`: derived from the source, then **sanitized** for safe directory names.
  - For URLs: last path segment (e.g. `https://example.com/doc.pdf` → `doc`)
  - For local files: filename without extension (e.g. `report.pdf` → `report`)
  - **Sanitization**: replace spaces and shell-unsafe characters (`space`, `(`, `)`, `[`, `]`, `&`, `'`, `"`, `!`, `#`, `$`, `` ` ``) with `_`. Collapse consecutive `_` into one. Keep alphanumeric, `-`, `_`, `.`, and CJK characters.
- `<hash>`: first 6 characters of the MD5 hash of the **full original source path or URL** (before sanitization). This ensures:
  - Different URLs with similar basenames get unique directories
  - Re-running the same source reuses the same directory (idempotent)

**How the agent should generate the hash:**

```bash
echo -n "<full_source_path>" | md5sum | cut -c1-6
```

**Examples:**

| Source | Output directory |
|--------|-----------------|
| `report.pdf` | `~/xParser-Skill/report_f1a2b3/` |
| `./docs/年报2024.pdf` | `~/xParser-Skill/年报2024_c7e9d4/` |
| `https://example.com/q1.pdf` | `~/xParser-Skill/q1_a3f2b1/` |

**When the user specifies `-o`**: use the user's path as-is, do NOT override with the default directory.

### Post-parse friendly hints

After a successful parse, the agent MAY append a brief hint (one per session, don't repeat):

- If no `--get-image`: "Add `--get-image objects` to also extract images."
- If `--table-flavor` not set: "Tables default to HTML; use `--table-flavor md` for Markdown."
- If `scan` mode on a simple electronic PDF: "`--parse-mode parse` would be faster for this type of document."

Keep the hint to ONE short sentence. Do NOT repeat if the user has already seen it in this session.

## Exit codes

| Code | Meaning | Recovery |
|------|---------|----------|
| 0 | Success | — |
| 1 | General error | Check network; retry; add `-v` for HTTP debug |
| 2 | Bad parameters | Check flag names and values |
| 3 | Auth failure | Run `xparser auth` to configure credentials |
| 4 | File error | Check file type/size/dimensions, or supply `--pdf-pwd` |
| 5 | Parse failed | Document may be corrupted; try a different `--parse-mode` |
| 6 | Server error | Textin service issue; retry later |
| 7 | Quota exhausted | Recharge at <https://www.textin.com> |

### API error code mapping

| API code | Exit | Description |
|----------|------|-------------|
| 200 | 0 | Success |
| 40101 | 3 | Missing app-id or secret-code |
| 40102 | 3 | Invalid credentials |
| 40103 | 3 | IP not whitelisted |
| 40003 | 7 | Insufficient balance |
| 40004 | 2 | Parameter error |
| 40301 | 4 | Image type unsupported |
| 40302 | 4 | File exceeds 500 MB |
| 40303 | 4 | File type unsupported |
| 40304 | 4 | Image dimensions out of range |
| 40305 | 4 | No file uploaded |
| 40422 | 4 | File corrupted |
| 40423 | 4 | PDF password required / incorrect |
| 40424 | 4 | Page number out of range |
| 40425 | 4 | Input file format not supported |
| 40427 | 2 | Invalid DPI value |
| 40428 | 5 | Office conversion failed |
| 50207 | 5 | Partial pages failed |
| 30203 | 6 | Server error, retry later |
| 500 | 6 | Server internal error |

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| "no API credentials found" | `xparser auth` or set `XPARSER_APP_ID` + `XPARSER_SECRET_CODE` |
| Timeout on large files | `--timeout 1200` |
| Poor extraction quality | Try different `--parse-mode`, e.g. `vlm` |
| Tables look wrong | `--table-flavor md` for Markdown, or `--table-flavor html` |
| Need PDF password | `--pdf-pwd your_password` |
| Page out of range | Adjust `--page-start` / `--page-count` |
| Multiple formats to stdout | Add `-o <dir>` |
| Private deployment | `--base-url https://your-server.com` |
| Batch partially failed | Check stderr for per-file status; succeeded files are saved |

## Notes

- All status/progress messages go to stderr; only document content goes to stdout
- `lite` and `vlm` modes return `elements` array instead of `detail`/`pages` structure
- Batch mode reports progress on stderr as `[n/total]`
- Credentials are stored in `~/.xparser/config.yaml` after `xparser auth`
- Image IDs from `--get-image` are valid for 30 days
