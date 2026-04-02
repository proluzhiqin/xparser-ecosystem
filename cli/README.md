# xParser CLI

Textin xParser 文档解析命令行工具，基于 [Textin xParser API](https://www.textin.com) 实现。

支持将 PDF、图片、Office 文档等 20+ 格式转换为 Markdown 及结构化数据。

## 安装

### 一键安装

**Linux / macOS**

```bash
curl -fsSL https://dllf.intsig.net/download/2026/Solution/xparser/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://dllf.intsig.net/download/2026/Solution/xparser/install.ps1 | iex
```

### 从源码构建

> 要求 Go 1.23+

**单平台构建（当前系统）：**

```bash
cd cli
go build -o xparser .
```

**带版本信息构建：**

```bash
VERSION=v0.0.1
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PKG=github.com/textin/xparser-ecosystem/cli/cmd

go build -ldflags "-s -w \
  -X ${PKG}.version=${VERSION} \
  -X ${PKG}.commit=${COMMIT} \
  -X ${PKG}.date=${DATE}" \
  -o xparser .
```

**交叉编译全平台：**

```bash
VERSION=v0.0.1
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PKG=github.com/textin/xparser-ecosystem/cli/cmd
LDFLAGS="-s -w -X ${PKG}.version=${VERSION} -X ${PKG}.commit=${COMMIT} -X ${PKG}.date=${DATE}"

GOOS=linux   GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/xparser-linux-amd64 .
GOOS=linux   GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/xparser-linux-arm64 .
GOOS=darwin  GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/xparser-darwin-amd64 .
GOOS=darwin  GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/xparser-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/xparser-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/xparser-windows-arm64.exe .
```

产物位于 `cli/dist/` 目录，共 6 个二进制文件：

| 平台 | 文件 |
|------|------|
| Linux x86_64 | `xparser-linux-amd64` |
| Linux ARM64 | `xparser-linux-arm64` |
| macOS Intel | `xparser-darwin-amd64` |
| macOS Apple Silicon | `xparser-darwin-arm64` |
| Windows x86_64 | `xparser-windows-amd64.exe` |
| Windows ARM64 | `xparser-windows-arm64.exe` |

## 快速开始

### 1. 配置凭证

前往 [Textin 控制台](https://www.textin.com/console/dashboard/setting) 获取 `x-ti-app-id` 和 `x-ti-secret-code`，然后运行：

```bash
./xparser auth
```

按提示输入 App ID 和 Secret Code，凭证将保存至 `~/.xparser/config.yaml`。

### 2. 解析文档

```bash
# 输出 Markdown 到终端
./xparser parse report.pdf

# 保存到文件
./xparser parse report.pdf -o result.md

# 保存到目录
./xparser parse report.pdf -o ./output/
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `./xparser parse` | 解析文档，输出 Markdown / JSON |
| `./xparser auth` | 配置 API 凭证（交互式） |
| `./xparser config` | 管理配置（show / set / reset / path） |
| `./xparser download` | 下载解析结果中的图片（通过 image_id） |
| `./xparser update` | 自更新 CLI 到最新版本 |
| `./xparser version` | 显示版本信息 |

## 使用示例

### 基本解析

```bash
# 解析 PDF 到 stdout
./xparser parse report.pdf

# 保存为 Markdown 文件
./xparser parse report.pdf -o report.md

# 保存到指定目录
./xparser parse report.pdf -o ./out/

# 输出完整 JSON（含结构化数据）
./xparser parse report.pdf -f json

# 同时输出 Markdown 和 JSON
./xparser parse report.pdf -f md,json -o ./out/
```

### 解析模式

```bash
# 自动模式（引擎自动选择）
./xparser parse report.pdf --parse-mode auto

# 扫描模式（统一当成图片解析，默认）
./xparser parse report.pdf --parse-mode scan

# 轻量模式（仅表格和文字）
./xparser parse report.pdf --parse-mode lite

# 电子档文字解析（速度最快）
./xparser parse report.pdf --parse-mode parse

# VLM 视觉语言模型
./xparser parse report.pdf --parse-mode vlm
```

### 图片提取

```bash
# 提取页面内子图
./xparser parse report.pdf --get-image objects -o ./out/

# 提取整页图像
./xparser parse report.pdf --get-image page -o ./out/

# 同时提取
./xparser parse report.pdf --get-image both -o ./out/

# 以 base64 格式返回图片
./xparser parse report.pdf --get-image objects --image-output-type base64str -o ./out/
```

### 表格与公式

```bash
# 表格以 Markdown 格式输出（默认 html）
./xparser parse report.pdf --table-flavor md

# 不识别表格
./xparser parse report.pdf --table-flavor none

# 仅识别行间公式
./xparser parse report.pdf --formula-level 1

# 关闭公式识别
./xparser parse report.pdf --formula-level 2

# 开启图表识别
./xparser parse report.pdf --apply-chart 1
```

### PDF 页码控制

```bash
# 从第 5 页开始解析
./xparser parse report.pdf --page-start 5

# 解析前 10 页
./xparser parse report.pdf --page-count 10

# 解析第 10-20 页
./xparser parse report.pdf --page-start 10 --page-count 10

# 加密 PDF
./xparser parse encrypted.pdf --pdf-pwd mypassword
```

### 高级选项

```bash
# 关闭标题层级生成
./xparser parse report.pdf --apply-document-tree 0

# 关闭段落/表格合并
./xparser parse report.pdf --apply-merge 0

# 开启 LLM 图像分析
./xparser parse report.pdf --apply-image-analysis 1

# 切边矫正 + 去水印
./xparser parse report.pdf --crop-dewarp 1 --remove-watermark 1

# 设置 DPI
./xparser parse report.pdf --dpi 216

# 获取 Excel 导出
./xparser parse report.pdf --get-excel 1 -o ./out/

# 下划线识别
./xparser parse report.pdf --underline-level 2
```

### 输出控制

```bash
# 不返回 detail 字段（减小响应体积）
./xparser parse report.pdf --markdown-details 0

# 不返回 pages 字段
./xparser parse report.pdf --page-details 0

# 返回原始 OCR 结果
./xparser parse report.pdf --raw-ocr 1

# 返回字符坐标信息
./xparser parse report.pdf --char-details 1

# 返回目录结构
./xparser parse report.pdf --catalog-details 1
```

### 批量处理

```bash
# 批量解析多个文件
./xparser parse *.pdf -o ./results/

# 从文件列表读取
./xparser parse --list files.txt -o ./results/

# 从 stdin 读取文件列表
find . -name "*.pdf" | ./xparser parse --stdin-list -o ./results/
```

### 其他输入方式

```bash
# 解析在线 URL
./xparser parse https://example.com/document.pdf

# 从 stdin 读取文件内容
cat report.pdf | ./xparser parse --stdin -o report.md
cat report.pdf | ./xparser parse --stdin --stdin-name report.pdf -o report.md
```

### 下载图片

解析时指定 `--get-image` 后，API 返回的 `image_id` 可用此命令下载：

```bash
# 下载单张图片
./xparser download abc123def456

# 批量下载到指定目录
./xparser download abc123 def456 ghi789 -o ./images/

# 下载到指定文件
./xparser download abc123 -o page1.jpg
```

### 配置管理

```bash
# 查看当前配置
./xparser config show

# 设置配置项
./xparser config set base_url https://your-server.com
./xparser config set app_id your_app_id

# 重置为默认配置
./xparser config reset

# 查看配置文件路径
./xparser config path
```

### 自更新

```bash
# 更新到最新版本
./xparser update
```

### 调试模式

```bash
# 打印 HTTP 请求/响应详情
./xparser parse report.pdf -v
```

## 凭证管理

### 配置方式

凭证按以下优先级解析（高 → 低）：

| 优先级 | 方式 | 说明 |
|--------|------|------|
| 1 | 命令行参数 | `--app-id` 和 `--secret-code` |
| 2 | 环境变量 | `XPARSER_APP_ID` 和 `XPARSER_SECRET_CODE` |
| 3 | 配置文件 | `~/.xparser/config.yaml` |

### 交互式配置

```bash
./xparser auth
```

### 环境变量

```bash
export XPARSER_APP_ID=your_app_id
export XPARSER_SECRET_CODE=your_secret_code
./xparser parse report.pdf
```

### 命令行参数

```bash
./xparser parse report.pdf --app-id your_app_id --secret-code your_secret_code
```

### 查看当前凭证

```bash
./xparser auth --show
```

### 私有部署

```bash
# 通过参数指定
./xparser parse report.pdf --base-url https://your-server.com

# 或写入配置文件 ~/.xparser/config.yaml
# base_url: https://your-server.com
```

## 配置文件

路径：`~/.xparser/config.yaml`

```yaml
app_id: your_app_id
secret_code: your_secret_code
base_url: https://api.textin.com    # 可选，私有部署时使用
```

## 支持的文件格式

| 类型 | 格式 |
|------|------|
| 文档 | PDF, DOC, DOCX, TXT, RTF, OFD |
| 图片 | PNG, JPG, JPEG, BMP, TIFF, WebP |
| 表格 | XLS, XLSX, CSV |
| 演示 | PPT, PPTX |
| 网页 | HTML, MHTML |

限制：
- 文件大小不超过 500MB
- PDF 页数不超过 1000 页
- XLS/XLSX/CSV 每个 sheet 行数不超过 2000，列数不超过 100
- TXT 文件大小不超过 100KB
- 图片宽高在 20～20000 像素范围内

## 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2 | 参数错误 |
| 3 | 认证错误 |
| 4 | 文件错误 |
| 5 | 解析失败 |
| 6 | 服务器错误 |
| 7 | 余额不足 |

## parse 命令完整参数

```
Flags:
      --apply-chart int            图表识别: 0=关, 1=开
      --apply-document-tree int    标题层级: 0=关, 1=开 (默认 1)
      --apply-image-analysis int   LLM 图像分析: 0=关, 1=开
      --apply-merge int            段落/表格合并: 0=关, 1=开 (默认 1)
      --catalog-details int        返回目录信息: 0=否, 1=是
      --char-details int           返回字符坐标: 0=否, 1=是
      --crop-dewarp int            切边矫正: 0=关, 1=开
      --dpi int                    PDF 坐标 DPI: 72, 144, 216 (默认 144)
  -f, --format string              输出格式: md, json (默认 "md")
      --formula-level int          公式识别: 0=全部, 1=仅行间, 2=关闭
      --get-excel int              返回 Excel: 0=否, 1=是
      --get-image string           图片模式: none, page, objects, both (默认 "none")
      --image-output-type string   图片输出: default (url), base64str (默认 "default")
      --list string                从文件读取输入列表
      --markdown-details int       返回 detail 字段: 0=否, 1=是 (默认 1)
  -o, --output string              输出路径（文件或目录）
      --page-count int             解析页数 (最大 1000) (默认 1000)
      --page-details int           返回 pages 字段: 0=否, 1=是 (默认 1)
      --page-start int             PDF 起始页 (从 0 开始)
      --paratext-mode string       非正文模式: none, annotation, body (默认 "annotation")
      --parse-mode string          解析模式: auto, scan, lite, parse, vlm (默认 "scan")
      --pdf-pwd string             加密 PDF 密码
      --raw-ocr int                返回原始 OCR: 0=否, 1=是
      --remove-watermark int       去水印: 0=关, 1=开
      --stdin                      从 stdin 读取文件内容
      --stdin-list                 从 stdin 读取文件列表
      --stdin-name string          stdin 文件名提示 (默认 "stdin.pdf")
      --table-flavor string        表格格式: md, html, none (默认 "html")
      --timeout int                超时秒数 (默认 600)
      --underline-level int        下划线识别: 0=关, 1=仅空白, 2=全部

Global Flags:
      --app-id string        Textin App ID
      --base-url string      API 地址（私有部署）
      --secret-code string   Textin Secret Code
  -v, --verbose              打印 HTTP 调试信息
```
