// Package exitcode defines CLI exit codes, fixed error messages, and API error mapping.
package exitcode

// Exit codes per V1 spec.
const (
	Success      = 0 // Success
	GeneralError = 1 // General error — plain text on stderr
	UsageError   = 2 // Usage / parameter error — plain text on stderr
	APIError     = 3 // API error — "api_code：message" on stderr
)

// ── Exit 1: General error messages ──

const (
	ErrCredentialsConfig = "credentials configuration error"
	ErrNetworkRequest    = "network or request failed"
	ErrNoResultData      = "API returned success but no result data"
	ErrCreateOutputDir   = "failed to create output directory"
	ErrBatchPartial      = "batch completed with errors"
	ErrSaveResult        = "failed to save result"
)

// ── Exit 2: Usage error messages ──

const (
	ErrInvalidView      = "invalid --view value, must be 'markdown' or 'json'"
	ErrInvalidAPI       = "invalid --api value, must be 'free' or 'paid'"
	ErrOpenListFile     = "failed to open or read --list file"
	ErrNoInput          = "no input files specified"
	ErrFlagValueNotFile = "flag value is not a file"
	ErrFileNotFound     = "file not found"
	ErrListRequiresOut  = "--list requires --output"
	ErrMultiRequiresOut = "multiple inputs require --output"
	ErrPaidNoCreds      = "paid API requires credentials"
)

// ── Exit 3: API error mapping ──

// APIErrorInfo holds the mapped API error for exit code 3.
type APIErrorInfo struct {
	APICode   int
	Message   string
	Retryable bool
}

// apiErrorMap maps Textin API error codes to their official messages.
// Messages must match the official API documentation exactly.
var apiErrorMap = map[int]errorTemplate{
	400:   {message: "上传的文件不能为空", retryable: false},
	500:   {message: "服务器内部错误", retryable: true},
	30203: {message: "基础服务故障，请稍后重试", retryable: true},
	40003: {message: "余额不足，请充值后再使用", retryable: false},
	40004: {message: "参数错误，请查看技术文档，检查传参", retryable: false},
	40007: {message: "机器人不存在或未发布", retryable: false},
	40008: {message: "机器人未开通，请至市场开通后重试", retryable: false},
	40101: {message: "x-ti-app-id 或 x-ti-secret-code 为空", retryable: false},
	40102: {message: "x-ti-app-id 或 x-ti-secret-code 无效，验证失败", retryable: false},
	40103: {message: "客户端IP不在白名单", retryable: false},
	40301: {message: "图片类型不支持", retryable: false},
	40302: {message: "上传文件大小不符，文件大小不超过 500M", retryable: false},
	40303: {message: "文件类型不支持", retryable: false},
	40304: {message: "图片尺寸不符", retryable: false},
	40305: {message: "识别文件未上传", retryable: false},
	40306: {message: "qps超过限制", retryable: true},
	40307: {message: "今日免费额度已用完", retryable: false},
	40400: {message: "无效的请求链接，请检查链接是否正确", retryable: false},
	40422: {message: "文件损坏", retryable: false},
	40423: {message: "PDF密码错误", retryable: false},
	40424: {message: "页数设置超出文件范围", retryable: false},
	40425: {message: "文件格式不支持", retryable: false},
	40427: {message: "DPI参数不在支持列表中", retryable: false},
	40428: {message: "word和ppt转pdf失败或者超时", retryable: true},
	40429: {message: "不支持的引擎", retryable: false},
	50207: {message: "部分页面解析失败", retryable: true},
}

// errorTemplate defines the static parts of an API error mapping.
type errorTemplate struct {
	message   string
	retryable bool
}

// FromAPICode maps a Textin API error code to APIErrorInfo.
// Returns nil for success (200).
func FromAPICode(apiCode int) *APIErrorInfo {
	if apiCode == 200 {
		return nil
	}

	if tmpl, ok := apiErrorMap[apiCode]; ok {
		return &APIErrorInfo{
			APICode:   apiCode,
			Message:   tmpl.message,
			Retryable: tmpl.retryable,
		}
	}

	return &APIErrorInfo{
		APICode:   apiCode,
		Message:   "未知错误",
		Retryable: true,
	}
}
