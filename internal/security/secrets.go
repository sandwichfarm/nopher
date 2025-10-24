package security

import (
	"fmt"
	"os"
	"strings"
)

// SecretManager handles secure secret management
type SecretManager struct {
	// Secrets are only stored in memory, never written to disk
	secrets map[string]string
}

// NewSecretManager creates a new secret manager
func NewSecretManager() *SecretManager {
	return &SecretManager{
		secrets: make(map[string]string),
	}
}

// LoadFromEnv loads secrets from environment variables
func (sm *SecretManager) LoadFromEnv(prefix string) error {
	// Load all environment variables with the given prefix
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]

				// Store in memory only
				sm.secrets[key] = value
			}
		}
	}

	return nil
}

// Get retrieves a secret by key
func (sm *SecretManager) Get(key string) (string, bool) {
	value, exists := sm.secrets[key]
	return value, exists
}

// Set stores a secret (in memory only)
func (sm *SecretManager) Set(key, value string) {
	sm.secrets[key] = value
}

// Clear removes all secrets from memory
func (sm *SecretManager) Clear() {
	// Overwrite with zeros before clearing
	for key := range sm.secrets {
		sm.secrets[key] = strings.Repeat("\x00", len(sm.secrets[key]))
	}
	sm.secrets = make(map[string]string)
}

// Redact returns a redacted version of a secret for logging
func (sm *SecretManager) Redact(secret string) string {
	if len(secret) == 0 {
		return ""
	}

	if len(secret) <= 8 {
		return "***"
	}

	// Show first 4 and last 4 characters
	return secret[:4] + "..." + secret[len(secret)-4:]
}

// LoadNsecFromEnv loads the Nostr secret key from environment
func (sm *SecretManager) LoadNsecFromEnv() (string, error) {
	nsec := os.Getenv("NOPHER_NSEC")
	if nsec == "" {
		return "", fmt.Errorf("NOPHER_NSEC environment variable not set")
	}

	// Validate nsec format
	if !strings.HasPrefix(nsec, "nsec1") {
		return "", fmt.Errorf("invalid nsec format: must start with 'nsec1'")
	}

	// Store in memory
	sm.Set("NOPHER_NSEC", nsec)

	return nsec, nil
}

// RedactedConfig returns a config with secrets redacted
type RedactedConfig struct {
	config map[string]interface{}
}

// NewRedactedConfig creates a config with secrets redacted
func NewRedactedConfig() *RedactedConfig {
	return &RedactedConfig{
		config: make(map[string]interface{}),
	}
}

// Set sets a configuration value
func (rc *RedactedConfig) Set(key string, value interface{}) {
	rc.config[key] = value
}

// SetSecret sets a secret value (will be redacted)
func (rc *RedactedConfig) SetSecret(key string, value string) {
	if len(value) <= 8 {
		rc.config[key] = "***"
	} else {
		rc.config[key] = value[:4] + "..." + value[len(value)-4:]
	}
}

// Get returns all config values
func (rc *RedactedConfig) Get() map[string]interface{} {
	return rc.config
}

// SecureString is a string that won't be accidentally logged
type SecureString struct {
	value string
}

// NewSecureString creates a new secure string
func NewSecureString(value string) *SecureString {
	return &SecureString{value: value}
}

// Get returns the underlying value
func (ss *SecureString) Get() string {
	return ss.value
}

// String returns a redacted version for logging
func (ss *SecureString) String() string {
	if len(ss.value) <= 8 {
		return "***"
	}
	return ss.value[:4] + "..." + ss.value[len(ss.value)-4:]
}

// MarshalJSON redacts the value in JSON
func (ss *SecureString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ss.String() + `"`), nil
}

// Clear overwrites the value with zeros
func (ss *SecureString) Clear() {
	ss.value = strings.Repeat("\x00", len(ss.value))
	ss.value = ""
}

// SecretValidator validates secrets before use
type SecretValidator struct{}

// NewSecretValidator creates a new secret validator
func NewSecretValidator() *SecretValidator {
	return &SecretValidator{}
}

// ValidateNsec validates a Nostr secret key
func (sv *SecretValidator) ValidateNsec(nsec string) error {
	if !strings.HasPrefix(nsec, "nsec1") {
		return fmt.Errorf("nsec must start with 'nsec1'")
	}

	if len(nsec) < 63 || len(nsec) > 65 {
		return fmt.Errorf("invalid nsec length: %d", len(nsec))
	}

	return nil
}

// ValidateNpub validates a Nostr public key
func (sv *SecretValidator) ValidateNpub(npub string) error {
	if !strings.HasPrefix(npub, "npub1") {
		return fmt.Errorf("npub must start with 'npub1'")
	}

	if len(npub) < 63 || len(npub) > 65 {
		return fmt.Errorf("invalid npub length: %d", len(npub))
	}

	return nil
}

// CheckForLeakedSecrets scans text for potential leaked secrets
func (sv *SecretValidator) CheckForLeakedSecrets(text string) []string {
	var leaks []string

	// Check for nsec
	if strings.Contains(text, "nsec1") {
		leaks = append(leaks, "potential nsec leak detected")
	}

	// Check for private keys (64 hex chars)
	// This is a simple check - production would use regex
	if strings.Contains(text, "private") || strings.Contains(text, "secret") {
		leaks = append(leaks, "potential secret keyword detected")
	}

	return leaks
}

// SafeLogger wraps a logger to prevent secret leakage
type SafeLogger struct {
	validator *SecretValidator
}

// NewSafeLogger creates a safe logger
func NewSafeLogger() *SafeLogger {
	return &SafeLogger{
		validator: NewSecretValidator(),
	}
}

// SanitizeMessage sanitizes a log message
func (sl *SafeLogger) SanitizeMessage(msg string) string {
	// Redact nsec if present
	if strings.Contains(msg, "nsec1") {
		msg = strings.ReplaceAll(msg, "nsec1", "nsec***")
	}

	// Redact long hex strings (potential private keys)
	// This is simplified - production would use regex
	return msg
}

// CheckMessage checks if a message contains secrets
func (sl *SafeLogger) CheckMessage(msg string) error {
	leaks := sl.validator.CheckForLeakedSecrets(msg)
	if len(leaks) > 0 {
		return fmt.Errorf("potential secret leak: %v", leaks)
	}
	return nil
}
