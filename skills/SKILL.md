---
name: xparse-cli
description: Textin xparse document parsing CLI that converts PDFs, images, Office documents and 20+ file formats into Markdown and structured JSON via the Textin xParse API. Supports OCR, table extraction, formula recognition, and image extraction. Zero config, free API default, structured errors, stdout-friendly.
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

### Verify

```bash
xparse-cli version
```

## Quick start

Zero config — defaults to free API, no registration needed.

```bash
xparse-cli parse report.pdf                        # Markdown to stdout
xparse-cli parse report.pdf --view json             # JSON view
xparse-cli parse report.pdf -o ./result/            # Save to directory
xparse-cli parse report.pdf --page-range "1-5"      # Specific pages
xparse-cli parse secret.pdf --password mypassword   # Encrypted PDF
xparse-cli parse --list files.txt -o ./result/      # Batch mode
```

## Authentication (paid API)

```bash
xparse-cli auth                                     # Interactive setup
# or
export XPARSE_APP_ID=your_app_id
export XPARSE_SECRET_CODE=your_secret_code
xparse-cli parse report.pdf --api paid
```

API selection: no `--api` → paid if credentials exist, else free. `--api free` → force free. `--api paid` → force paid.

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--view` | | `markdown` | Output view: `markdown`, `json` |
| `--api` | | _(auto)_ | API mode: `free`, `paid` |
| `--page-range` | | | Page range: `"1-5"` or `"1-2,5-10"` |
| `--password` | | | Password for encrypted documents |
| `--include-char-details` | | `false` | Include character-level coordinates and confidence |
| `--list` | | | Read input list from file; requires `--output` |
| `--output` | `-o` | _(stdout)_ | Output file path or directory |
| `--verbose` | `-v` | `false` | Print HTTP request details for debugging |

## Supported formats

PDF, JPG, JPEG, PNG, BMP, TIFF, GIF, WebP, DOC, DOCX, XLS, XLSX, CSV, PPT, PPTX, HTML, MHTML, TXT, RTF, OFD, URL (`http://…`)

## Other commands

```bash
xparse-cli download --from result.json -o ./images/ # Download element images
xparse-cli config show                               # Show config
xparse-cli update                                    # Self-update
```

---

## How Agent captures and handles errors

CLI output contract:
- **stdout** → document content only (markdown or json)
- **stderr** → error messages only
- **$?** → exit code: 0, 1, 2, or 3

### Capture pattern

```bash
RESULT=$(xparse-cli parse "$FILE" 2>stderr.tmp)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  # Success — $RESULT contains markdown/json
  echo "$RESULT"
else
  ERROR=$(cat stderr.tmp)
  # Route by exit code — see decision table below
fi
rm -f stderr.tmp
```

### Decision table

```
EXIT_CODE=0  →  success, use stdout content
EXIT_CODE=1  →  match stderr against Exit 1 table → retry or escalate
EXIT_CODE=2  →  match stderr against Exit 2 table → fix command, re-run
EXIT_CODE=3  →  parse "api_code：message" from stderr → match Exit 3 table → retry / fix / escalate
```

---

## Exit 1 — general failures (stderr = plain text)

Agent matches the **exact stderr string** to decide action.

| stderr | category | action | retryable |
|--------|----------|--------|-----------|
| `credentials configuration error` | auth | Ask human: run `xparse-cli auth` or set env vars | no |
| `network or request failed` | technical | Retry with `--verbose`; max 2 retries, 2s backoff | yes |
| `API returned success but no result data` | technical | Retry once; escalate if persists | yes |
| `failed to create output directory` | filesystem | Ask human: check directory permissions | no |
| `batch completed with errors` | technical | Some files failed — check per-file exit 3 errors | no |
| `failed to save result` | filesystem | Ask human: check disk space / write permissions | no |

## Exit 2 — command errors (stderr = plain text)

Agent **always** fixes the command and re-runs. Never retry the same command.

