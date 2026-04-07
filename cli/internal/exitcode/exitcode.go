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
	// Auth errors
	40101: {
		errorType:  "auth_error",
		message:    "Authentication failed",
		suggestion: "Invalid or missing x-ti-app-id / x-ti-secret-code. Get valid credentials at https://www.textin.com/console/dashboard/setting",
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
		message:    "Quota exceeded or account suspended",
		suggestion: "Check your account status at https://www.textin.com/console/dashboard/setting",
		retryable:  false,
	},

	// Client errors
	40422: {
		errorType:  "password_error",
		message:    "Password required or incorrect",
		suggestion: "Use --password to provide the correct document password",
		retryable:  false,
	},
	40424: {
		errorType:  "page_range_error",
		message:    "Page number out of range",
		suggestion: "Adjust --page-range to a valid range within the document",
		retryable:  false,
	},
	40425: {
		errorType:  "format_error",
		message:    "Unsupported file format",
		suggestion: "Supported formats: PDF, PNG, JPG, BMP, TIFF, WebP, DOC(X), PPT(X), XLS(X), HTML, MHTML, TXT, RTF, OFD, CSV",
		retryable:  false,
	},
	40426: {
		errorType:  "file_error",
		message:    "File is corrupted or unreadable",
		suggestion: "Re-download or use a different copy of the file",
		retryable:  false,
	},
	40427: {
		errorType:  "param_error",
		message:    "Invalid DPI value",
		suggestion: "Allowed DPI values: 72, 144, 216. This may be an internal configuration issue — contact support if it persists",
		retryable:  false,
	},
	40428: {
		errorType:  "parse_error",
		message:    "Office file conversion failed",
		suggestion: "Try again or convert the file to PDF before parsing",
		retryable:  true,
	},
	40429: {
		errorType:  "file_error",
		message:    "PDF content is empty",
		suggestion: "The PDF has no extractable content — verify the file is valid",
		retryable:  false,
	},

	// Server errors
	500: {
		errorType:  "server_error",
		message:    "Server internal error",
		suggestion: "Please try again later. If the issue persists, contact support",
		retryable:  true,
	},
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
