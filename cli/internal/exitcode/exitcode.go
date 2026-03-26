// Package exitcode maps API errors to CLI exit codes.
package exitcode

// Exit codes
const (
	Success      = 0
	GeneralError = 1
	UsageError   = 2
	AuthError    = 3
	FileError    = 4
	ParseFailed  = 5
	ServerError  = 6
	QuotaError   = 7
)

// ErrorInfo holds error details for CLI output.
type ErrorInfo struct {
	Code    int
	Message string
	Hint    string
}

// FromAPICode maps a Textin API error code to an ErrorInfo.
func FromAPICode(apiCode int, message string) *ErrorInfo {
	switch apiCode {
	case 200:
		return nil

	case 40101:
		return &ErrorInfo{
			Code:    AuthError,
			Message: message,
			Hint:    "x-ti-app-id or x-ti-secret-code is empty. Run 'xparser auth' to configure.",
		}
	case 40102:
		return &ErrorInfo{
			Code:    AuthError,
			Message: message,
			Hint:    "Invalid app-id or secret-code. Run 'xparser auth' to reconfigure.",
		}
	case 40103:
		return &ErrorInfo{
			Code:    AuthError,
			Message: message,
			Hint:    "Client IP is not whitelisted. Check your Textin console settings.",
		}

	case 40003:
		return &ErrorInfo{
			Code:    QuotaError,
			Message: message,
			Hint:    "Insufficient balance. Please recharge at https://www.textin.com",
		}
	case 40004:
		return &ErrorInfo{
			Code:    UsageError,
			Message: message,
			Hint:    "Parameter error. Check your command arguments.",
		}

	case 40301:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "Image type not supported.",
		}
	case 40302:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "File size exceeds 500MB limit.",
		}
	case 40303:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "File type not supported. Supported: png, jpg, pdf, doc(x), xls(x), ppt(x), html, txt, etc.",
		}
	case 40304:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "Image dimensions out of range (20-20000px).",
		}
	case 40305:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "No file uploaded. Provide a file path or URL.",
		}

	case 40422:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "File is corrupted.",
		}
	case 40423:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "PDF password required or incorrect. Use --pdf-pwd to provide the password.",
		}
	case 40424:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint: "Page number out of range. Use --page-start and --page-count to adjust:\n" +
				"  xparser parse doc.pdf --page-start 0 --page-count 100",
		}
	case 40425:
		return &ErrorInfo{
			Code:    FileError,
			Message: message,
			Hint:    "Input file format not supported.",
		}
	case 40427:
		return &ErrorInfo{
			Code:    UsageError,
			Message: message,
			Hint:    "DPI must be 72, 144, or 216.",
		}
	case 40428:
		return &ErrorInfo{
			Code:    ParseFailed,
			Message: message,
			Hint:    "Office file conversion failed or timed out. Try again.",
		}

	case 50207:
		return &ErrorInfo{
			Code:    ParseFailed,
			Message: message,
			Hint:    "Partial pages failed to parse. Check the output for details.",
		}

	case 500, 30203:
		return &ErrorInfo{
			Code:    ServerError,
			Message: message,
			Hint:    "Server error. Please try again later.",
		}

	default:
		return &ErrorInfo{
			Code:    GeneralError,
			Message: message,
			Hint:    "",
		}
	}
}
