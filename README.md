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
- **Agent 友好** — 零配置免费 API、结构化错误输出、stdout/stderr 分离
- **灵活视图** — Markdown（默认）或 JSON，通过 `--view` 切换
- **批量处理** — 支持文件列表，带进度报告
- **图片提取** — 内嵌对象图片提取与批量下载
- **表格 & 公式** — 表格输出为 HTML，公式输出为 LaTeX
- **私有部署** — 支持自定义 `base_url`
- **Unix 友好** — 内容输出到 stdout，状态信息到 stderr，完美支持管道

## 快速开始

### 安装

**Linux / macOS**

```bash
curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://dllf.intsig.net/download/2026/Solution/xparse-cli/install.ps1 | iex
```

从源码构建请参考 [cli/README.md](cli/README.md#从源码构建)。

### 零配置解析（免费 API）

无需注册、无需配置，直接使用：

```bash
# Markdown 到 stdout
xparse-cli parse report.pdf

# JSON 视图
xparse-cli parse report.pdf --view json

# 保存到目录
xparse-cli parse report.pdf --output ./result/

# 指定页码范围
xparse-cli parse report.pdf --page-range "1-5"

# 加密 PDF
xparse-cli parse secret.pdf --password mypassword
```

### 付费 API（更高配额）

前往 [Textin 控制台](https://www.textin.com/console/dashboard/setting) 获取 `x-ti-app-id` 和 `x-ti-secret-code`，然后：

```bash
# 交互式配置（推荐）
xparse-cli auth

# 或通过环境变量
export XPARSE_APP_ID=your_x-ti-app-id
export XPARSE_SECRET_CODE=your_x-ti-secret-code

# 使用付费 API
xparse-cli parse report.pdf --api paid
```

| 环境变量 | 对应 Textin 凭证 |
|----------|------------------|
| `XPARSE_APP_ID` | `x-ti-app-id` |
| `XPARSE_SECRET_CODE` | `x-ti-secret-code` |

凭证优先级：CLI 参数 > 环境变量 > 配置文件（`~/.xparse-cli/config.yaml`）

### 批量处理

```bash
# 文件列表
xparse-cli parse --list files.txt --output ./results/
```

### 图片下载

```bash
# 从解析结果批量下载图片
xparse-cli download --from result.json -o ./images/
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `xparse-cli parse` | 文档解析（核心命令） |
| `xparse-cli auth` | 配置 API 凭证 |
| `xparse-cli download` | 批量下载解析结果中的图片 |
| `xparse-cli config` | 查看/设置/重置配置 |
| `xparse-cli update` | 自更新到最新版本 |
| `xparse-cli version` | 显示版本信息 |

## parse 命令参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--view` | `markdown` | 输出视图：`markdown`、`json` |
| `--api` | _(auto)_ | API 模式：`free`、`paid` |
| `--page-range` | | 页码范围：`"1-5"` 或 `"1-2,5-10"` |
| `--password` | | 加密文档密码 |
| `--include-char-details` | `false` | 返回字符级坐标和置信度 |
| `--list` | | 从文件读取输入列表（需配合 `--output`） |
| `-o, --output` | _(stdout)_ | 输出文件路径或目录 |
| `-v, --verbose` | `false` | 调试模式，打印 HTTP 请求详情 |

## 退出码

| 码 | 含义 | 建议 |
|----|------|------|
| 0 | 成功 | — |
| 1 | 一般错误 / 网络异常 | 使用 `--verbose` 查看详情；重试 |
| 2 | 参数错误 | 检查命令语法和参数值 |
| 3 | API 返回错误 | 解析 stderr 中的 JSON 错误信息 |

## 文件限制

- 文件大小 ≤ 500 MB
- PDF 页数 ≤ 1000
- Excel ≤ 2000 行 × 100 列/sheet
- TXT ≤ 100 KB
- 图片尺寸 20–20000 px

## 技术栈

- **语言**：Go 1.23+
- **CLI 框架**：[Cobra](https://github.com/spf13/cobra)
- **配置**：YAML（gopkg.in/yaml.v3）
- **HTTP**：Go 标准库，零外部依赖

## 贡献

请参阅 [skills/CONTRIBUTING.md](skills/CONTRIBUTING.md) 了解贡献指南。

## License

[AGPL-3.0](LICENSE)
