package security

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Validator provides input validation functions
type Validator struct {
	maxSelectorLength int
	maxQueryLength    int
	maxPathLength     int
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		maxSelectorLength: 1024,
		maxQueryLength:    2048,
		maxPathLength:     4096,
	}
}

// ValidateGopherSelector validates a Gopher selector
func (v *Validator) ValidateGopherSelector(selector string) error {
	if len(selector) > v.maxSelectorLength {
		return fmt.Errorf("selector too long: %d > %d", len(selector), v.maxSelectorLength)
	}

	// Check for null bytes
	if strings.Contains(selector, "\x00") {
		return fmt.Errorf("selector contains null bytes")
	}

	// Check for CRLF injection
	if strings.Contains(selector, "\r") || strings.Contains(selector, "\n") {
		return fmt.Errorf("selector contains CRLF characters")
	}

	// Check for directory traversal
	if strings.Contains(selector, "..") {
		return fmt.Errorf("selector contains directory traversal")
	}

	return nil
}

// ValidateGeminiPath validates a Gemini URL path
func (v *Validator) ValidateGeminiPath(path string) error {
	if len(path) > v.maxPathLength {
		return fmt.Errorf("path too long: %d > %d", len(path), v.maxPathLength)
	}

	// Parse as URL to validate
	_, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check for directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains directory traversal")
	}

	return nil
}

// ValidateGeminiQuery validates a Gemini query string
func (v *Validator) ValidateGeminiQuery(query string) error {
	if len(query) > v.maxQueryLength {
		return fmt.Errorf("query too long: %d > %d", len(query), v.maxQueryLength)
	}

	// Basic sanitization
	if strings.Contains(query, "\r") || strings.Contains(query, "\n") {
		return fmt.Errorf("query contains CRLF characters")
	}

	return nil
}

// ValidateFingerUsername validates a Finger username query
func (v *Validator) ValidateFingerUsername(username string) error {
	if len(username) > 256 {
		return fmt.Errorf("username too long: %d > 256", len(username))
	}

	// Allow alphanumeric, underscore, hyphen, dot
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("invalid username format")
	}

	return nil
}

// ValidatePubkey validates a Nostr pubkey (hex format)
func (v *Validator) ValidatePubkey(pubkey string) error {
	if len(pubkey) != 64 {
		return fmt.Errorf("invalid pubkey length: %d (expected 64)", len(pubkey))
	}

	// Check if valid hex
	validHex := regexp.MustCompile(`^[0-9a-f]{64}$`)
	if !validHex.MatchString(pubkey) {
		return fmt.Errorf("pubkey must be 64-character hex string")
	}

	return nil
}

// ValidateEventID validates a Nostr event ID (hex format)
func (v *Validator) ValidateEventID(eventID string) error {
	if len(eventID) != 64 {
		return fmt.Errorf("invalid event ID length: %d (expected 64)", len(eventID))
	}

	// Check if valid hex
	validHex := regexp.MustCompile(`^[0-9a-f]{64}$`)
	if !validHex.MatchString(eventID) {
		return fmt.Errorf("event ID must be 64-character hex string")
	}

	return nil
}

// ValidateNpub validates a Nostr npub (bech32 format)
func (v *Validator) ValidateNpub(npub string) error {
	if !strings.HasPrefix(npub, "npub1") {
		return fmt.Errorf("npub must start with 'npub1'")
	}

	if len(npub) < 63 || len(npub) > 65 {
		return fmt.Errorf("invalid npub length: %d", len(npub))
	}

	// Basic bech32 character check (alphanumeric, no 'b', 'i', 'o')
	validBech32 := regexp.MustCompile(`^npub1[qpzry9x8gf2tvdw0s3jn54khce6mua7l]+$`)
	if !validBech32.MatchString(npub) {
		return fmt.Errorf("invalid npub format")
	}

	return nil
}

// SanitizeInput removes potentially dangerous characters
func (v *Validator) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove CRLF
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\n", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// SanitizePath sanitizes a file path
func (v *Validator) SanitizePath(path string) string {
	// Remove directory traversal
	path = strings.ReplaceAll(path, "..", "")

	// Remove leading slashes (to prevent absolute paths)
	path = strings.TrimLeft(path, "/")

	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")

	return path
}

// ValidateInteger validates an integer within a range
func (v *Validator) ValidateInteger(value, min, max int) error {
	if value < min {
		return fmt.Errorf("value %d is below minimum %d", value, min)
	}

	if value > max {
		return fmt.Errorf("value %d is above maximum %d", value, max)
	}

	return nil
}

// ValidatePageNumber validates a pagination page number
func (v *Validator) ValidatePageNumber(page int) error {
	return v.ValidateInteger(page, 1, 10000)
}

// ValidateLimit validates a query limit
func (v *Validator) ValidateLimit(limit int) error {
	return v.ValidateInteger(limit, 1, 1000)
}

// IsSafeHTML checks if HTML contains no script tags
// This is a basic check - production would use a proper HTML sanitizer
func (v *Validator) IsSafeHTML(html string) bool {
	dangerous := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
	}

	lowerHTML := strings.ToLower(html)
	for _, pattern := range dangerous {
		if strings.Contains(lowerHTML, pattern) {
			return false
		}
	}

	return true
}

// ValidateURL validates a URL
func (v *Validator) ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "gemini" && u.Scheme != "gopher" {
		return fmt.Errorf("invalid URL scheme: %s", u.Scheme)
	}

	return nil
}

// ValidateHost validates a hostname
func (v *Validator) ValidateHost(host string) error {
	if len(host) > 253 {
		return fmt.Errorf("hostname too long")
	}

	// Basic hostname validation
	validHost := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	if !validHost.MatchString(host) {
		return fmt.Errorf("invalid hostname format")
	}

	return nil
}

// ValidatePort validates a port number
func (v *Validator) ValidatePort(port int) error {
	return v.ValidateInteger(port, 1, 65535)
}

// InputSanitizer provides sanitization methods
type InputSanitizer struct {
	validator *Validator
}

// NewInputSanitizer creates a new input sanitizer
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		validator: NewValidator(),
	}
}

// SanitizeAndValidateSelector sanitizes and validates a Gopher selector
func (is *InputSanitizer) SanitizeAndValidateSelector(selector string) (string, error) {
	selector = is.validator.SanitizeInput(selector)

	if err := is.validator.ValidateGopherSelector(selector); err != nil {
		return "", err
	}

	return selector, nil
}

// SanitizeAndValidatePath sanitizes and validates a Gemini path
func (is *InputSanitizer) SanitizeAndValidatePath(path string) (string, error) {
	path = is.validator.SanitizePath(path)

	if err := is.validator.ValidateGeminiPath(path); err != nil {
		return "", err
	}

	return path, nil
}

// SanitizeAndValidateQuery sanitizes and validates a query string
func (is *InputSanitizer) SanitizeAndValidateQuery(query string) (string, error) {
	query = is.validator.SanitizeInput(query)

	if err := is.validator.ValidateGeminiQuery(query); err != nil {
		return "", err
	}

	return query, nil
}
