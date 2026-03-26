# xParser Ecosystem

xParser Ecosystem 是基于 [Textin xParser API](https://api.textin.com) 的开源文档解析工具集，包含 CLI 命令行工具和 AI Skill 集成，支持将 20+ 种文件格式转换为 Markdown / JSON，零依赖、开箱即用。

## 项目结构

```
xparser-ecosystem/
├── cli/                # Go CLI 应用
│   ├── main.go
│   ├── cmd/            # 命令实现（parse, auth, download, config, update, version）
│   └── internal/       # 内部包（config, exitcode, output）
├── skills/             # Claude AI Skill 集成
│   ├── SKILL.md
│   ├── CONTRIBUTING.md
│   └── _meta.json
└── README.md
```

## 功能特性

- **20+ 格式支持** — PDF、DOC/DOCX、XLS/XLSX、PPT/PPTX、PNG/JPG、HTML、TXT、OFD 等
- **多种解析模式** — `auto` / `scan`（扫描件）/ `lite`（轻量）/ `parse`（电子 PDF）/ `vlm`（视觉大模型）
- **灵活输出** — Markdown（默认）、JSON，或同时输出两种格式
- **批量处理** — 支持 glob、文件列表、stdin 管道输入，带进度报告
- **图片提取** — 页面图片 / 内嵌对象图片提取与批量下载
- **表格 & 公式** — 表格输出为 HTML/Markdown，公式输出为 LaTeX
- **图表识别 & 文档结构树** — 可选开启
- **私有部署** — 支持自定义 `base_url`
- **Unix 友好** — 内容输出到 stdout，状态信息到 stderr，完美支持管道

## 快速开始

### 安装

从 [Releases](https://github.com/proluzhiqin/xparser-ecosystem/releases) 下载对应平台的二进制文件，或从源码构建：

```bash
cd cli
go build -o xparser -ldflags "\
  -X github.com/textin/xparser-ecosystem/cli/cmd.version=0.1.0 \
  -X github.com/textin/xparser-ecosystem/cli/cmd.commit=$(git rev-parse --short HEAD) \
  -X github.com/textin/xparser-ecosystem/cli/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" .
```

### 配置认证

```bash
# 交互式配置
xparser auth

# 或通过环境变量
export XPARSER_APP_ID=your_app_id
export XPARSER_SECRET_CODE=your_secret_code
```

凭证优先级：CLI 参数 > 环境变量 > 配置文件（`~/.xparser/config.yaml`）

### 解析文档

```bash
# 单文件解析，输出到 stdout
xparser parse report.pdf

# 指定输出文件
xparser parse report.pdf -o result.md

# 输出 JSON 格式
xparser parse report.pdf -f json -o result.json

# 同时输出 Markdown 和 JSON
xparser parse report.pdf -f md,json -o ./output/

# 指定解析模式
xparser parse scanned.pdf --parse-mode scan
xparser parse digital.pdf --parse-mode parse

# 解析远程 URL
xparser parse https://example.com/doc.pdf -o result.md
```

### 批量处理

```bash
# glob 模式
xparser parse *.pdf -o ./results/

# 文件列表
xparser parse --list files.txt -o ./results/

# stdin 管道
find . -name "*.pdf" | xparser parse --stdin-list -o ./results/
```

### 图片提取与下载

```bash
# 解析时提取图片
xparser parse report.pdf --get-image both -f json -o result.json

# 从解析结果批量下载图片
xparser download --from result.json -o ./images/
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `xparser parse` | 文档解析（核心命令） |
| `xparser auth` | 配置 API 凭证 |
| `xparser download` | 批量下载解析结果中的图片 |
| `xparser config` | 查看/设置/重置配置 |
| `xparser update` | 自更新到最新版本 |
| `xparser version` | 显示版本信息 |

## 常用参数

| 参数 | 说明 |
|------|------|
| `-f, --format` | 输出格式：`md`（默认）、`json`、`md,json` |
| `-o, --output` | 输出文件或目录 |
| `--parse-mode` | 解析模式：`auto`/`scan`/`lite`/`parse`/`vlm` |
| `--table-flavor` | 表格格式：`html`/`md`/`none` |
| `--get-image` | 图片提取：`none`/`page`/`objects`/`both` |
| `--formula-level` | 公式识别：`0`=全部 / `1`=仅行间 / `2`=关闭 |
| `--apply-chart` | 图表识别：`0`/`1` |
| `--apply-document-tree` | 文档结构树：`0`/`1` |
| `--page-start, --page-count` | PDF 页面范围 |
| `--dpi` | PDF 渲染 DPI：`72`/`144`/`216` |
| `-v, --verbose` | 调试模式，打印 HTTP 请求详情 |

## 退出码

| 码 | 含义 | 建议 |
|----|------|------|
| 0 | 成功 | — |
| 1 | 一般错误 | 使用 `-v` 查看详情 |
| 2 | 参数错误 | 检查参数名称和值 |
| 3 | 认证错误 | 运行 `xparser auth` |
| 4 | 文件错误 | 检查文件类型、大小、格式 |
| 5 | 解析失败 | 尝试其他 `--parse-mode` |
| 6 | 服务端错误 | 稍后重试 |
| 7 | 余额不足 | 前往 textin.com 充值 |

## 文件限制

- 文件大小 ≤ 500 MB
- PDF 页数 ≤ 1000
- Excel ≤ 2000 行 × 100 列/sheet
- TXT ≤ 100 KB
- 图片尺寸 20–20000 px

## 技术栈

- **语言**：Go 1.18+
- **CLI 框架**：[Cobra](https://github.com/spf13/cobra)
- **配置**：YAML（gopkg.in/yaml.v3）
- **HTTP**：Go 标准库，零外部依赖

## 贡献

请参阅 [skills/CONTRIBUTING.md](skills/CONTRIBUTING.md) 了解贡献指南。

## License

[AGPL-3.0](LICENSE)
