// Package exitcode defines CLI exit codes and structured API error mapping.
package exitcode

import "encoding/json"

// Exit codes per V1 spec.
const (
	Success      = 0 // Success
	GeneralError = 1 // General API or unknown error — check network, retry
	UsageError   = 2 // Invalid parameters / usage error — check command syntax
	APIError     = 3 // API returned an error — structured info on stderr
)

// APIErrorInfo is the structured error written to stderr when exit code is 3.
type APIErrorInfo struct {
	ErrorType  string `json:"error_type"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Retryable  bool   `json:"retryable"`
}

// JSON returns the JSON-encoded string for stderr output.
func (e *APIErrorInfo) JSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// errorTemplate defines the static parts of an API error mapping.
type errorTemplate struct {
	errorType  string
	message    string
	suggestion string
	retryable  bool
}

// apiErrorMap maps Textin API error codes to their templates.
var apiErrorMap = map[int]errorTemplate{
	// ── Server errors ──
	500: {
		errorType:  "server_error",
		message:    "Server internal error",
		suggestion: "Please try again later. If the issue persists, contact support",
		retryable:  true,
	},
	30203: {
		errorType:  "server_error",
		message:    "Base service failure",
		suggestion: "Internal service is temporarily unavailable. Please retry later",
		retryable:  true,
	},

	// ── Quota / rate-limit errors ──
	40003: {
		errorType:  "quota_error",
		message:    "Insufficient balance",
		suggestion: "Account balance is depleted. Top up at https://www.textin.com/console/dashboard/setting, or use --api free",
		retryable:  false,
	},

	// ── Parameter errors ──
	40004: {
		errorType:  "param_error",
		message:    "Invalid parameter",
		suggestion: "Check the command flags and values. Run xparse-cli parse --help for usage",
		retryable:  false,
	},
	40007: {
		errorType:  "param_error",
		message:    "Service not found or not published",
		suggestion: "The requested service does not exist. Check your configuration",
		retryable:  false,
	},
	40008: {
		errorType:  "param_error",
		message:    "Service not activated",
		suggestion: "The service is not activated. Activate it at https://www.textin.com/market/detail/pdf_to_markdown",
		retryable:  false,
	},

	// ── Auth errors ──
	40101: {
		errorType:  "auth_error",
		message:    "Credentials empty",
		suggestion: "x-ti-app-id or x-ti-secret-code is empty. Get credentials at https://www.textin.com/console/dashboard/setting",
		retryable:  false,
	},
	40102: {
		errorType:  "auth_error",
		message:    "Authentication failed",
		suggestion: "Invalid x-ti-app-id or x-ti-secret-code. Get valid credentials at https://www.textin.com/console/dashboard/setting",
		retryable:  false,
	},
	40103: {
		errorType:  "auth_error",
		message:    "IP not whitelisted",
		suggestion: "Your client IP is not in the allowlist. Check IP whitelist settings at https://www.textin.com/console/dashboard/setting",
		retryable:  false,
	},

	// ── File / format errors ──
	40301: {
		errorType:  "format_error",
		message:    "Image type not supported",
		suggestion: "Supported image types: JPG, JPEG, PNG, BMP, TIFF, GIF, WebP",
		retryable:  false,
	},
	40302: {
		errorType:  "file_error",
		message:    "File size exceeds limit",
		suggestion: "Maximum file size is 500MB. Split or compress the file before uploading",
		retryable:  false,
	},
	40303: {
		errorType:  "format_error",
		message:    "File type not supported",
		suggestion: "Supported formats: PDF, PNG, JPG, BMP, TIFF, WebP, DOC(X), PPT(X), XLS(X), HTML, MHTML, TXT, RTF, OFD, CSV",
		retryable:  false,
	},
	40304: {
		errorType:  "file_error",
		message:    "Image dimensions out of range",
		suggestion: "Image width and height must be 20-20000px (aspect ratio < 2) or 20-10000px (other). Resize the image",
		retryable:  false,
	},
	40305: {
		errorType:  "file_error",
		message:    "No file uploaded",
		suggestion: "Provide a file path or URL as input",
		retryable:  false,
	},
	40306: {
		errorType:  "rate_limit_error",
		message:    "QPS limit exceeded",
		suggestion: "Too many requests per second. Wait a moment and retry",
		retryable:  true,
	},
	40307: {
		errorType:  "quota_error",
		message:    "Daily free quota exhausted",
		suggestion: "Today's free API quota is used up. Try again tomorrow, or switch to paid API with --api paid",
		retryable:  false,
	},

	// ── URL / request errors ──
	40400: {
		errorType:  "param_error",
		message:    "Invalid request URL",
		suggestion: "The provided URL is invalid. Check that the URL is correct and accessible",
		retryable:  false,
	},
	40422: {
		errorType:  "file_error",
		message:    "The file is corrupted",
		suggestion: "The file may be damaged. Re-download or use a different copy of the file",
		retryable:  false,
	},
	40423: {
		errorType:  "password_error",
		message:    "Password required or incorrect password",
		suggestion: "Use --password to provide the correct document password",
		retryable:  false,
	},
	40424: {
		errorType:  "page_range_error",
		message:    "Page number out of range",
		suggestion: "The --page-range exceeds the document page count. Adjust to a valid range",
		retryable:  false,
	},
	40425: {
		errorType:  "format_error",
		message:    "The input file format is not supported",
		suggestion: "Supported formats: PDF, PNG, JPG, BMP, TIFF, WebP, DOC(X), PPT(X), XLS(X), HTML, MHTML, TXT, RTF, OFD, CSV",
		retryable:  false,
	},
	40427: {
		errorType:  "param_error",
		message:    "DPI value is not in the allowed list",
		suggestion: "Allowed DPI values: 72, 144, 216",
		retryable:  false,
	},
	40428: {
		errorType:  "parse_error",
		message:    "Office file conversion failed or timed out",
		suggestion: "Try again. If the issue persists, convert the file to PDF before parsing",
		retryable:  true,
	},
	40429: {
		errorType:  "parse_error",
		message:    "Unsupported engine",
		suggestion: "The requested parsing engine is not available. Contact support",
		retryable:  false,
	},

	// ── Partial errors ──
	50207: {
		errorType:  "partial_error",
		message:    "Some pages failed to parse",
		suggestion: "Check the output for partial results. Retry may resolve the issue",
		retryable:  true,
	},
}

// FromAPICode maps a Textin API error code to a structured APIErrorInfo.
// Returns nil for success (200).
func FromAPICode(apiCode int, message string) *APIErrorInfo {
	if apiCode == 200 {
		return nil
	}

	if tmpl, ok := apiErrorMap[apiCode]; ok {
		return &APIErrorInfo{
			ErrorType:  tmpl.errorType,
			Code:       apiCode,
			Message:    coalesce(message, tmpl.message),
			Suggestion: tmpl.suggestion,
			Retryable:  tmpl.retryable,
		}
	}

	return &APIErrorInfo{
		ErrorType:  "unknown_error",
		Code:       apiCode,
		Message:    coalesce(message, "Unknown API error"),
		Suggestion: "Use --verbose for details. If the issue persists, contact support",
		Retryable:  true,
	}
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
