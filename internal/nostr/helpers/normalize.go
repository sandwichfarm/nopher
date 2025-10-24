package helpers

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// NormalizePubkey converts npub or hex pubkey to hex format
// Returns hex pubkey or error if invalid
func NormalizePubkey(input string) (string, error) {
	input = strings.TrimSpace(input)

	// Check if it's an npub
	if strings.HasPrefix(input, "npub1") {
		prefix, pubkey, err := nip19.Decode(input)
		if err != nil {
			return "", fmt.Errorf("invalid npub: %w", err)
		}
		if prefix != "npub" {
			return "", fmt.Errorf("expected npub, got %s", prefix)
		}
		return pubkey.(string), nil
	}

	// Check if it's a hex pubkey (64 chars)
	if len(input) == 64 {
		// Validate hex
		if _, err := hex.DecodeString(input); err != nil {
			return "", fmt.Errorf("invalid hex pubkey: %w", err)
		}
		return input, nil
	}

	return "", fmt.Errorf("invalid pubkey format (expected npub1... or 64-char hex)")
}

// NormalizeEventID converts note1 or hex event ID to hex format
// Returns hex event ID or error if invalid
func NormalizeEventID(input string) (string, error) {
	input = strings.TrimSpace(input)

	// Check if it's a note1 (bech32 event ID)
	if strings.HasPrefix(input, "note1") {
		prefix, eventID, err := nip19.Decode(input)
		if err != nil {
			return "", fmt.Errorf("invalid note ID: %w", err)
		}
		if prefix != "note" {
			return "", fmt.Errorf("expected note, got %s", prefix)
		}
		return eventID.(string), nil
	}

	// Check if it's a hex event ID (64 chars)
	if len(input) == 64 {
		// Validate hex
		if _, err := hex.DecodeString(input); err != nil {
			return "", fmt.Errorf("invalid hex event ID: %w", err)
		}
		return input, nil
	}

	return "", fmt.Errorf("invalid event ID format (expected note1... or 64-char hex)")
}

// EncodePubkey converts hex pubkey to npub
func EncodePubkey(hexPubkey string) (string, error) {
	if len(hexPubkey) != 64 {
		return "", fmt.Errorf("pubkey must be 64 hex characters")
	}

	npub, err := nip19.EncodePublicKey(hexPubkey)
	if err != nil {
		return "", fmt.Errorf("failed to encode pubkey: %w", err)
	}

	return npub, nil
}

// EncodeEventID converts hex event ID to note1
func EncodeEventID(hexEventID string) (string, error) {
	if len(hexEventID) != 64 {
		return "", fmt.Errorf("event ID must be 64 hex characters")
	}

	note, err := nip19.EncodeNote(hexEventID)
	if err != nil {
		return "", fmt.Errorf("failed to encode event ID: %w", err)
	}

	return note, nil
}

// IsValidEvent performs basic validation on a Nostr event
func IsValidEvent(event *nostr.Event) bool {
	if event == nil {
		return false
	}

	// Check required fields
	if event.ID == "" || event.PubKey == "" {
		return false
	}

	// Validate ID is 64 hex chars
	if len(event.ID) != 64 {
		return false
	}
	if _, err := hex.DecodeString(event.ID); err != nil {
		return false
	}

	// Validate pubkey is 64 hex chars
	if len(event.PubKey) != 64 {
		return false
	}
	if _, err := hex.DecodeString(event.PubKey); err != nil {
		return false
	}

	return true
}
