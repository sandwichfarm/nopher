package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// KeyBuilder helps build cache keys
type KeyBuilder struct {
	parts []string
}

// NewKeyBuilder creates a new key builder
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{
		parts: make([]string, 0, 8),
	}
}

// Add adds a part to the key
func (kb *KeyBuilder) Add(part string) *KeyBuilder {
	kb.parts = append(kb.parts, part)
	return kb
}

// Build builds the final key
func (kb *KeyBuilder) Build() string {
	return strings.Join(kb.parts, ":")
}

// BuildHashed builds the final key with a hash suffix for long keys
func (kb *KeyBuilder) BuildHashed() string {
	key := kb.Build()

	// If key is longer than 200 chars, add a hash suffix
	if len(key) > 200 {
		hash := HashKey(key)
		return fmt.Sprintf("%s:%s", key[:150], hash[:16])
	}

	return key
}

// HashKey generates a SHA256 hash of the key
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Protocol-specific key builders

// GopherKey generates a cache key for Gopher responses
func GopherKey(selector string) string {
	return NewKeyBuilder().
		Add("gopher").
		Add(selector).
		BuildHashed()
}

// GeminiKey generates a cache key for Gemini responses
func GeminiKey(path string, query string) string {
	kb := NewKeyBuilder().
		Add("gemini").
		Add(path)

	if query != "" {
		kb.Add("q").Add(query)
	}

	return kb.BuildHashed()
}

// FingerKey generates a cache key for Finger responses
func FingerKey(username string) string {
	return NewKeyBuilder().
		Add("finger").
		Add(username).
		Build()
}

// EventKey generates a cache key for event rendering
func EventKey(eventID string, protocol string, format string) string {
	return NewKeyBuilder().
		Add("event").
		Add(eventID).
		Add(protocol).
		Add(format).
		Build()
}

// SectionKey generates a cache key for section rendering
func SectionKey(sectionName string, protocol string, page int) string {
	return NewKeyBuilder().
		Add("section").
		Add(sectionName).
		Add(protocol).
		Add(fmt.Sprintf("p%d", page)).
		Build()
}

// ThreadKey generates a cache key for thread rendering
func ThreadKey(rootEventID string, protocol string) string {
	return NewKeyBuilder().
		Add("thread").
		Add(rootEventID).
		Add(protocol).
		Build()
}

// ProfileKey generates a cache key for profile rendering
func ProfileKey(pubkey string, protocol string) string {
	return NewKeyBuilder().
		Add("profile").
		Add(pubkey).
		Add(protocol).
		Build()
}

// AggregateKey generates a cache key for aggregate data
func AggregateKey(eventID string) string {
	return NewKeyBuilder().
		Add("aggregate").
		Add(eventID).
		Build()
}

// Kind0Key generates a cache key for kind 0 (profile metadata) events
func Kind0Key(pubkey string) string {
	return NewKeyBuilder().
		Add("kind0").
		Add(pubkey).
		Build()
}

// Kind3Key generates a cache key for kind 3 (contact list) events
func Kind3Key(pubkey string) string {
	return NewKeyBuilder().
		Add("kind3").
		Add(pubkey).
		Build()
}

// ListKey generates a cache key for list queries
func ListKey(listType string, filters ...string) string {
	kb := NewKeyBuilder().
		Add("list").
		Add(listType)

	for _, filter := range filters {
		kb.Add(filter)
	}

	return kb.BuildHashed()
}

// Pattern generators for bulk operations

// GopherPattern returns a pattern for matching all Gopher keys
func GopherPattern() string {
	return "gopher:*"
}

// GeminiPattern returns a pattern for matching all Gemini keys
func GeminiPattern() string {
	return "gemini:*"
}

// FingerPattern returns a pattern for matching all Finger keys
func FingerPattern() string {
	return "finger:*"
}

// EventPattern returns a pattern for matching event keys
func EventPattern(eventID string) string {
	if eventID == "" {
		return "event:*"
	}
	return fmt.Sprintf("event:%s:*", eventID)
}

// ProfilePattern returns a pattern for matching profile keys
func ProfilePattern(pubkey string) string {
	if pubkey == "" {
		return "profile:*"
	}
	return fmt.Sprintf("profile:%s:*", pubkey)
}

// InvalidationPatterns returns all patterns that should be invalidated
// for a given event
func InvalidationPatterns(eventID string, kind int, pubkey string) []string {
	patterns := []string{
		EventPattern(eventID),
		AggregateKey(eventID),
	}

	// Add kind-specific patterns
	switch kind {
	case 0: // Profile metadata
		patterns = append(patterns,
			Kind0Key(pubkey),
			ProfilePattern(pubkey),
		)
	case 1: // Short text note
		patterns = append(patterns,
			SectionKey("notes", "*", 0),
		)
	case 3: // Contact list
		patterns = append(patterns,
			Kind3Key(pubkey),
		)
	case 7: // Reaction
		// Reactions invalidate the parent event's aggregates
		// Parent event ID would need to be extracted from tags
	case 9735: // Zap
		// Zaps invalidate the parent event's aggregates
		// Parent event ID would need to be extracted from tags
	}

	return patterns
}
