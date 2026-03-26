---
name: xparser
description: Textin xParser document parsing CLI that converts PDFs, images, Office documents, and 20+ file formats into Markdown and structured JSON via the Textin xParser API. Supports multiple parse modes (auto, scan, lite, parse, vlm), table/formula/chart recognition, image extraction, batch processing, and piped workflows. Use this skill when extracting text from PDFs, converting documents to Markdown, performing OCR, parsing Word/Excel/PowerPoint files, extracting tables or formulas, extracting images from documents, batch processing, or troubleshooting document parsing errors.
metadata: {"openclaw":{"emoji":"📄","requires":{"bins":["xparser"]},"install":[{"id":"install-unix","kind":"download","os":["darwin","linux"],"bins":["xparser"],"url":"https://dllf.intsig.net/download/2026/Solution/xparser","label":"Install xparser CLI (Linux/macOS)"}]}}
---

# xParser CLI Skill

xParser turns documents into Markdown. It wraps the [Textin xParser API](https://www.textin.com) as a single binary with zero dependencies.

## Quick reference

```bash
./xparser parse report.pdf                          # Markdown → stdout
./xparser parse report.pdf -o ./out/                # Save to directory
./xparser parse report.pdf --parse-mode vlm         # VLM mode
./xparser parse *.pdf -o ./results/                 # Batch convert
./xparser parse https://example.com/doc.pdf         # Parse a URL
cat doc.pdf | ./xparser parse --stdin -o result.md  # Pipe in
```

## Finding xparser

Before running any command, the agent MUST locate the xparser binary. Do NOT assume it is in PATH or the current directory.

**Search order** (use the first one found):

```bash
# 1. PATH
which xparser 2>/dev/null

# 2. Current directory
test -x ./xparser && echo ./xparser

# 3. Home directory (common install location)
test -x ~/xparser && echo ~/xparser

# 4. ~/.local/bin
test -x ~/.local/bin/xparser && echo ~/.local/bin/xparser

# 5. Broad search (fallback)
find ~ /usr/local/bin /opt -name xparser -type f 2>/dev/null | head -1
```

Once found, use the **full path** in all subsequent commands (e.g. `/home/user/xparser parse ...`). If not found, install it.

## Install (Linux / macOS)

```bash
curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparser -o ~/xparser && chmod +x ~/xparser
~/xparser version   # verify
```

Recommended: add to PATH for convenience:
```bash
sudo mv ~/xparser /usr/local/bin/xparser
# or
echo 'export PATH="$HOME:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

## Authentication

Every parse call needs Textin API credentials. Get them at <https://www.textin.com/console/dashboard/setting>.

```bash
./xparser auth                          # interactive setup → saves to ~/.xparser/config.yaml
./xparser auth --show                   # show current credentials (masked)
```

**Resolution order** (highest wins):

| Priority | Method | Example |
|----------|--------|---------|
| 1 | Flags | `--app-id ID --secret-code CODE` |
| 2 | Env vars | `XPARSER_APP_ID` / `XPARSER_SECRET_CODE` |
| 3 | Config file | `~/.xparser/config.yaml` |

For private deployments, set `--base-url https://your-server.com` or `./xparser config set base_url https://your-server.com`.

---

## Choosing a parse mode

默认为 `scan` 模式。

| Mode | Default | Description |
|------|---------|-------------|
| `auto` | | 由引擎自动选择，适用范围最广 |
| `scan` | **Yes** | 文档统一当成图片解析 |
| `lite` | | 轻量版，只输出表格和文字结果 |
| `parse` | | 仅电子档文字解析，速度最快 |
| `vlm` | | 视觉语言模型解析模式 |

### When to suggest a specific mode

- User says nothing special → omit `--parse-mode` (defaults to `scan`)
- User mentions "视觉语言模型", "VLM" → `--parse-mode vlm`
- User wants speed, electronic PDF only, "电子档" → `--parse-mode parse`
- User wants tables + text only, lightweight, "轻量" → `--parse-mode lite`
- User is unsure, mixed document types, "自动" → `--parse-mode auto`

---

## parse — The main command

### Basic usage

```bash
./xparser parse <file-or-url> [flags]
./xparser parse report.pdf                              # Markdown to stdout
./xparser parse report.pdf -o report.md                 # Save to file
./xparser parse report.pdf -o ./out/                    # Save to directory
./xparser parse report.pdf -f json                      # Full JSON output
./xparser parse report.pdf -f md,json -o ./out/         # Both formats (needs -o dir)
```

### All parse flags

**Output:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | _(stdout)_ | 输出路径（文件或目录） |
| `-f, --format` | `md` | 输出格式：`md`、`json`，逗号分隔可同时输出 |

**Parse mode & pages:**

| Flag | Default | Description |
|------|---------|-------------|
| `--parse-mode` | `scan` | 文档的解析模式，默认为scan模式。`auto` 由引擎自动选择，适用范围最广；`scan` 文档统一当成图片解析；`lite` 轻量版，只输出表格和文字结果；`parse` 仅电子档文字解析，速度最快；`vlm` 视觉语言模型解析模式 |
| `--pdf-pwd` | | 当pdf为加密文档时，需要提供密码 |
| `--page-start` | `0` | 当上传的是pdf时，表示从第几页开始解析，不传该参数时默认从首页开始 |
| `--page-count` | `1000` | 当上传的是pdf时，表示要进行转换的pdf页数，总页数不得超过1000页 |
| `--dpi` | `144` | pdf文档的坐标基准，默认144dpi。可选值：`72`、`144`、`216`。与parse_mode参数联动：当parse_mode=auto时默认动态；当parse_mode=scan时默认144 |

**Tables & structure:**

| Flag | Default | Description |
|------|---------|-------------|
| `--table-flavor` | `html` | markdown里的表格格式，默认为html。`md` 按md语法输出表格；`html` 按html语法输出表格；`none` 不进行表格识别，把表格图像当成普通文字段落来识别 |
| `--apply-document-tree` | `1` | markdown中是否生成标题层级，默认为1生成标题。`0` 不生成标题，同时也不会返回catalog字段；`1` 生成标题 |
| `--paratext-mode` | `annotation` | markdown中非正文文本内容展示模式。非正文内容包括页眉页脚，子图中的文本。`none` 不展示；`annotation` 以注释格式插入到markdown中，页眉页脚中的图片只保留文本；`body` 以正文格式插入到markdown中 |
| `--apply-merge` | `1` | 是否进行段落合并和表格合并。`0` 不合并；`1` 合并 |

**Images:**

| Flag | Default | Description |
|------|---------|-------------|
| `--get-image` | `none` | 获取markdown里的图片，默认为none不返回任何图像。`none` 不返回任何图像；`page` 返回每一页的整页图像（pdf页的完整页图片）；`objects` 返回页面内的子图像（pdf页内的各个子图片）；`both` 返回整页图像和图像对象 |
| `--image-output-type` | `default` | 指定引擎返回的图片对象输出类型。`default` 子图片对象为图片url，页图片对象为图片id；`base64str` 所有图片对象为base64字符串，适用于没有云存储的用户，但返回体积会很大。page_count超过1000页时不支持base64返回 |

**Recognition:**

| Flag | Default | Description |
|------|---------|-------------|
| `--formula-level` | `0` | 公式识别等级，默认为0全识别，开启公式识别后会使用latex表达式。`0` 全识别；`1` 仅识别行间公式，行内公式不识别；`2` 不识别 |
| `--apply-chart` | `0` | 是否开启图表识别，开启后会将识别到的图表以表格形式输出。`0` 不开启；`1` 开启 |
| `--underline-level` | `0` | 控制下划线识别范围，默认为0不识别。`0` 不识别；`1` 仅识别无文字的下划线（仅scan模式可用）；`2` 识别全部的下划线（仅scan模式可用） |
| `--apply-image-analysis` | `0` | 利用大模型对文档中的子图进行分析，分析结果以markdown格式输出并替换掉子图的文本识别内容。`0` 不进行图像分析；`1` 进行图像分析 |
| `--crop-dewarp` | `0` | 是否进行切边矫正处理。`0` 不进行切边矫正；`1` 进行切边矫正 |
| `--remove-watermark` | `0` | 是否进行去水印处理。`0` 不去水印；`1` 去水印 |

**Output detail:**

| Flag | Default | Description |
|------|---------|-------------|
| `--markdown-details` | `1` | 是否返回结果中的detail字段，保存markdown各类型元素的详细信息。`0` 不生成；`1` 生成 |
| `--page-details` | `1` | 是否返回结果中的pages字段，保存每一页更加详细的解析结果。`0` 不返回；`1` 返回 |
| `--raw-ocr` | `0` | 是否返回全部文字识别结果（包含字符坐标信息），结果字段为raw_ocr。与page_details参数联动，当page_details为0时不返回。`0` 不返回；`1` 返回 |
| `--char-details` | `0` | 是否返回结果中的char_pos字段（保存每个字符的位置信息）和raw_ocr中的char_相关字段。`0` 不返回；`1` 返回 |
| `--catalog-details` | `0` | 是否返回结果中的catalog字段，保存目录相关信息。与apply_document_tree参数联动，当apply_document_tree为0时不返回。`0` 不返回；`1` 返回 |
| `--get-excel` | `0` | 是否返回excel的base64结果，结果字段为excel_base64，可根据该字段进行后处理保存excel文件。`0` 不返回；`1` 返回 |

**Input modes:**

| Flag | Default | Description |
|------|---------|-------------|
| `--list` | | 从文件读取输入列表（每行一个路径） |
| `--stdin-list` | `false` | 从stdin读取文件列表 |
| `--stdin` | `false` | 从stdin读取文件内容 |
| `--stdin-name` | `stdin.pdf` | stdin模式的文件名提示 |
| `--timeout` | `600` | 超时秒数 |

---

## Recipes

### Batch processing

```bash
# All PDFs in current directory
./xparser parse *.pdf -o ./results/

# From a file list
./xparser parse --list files.txt -o ./results/

# From find/pipe
find ./docs -name "*.pdf" | ./xparser parse --stdin-list -o ./results/
```

Batch mode always requires `-o <directory>`. Progress is reported on stderr as `[n/total]`.

### Image extraction

Two steps: parse with `--get-image` to get JSON result, then use `--from` to batch download.

```bash
# Step 1: parse with image extraction, save JSON result
./xparser parse report.pdf --get-image objects -f json -o ./out/

# Step 2: download all images directly from the JSON result
./xparser download --from ./out/report.json -o ./out/images/
```

`--from` automatically extracts image IDs from `metrics[].image_id`、`result.pages[].image_id`、`result.detail[].image_url`，去重后逐个下载，无需手动提取 ID。

也可以手动指定 ID 下载：
```bash
./xparser download <image_id_1> <image_id_2> -o ./out/images/
```

For inline base64 images (larger response, but no second step):
```bash
./xparser parse report.pdf --get-image objects --image-output-type base64str -o ./out/
```

Image IDs are valid for **30 days** after parsing.

下载接口返回 JSON 格式 `{"code": 200, "data": {"image": "<base64>"}}`，CLI 会自动解析并将 base64 转为图片文件。

### Tables

Tables default to HTML. For Markdown tables:
```bash
./xparser parse report.pdf --table-flavor md
```

Use `--table-flavor none` to skip table recognition entirely.

### Formulas and charts

```bash
./xparser parse paper.pdf --formula-level 0              # All formulas (default)
./xparser parse paper.pdf --formula-level 1              # Display formulas only
./xparser parse paper.pdf --apply-chart 1                # Charts → tables
```

### PDF page ranges

```bash
./xparser parse book.pdf --page-start 10 --page-count 10   # Pages 10–19
./xparser parse secret.pdf --pdf-pwd mypassword             # Encrypted PDF
```

### Piping into other tools

xParser outputs Markdown to stdout by default (status goes to stderr), so it pipes cleanly:

```bash
# Parse and search
./xparser parse report.pdf | grep "revenue"

# Parse and feed to another model
./xparser parse paper.pdf --parse-mode vlm | llm "summarize this paper"

# Parse from stdin
cat report.pdf | ./xparser parse --stdin
```

### JSON output for programmatic use

```bash
# Full structured JSON
./xparser parse report.pdf -f json

# Minimal JSON (no detail/pages arrays — smaller payload)
./xparser parse report.pdf -f json --markdown-details 0 --page-details 0
```

**Note:** `lite` and `vlm` modes return an `elements` array instead of the `detail`/`pages` structure used by other modes.

---

## Other commands

### download — Fetch images

```bash
./xparser download --from result.json -o ./images/  # From parse result JSON (recommended)
./xparser download <image_id> -o ./images/           # By image ID
./xparser download id1 id2 id3 -o ./images/          # Batch by IDs
./xparser download --from result.json extra_id -o .  # Mixed: JSON + manual IDs
```

### config — Manage settings

```bash
./xparser config show                              # Display all settings
./xparser config set base_url https://my-server    # Set a value
./xparser config set app_id YOUR_ID                # Set app ID
./xparser config reset                             # Reset to defaults
./xparser config path                              # Show config file location
```

Keys: `app_id`, `secret_code`, `base_url`.

### update — Self-update

```bash
./xparser update              # Download latest version
./xparser update --check      # Check only, don't install
```

### version

```bash
./xparser version             # Show version, commit, Go version, OS/arch
```

---

## Supported formats

| Category | Extensions |
|----------|-----------|
| PDF | `.pdf` |
| Images | `.png`, `.jpg`, `.jpeg`, `.bmp`, `.tiff`, `.webp` |
| Word | `.doc`, `.docx` |
| Excel | `.xls`, `.xlsx`, `.csv` |
| PowerPoint | `.ppt`, `.pptx` |
| Web | `.html`, `.mhtml` |
| Text | `.txt`, `.rtf` |
| Other | `.ofd` |
| URL | `http://…`, `https://…` |

**Limits:** files ≤ 500 MB, PDFs ≤ 1000 pages, Excel ≤ 2000 rows / 100 cols per sheet, TXT ≤ 100 KB, images 20–20000 px (aspect ratio < 2) or 20–10000 px (others).

---

## Output behavior

| Scenario | Behavior |
|----------|----------|
| No `-o` | Content → stdout, progress → stderr |
| `-o file.md` | Content → file, progress → stderr |
| `-o ./dir/` | Content → dir, progress → stderr |
| `-f md,json` | Requires `-o <directory>` |
| `--get-excel 1` | Excel file saved alongside Markdown |
| Batch mode | Requires `-o <directory>` |

---

## Default output directory

When the agent saves output on behalf of the user and no `-o` is specified, generate:

```
~/xParser-Skill/<name>_<hash>/
```

- `<name>`: filename without extension (sanitize spaces and special chars to `_`)
- `<hash>`: first 6 chars of `echo -n "<full_source_path>" | md5sum | cut -c1-6`

| Source | Output directory |
|--------|-----------------|
| `report.pdf` | `~/xParser-Skill/report_f1a2b3/` |
| `./docs/年报2024.pdf` | `~/xParser-Skill/年报2024_c7e9d4/` |
| `https://example.com/q1.pdf` | `~/xParser-Skill/q1_a3f2b1/` |

If the user specifies `-o`, use their path as-is.

---

## Exit codes

| Code | Meaning | What to do |
|------|---------|------------|
| 0 | Success | — |
| 1 | General error | Check network; retry; add `-v` for HTTP debug |
| 2 | Bad parameters | Check flag names and values |
| 3 | Auth failure | Run `./xparser auth` to configure credentials |
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

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| "no API credentials found" | `./xparser auth` or set `XPARSER_APP_ID` + `XPARSER_SECRET_CODE` |
| Timeout on large files | `--timeout 1200` |
| Poor extraction quality | Try different `--parse-mode`, e.g. `vlm` |
| Tables look wrong | `--table-flavor md` for Markdown, or `--table-flavor html` |
| Need PDF password | `--pdf-pwd your_password` |
| Page out of range | Adjust `--page-start` / `--page-count` |
| Multiple formats to stdout | Add `-o <dir>` |
| Private deployment | `--base-url https://your-server.com` |
| Batch partially failed | Check stderr for per-file status; succeeded files are saved |

---

## Agent guidelines

### File path quoting

Always double-quote paths with spaces or special characters:
```bash
./xparser parse "report 01.pdf"          # correct
./xparser parse report 01.pdf            # wrong — treated as two args
```

### Post-parse hints

After a successful parse, give **one** short tip (don't repeat in the same session):

- If no `--get-image`: "Add `--get-image objects` to also extract images."
- If `--table-flavor` not set: "Tables default to HTML; use `--table-flavor md` for Markdown."
- If `scan` mode on a simple electronic PDF: "`--parse-mode parse` would be faster for this type of document."

### Error handling

On failure, explain the exit code and suggest a fix — don't blindly re-run the command. Use `-v` to show HTTP details if the cause is unclear.
