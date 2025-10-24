package helpers

import (
	"strings"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

func TestNormalizePubkey(t *testing.T) {
	// Valid hex pubkey
	validHex := "9822242c03e3af313cc6abd17af6a9b777f1aa18f5b347020a84664629212173"
	// Valid npub (same as above hex)
	validNpub := "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid hex pubkey",
			input:       validHex,
			expected:    validHex,
			shouldError: false,
		},
		{
			name:        "valid npub",
			input:       validNpub,
			expected:    validHex,
			shouldError: false,
		},
		{
			name:        "npub with whitespace",
			input:       "  " + validNpub + "  ",
			expected:    validHex,
			shouldError: false,
		},
		{
			name:        "invalid hex (too short)",
			input:       "9822242c03e3af313cc6abd17af6a9b777f1aa18f5b347020a84664629212",
			shouldError: true,
		},
		{
			name:        "invalid hex (non-hex chars)",
			input:       "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			shouldError: true,
		},
		{
			name:        "invalid npub (bad checksum)",
			input:       "npub1test0000000000000000000000000000000000000000000000000000000",
			shouldError: true,
		},
		{
			name:        "empty string",
			input:       "",
			shouldError: true,
		},
		{
			name:        "random text",
			input:       "not a pubkey",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizePubkey(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestNormalizeEventID(t *testing.T) {
	// Valid hex event ID
	validHex := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid hex event ID",
			input:       validHex,
			expected:    validHex,
			shouldError: false,
		},
		{
			name:        "event ID with whitespace",
			input:       "  " + validHex + "  ",
			expected:    validHex,
			shouldError: false,
		},
		{
			name:        "invalid hex (too short)",
			input:       "abcdef1234567890abcdef1234567890abcdef1234567890abcdef12345678",
			shouldError: true,
		},
		{
			name:        "invalid hex (non-hex chars)",
			input:       "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			shouldError: true,
		},
		{
			name:        "empty string",
			input:       "",
			shouldError: true,
		},
		{
			name:        "random text",
			input:       "not an event ID",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeEventID(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestEncodePubkey(t *testing.T) {
	validHex := "9822242c03e3af313cc6abd17af6a9b777f1aa18f5b347020a84664629212173"
	expectedNpub := "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid hex pubkey",
			input:       validHex,
			expected:    expectedNpub,
			shouldError: false,
		},
		{
			name:        "invalid hex (too short)",
			input:       "9822242c03e3af313cc6abd17af6a9b777f1aa18f5b347020a8466462921217",
			shouldError: true,
		},
		{
			name:        "invalid hex (too long)",
			input:       "9822242c03e3af313cc6abd17af6a9b777f1aa18f5b347020a84664629212173aa",
			shouldError: true,
		},
		{
			name:        "empty string",
			input:       "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodePubkey(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestEncodeEventID(t *testing.T) {
	validHex := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "valid hex event ID",
			input:       validHex,
			shouldError: false,
		},
		{
			name:        "invalid hex (too short)",
			input:       "abcdef1234567890abcdef1234567890abcdef1234567890abcdef12345678",
			shouldError: true,
		},
		{
			name:        "empty string",
			input:       "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeEventID(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Just verify it starts with note1
				if !strings.HasPrefix(result, "note1") {
					t.Errorf("Expected note1... encoding, got %s", result)
				}
			}
		})
	}
}

func TestIsValidEvent(t *testing.T) {
	validID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	validPubkey := "9822242c03e3af313cc6abd17af6a9b777f1aa18f5b347020a84664629212173"

	tests := []struct {
		name     string
		event    *nostr.Event
		expected bool
	}{
		{
			name: "valid event",
			event: &nostr.Event{
				ID:      validID,
				PubKey:  validPubkey,
				Kind:    1,
				Content: "Hello world",
			},
			expected: true,
		},
		{
			name:     "nil event",
			event:    nil,
			expected: false,
		},
		{
			name: "missing ID",
			event: &nostr.Event{
				PubKey:  validPubkey,
				Kind:    1,
				Content: "Hello",
			},
			expected: false,
		},
		{
			name: "missing pubkey",
			event: &nostr.Event{
				ID:      validID,
				Kind:    1,
				Content: "Hello",
			},
			expected: false,
		},
		{
			name: "invalid ID (too short)",
			event: &nostr.Event{
				ID:      "abc123",
				PubKey:  validPubkey,
				Kind:    1,
				Content: "Hello",
			},
			expected: false,
		},
		{
			name: "invalid pubkey (non-hex)",
			event: &nostr.Event{
				ID:      validID,
				PubKey:  "notahexpubkey",
				Kind:    1,
				Content: "Hello",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEvent(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