| stderr | fix |
|--------|-----|
| `invalid --view value, must be 'markdown' or 'json'` | Change `--view` to `markdown` or `json` |
| `invalid --api value, must be 'free' or 'paid'` | Change `--api` to `free` or `paid` |
| `failed to open or read --list file` | Fix `--list` file path |
| `no input files specified` | Add file argument or `--list` |
| `flag value is not a file` | Use `--include-char-details` without a value (it's a bool flag) |
| `file not found` | Fix file path |
| `--list requires --output` | Add `-o <directory>` |
| `multiple inputs require --output` | Add `-o <directory>` |
| `paid API requires credentials` | Run `xparse-cli auth` or set `XPARSE_APP_ID` + `XPARSE_SECRET_CODE` |

## Exit 3 — API errors (stderr = `api_code：message`)

Agent parses stderr as `api_code：message`, then matches `api_code` to decide action.

### Retryable — Agent retries automatically (max 2 times, 2s backoff)

| api_code | stderr message | action |
|----------|----------------|--------|
| 500 | `500：服务器内部错误` | Retry |
| 30203 | `30203：基础服务故障，请稍后重试` | Retry |
| 40306 | `40306：qps超过限制` | Wait 3s, then retry |
| 40428 | `40428：word和ppt转pdf失败或者超时` | Retry; or convert to PDF first, then parse PDF |
| 50207 | `50207：部分页面解析失败` | Use partial results if available; retry for full |

### Auth — escalate to human

| api_code | stderr message | action |
|----------|----------------|--------|
| 40101 | `40101：x-ti-app-id 或 x-ti-secret-code 为空` | Ask human: `xparse-cli auth` |
| 40102 | `40102：x-ti-app-id 或 x-ti-secret-code 无效，验证失败` | Ask human: check credentials |
| 40103 | `40103：客户端IP不在白名单` | Ask human: check IP whitelist |

### Quota / limit — escalate to human

| api_code | stderr message | action |
|----------|----------------|--------|
| 40003 | `40003：余额不足，请充值后再使用` | Ask human to top up, or switch `--api free` |
| 40302 | `40302：上传文件大小不符，文件大小不超过 500M` | Split or compress file |
| 40304 | `40304：图片尺寸不符` | Resize image (20–20000px) |
| 40307 | `40307：今日免费额度已用完` | Ask human: try tomorrow or `--api paid` |

### Parameter — Agent fixes command and re-runs

| api_code | stderr message | fix |
|----------|----------------|-----|
| 400 | `400：上传的文件不能为空` | File is empty (0 bytes) — use a different file |
| 40004 | `40004：参数错误，请查看技术文档，检查传参` | Check flags and values |
| 40007 | `40007：机器人不存在或未发布` | Check configuration |
| 40008 | `40008：机器人未开通，请至市场开通后重试` | Activate service at textin.com |
| 40301 | `40301：图片类型不支持` | Use supported format: JPG/PNG/BMP/TIFF/WebP |
| 40303 | `40303：文件类型不支持` | Use supported format (see list above) |
| 40305 | `40305：识别文件未上传` | Check file argument |
| 40400 | `40400：无效的请求链接，请检查链接是否正确` | Fix URL |
| 40422 | `40422：文件损坏` | Re-download or use different copy |
| 40423 | `40423：PDF密码错误` | Add `--password <correct_password>` |
| 40424 | `40424：页数设置超出文件范围` | Adjust `--page-range` to valid range |
| 40425 | `40425：文件格式不支持` | Use supported format |
| 40427 | `40427：DPI参数不在支持列表中` | Use DPI: 72, 144, or 216 |
| 40429 | `40429：不支持的引擎` | Contact support |

### Unknown api_code

If `api_code` is not in the tables above, retry once. If still fails, escalate to human with the full stderr output.

---

## Agent rules

- **stdout is sacred** — never mix errors into stdout; always capture stderr separately
- **Quote file paths** with spaces: `xparse-cli parse "report 01.pdf"`
- On exit 1 with `network or request failed`, add `--verbose` on retry to capture HTTP details
- Never retry exit 2 with the same command — always fix first
- On exit 3 retryable errors, max 2 retries with 2s backoff
- When user asks to **upgrade**: `curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh | sh`

### Default output directory

When Agent saves output and user didn't specify `--output`:

```
~/xparse-cli/<name>_<hash>/
```

- `<name>`: filename without extension, sanitized (`_` for spaces/special chars)
- `<hash>`: `echo -n "<full_path>" | md5sum | cut -c1-6`

When user specifies `--output`, use their path as-is.

## Notes

- stdout = document content only; stderr = errors only
- Free API: zero config, no registration
- Paid API: `XPARSE_APP_ID` + `XPARSE_SECRET_CODE`
- Credential priority: flags > env vars > `~/.xparse-cli/config.yaml`
