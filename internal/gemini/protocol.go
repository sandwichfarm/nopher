package gemini

import "fmt"

// Status represents a Gemini protocol status code
type Status int

// Gemini status codes per specification
const (
	// 1x - Input
	StatusInput          Status = 10
	StatusSensitiveInput Status = 11

	// 2x - Success
	StatusSuccess Status = 20

	// 3x - Redirect
	StatusRedirectTemporary Status = 30
	StatusRedirectPermanent Status = 31

	// 4x - Temporary Failure
	StatusTemporaryFailure     Status = 40
	StatusServerUnavailable    Status = 41
	StatusCGIError             Status = 42
	StatusProxyError           Status = 43
	StatusSlowDown             Status = 44

	// 5x - Permanent Failure
	StatusPermanentFailure     Status = 50
	StatusNotFound             Status = 51
	StatusGone                 Status = 52
	StatusProxyRequestRefused  Status = 53
	StatusBadRequest           Status = 59

	// 6x - Client Certificate Required
	StatusClientCertRequired      Status = 60
	StatusCertNotAuthorized       Status = 61
	StatusCertNotValid            Status = 62
)

// String returns a human-readable description of the status
func (s Status) String() string {
	switch s {
	case StatusInput:
		return "Input"
	case StatusSensitiveInput:
		return "Sensitive Input"
	case StatusSuccess:
		return "Success"
	case StatusRedirectTemporary:
		return "Redirect - Temporary"
	case StatusRedirectPermanent:
		return "Redirect - Permanent"
	case StatusTemporaryFailure:
		return "Temporary Failure"
	case StatusServerUnavailable:
		return "Server Unavailable"
	case StatusCGIError:
		return "CGI Error"
	case StatusProxyError:
		return "Proxy Error"
	case StatusSlowDown:
		return "Slow Down"
	case StatusPermanentFailure:
		return "Permanent Failure"
	case StatusNotFound:
		return "Not Found"
	case StatusGone:
		return "Gone"
	case StatusProxyRequestRefused:
		return "Proxy Request Refused"
	case StatusBadRequest:
		return "Bad Request"
	case StatusClientCertRequired:
		return "Client Certificate Required"
	case StatusCertNotAuthorized:
		return "Certificate Not Authorized"
	case StatusCertNotValid:
		return "Certificate Not Valid"
	default:
		return "Unknown Status"
	}
}

// FormatResponse formats a Gemini protocol response
// Format: <STATUS><SPACE><META><CRLF>[<BODY>]
func FormatResponse(status Status, meta string, body string) []byte {
	// Ensure meta is not too long (spec recommends max 1024 bytes for entire header)
	if len(meta) > 1000 {
		meta = meta[:1000]
	}

	// Format header: status code (2 digits) + space + meta + CRLF
	header := fmt.Sprintf("%d %s\r\n", status, meta)

	// Combine header and body
	response := []byte(header)
	if body != "" {
		response = append(response, []byte(body)...)
	}

	return response
}

// FormatSuccessResponse creates a successful response with text/gemini content
func FormatSuccessResponse(body string) []byte {
	return FormatResponse(StatusSuccess, "text/gemini; charset=utf-8", body)
}

// FormatErrorResponse creates an error response
func FormatErrorResponse(status Status, message string) []byte {
	return FormatResponse(status, message, "")
}

// FormatInputResponse creates an input request response
func FormatInputResponse(prompt string, sensitive bool) []byte {
	status := StatusInput
	if sensitive {
		status = StatusSensitiveInput
	}
	return FormatResponse(status, prompt, "")
}

// FormatRedirectResponse creates a redirect response
func FormatRedirectResponse(url string, permanent bool) []byte {
	status := StatusRedirectTemporary
	if permanent {
		status = StatusRedirectPermanent
	}
	return FormatResponse(status, url, "")
}
