package cache

import (
	"strings"
	"testing"
)

func TestKeyBuilder(t *testing.T) {
	t.Run("Simple key", func(t *testing.T) {
		key := NewKeyBuilder().
			Add("protocol").
			Add("selector").
			Build()

		expected := "protocol:selector"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})

	t.Run("Hashed key for long strings", func(t *testing.T) {
		longPart := strings.Repeat("a", 300)
		key := NewKeyBuilder().
			Add("protocol").
			Add(longPart).
			BuildHashed()

		// Should be truncated and hashed
		if len(key) > 200 {
			t.Errorf("expected key length <= 200, got %d", len(key))
		}

		// Should contain a hash
		if !strings.Contains(key, ":") {
			t.Error("expected key to contain hash separator")
		}
	})
}

func TestProtocolKeys(t *testing.T) {
	t.Run("Gopher key", func(t *testing.T) {
		key := GopherKey("/test/selector")
		if !strings.HasPrefix(key, "gopher:") {
			t.Errorf("expected gopher prefix, got %s", key)
		}
	})

	t.Run("Gemini key", func(t *testing.T) {
		key := GeminiKey("/test/path", "query=test")
		if !strings.HasPrefix(key, "gemini:") {
			t.Errorf("expected gemini prefix, got %s", key)
		}
		if !strings.Contains(key, "q:query=test") {
			t.Errorf("expected query in key, got %s", key)
		}
	})

	t.Run("Gemini key without query", func(t *testing.T) {
		key := GeminiKey("/test/path", "")
		if !strings.HasPrefix(key, "gemini:") {
			t.Errorf("expected gemini prefix, got %s", key)
		}
		if strings.Contains(key, ":q:") {
			t.Errorf("expected no query marker in key, got %s", key)
		}
	})

	t.Run("Finger key", func(t *testing.T) {
		key := FingerKey("user")
		expected := "finger:user"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})
}

func TestEventKeys(t *testing.T) {
	t.Run("Event key", func(t *testing.T) {
		key := EventKey("event123", "gopher", "text")
		expected := "event:event123:gopher:text"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})

	t.Run("Section key", func(t *testing.T) {
		key := SectionKey("notes", "gemini", 2)
		expected := "section:notes:gemini:p2"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})

	t.Run("Thread key", func(t *testing.T) {
		key := ThreadKey("root123", "gopher")
		expected := "thread:root123:gopher"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})

	t.Run("Profile key", func(t *testing.T) {
		key := ProfileKey("pubkey123", "gemini")
		expected := "profile:pubkey123:gemini"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})
}

func TestAggregateKeys(t *testing.T) {
	t.Run("Aggregate key", func(t *testing.T) {
		key := AggregateKey("event123")
		expected := "aggregate:event123"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})

	t.Run("Kind 0 key", func(t *testing.T) {
		key := Kind0Key("pubkey123")
		expected := "kind0:pubkey123"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})

	t.Run("Kind 3 key", func(t *testing.T) {
		key := Kind3Key("pubkey123")
		expected := "kind3:pubkey123"
		if key != expected {
			t.Errorf("expected %s, got %s", expected, key)
		}
	})
}

func TestPatterns(t *testing.T) {
	t.Run("Protocol patterns", func(t *testing.T) {
		tests := []struct {
			name     string
			pattern  string
			expected string
		}{
			{"Gopher", GopherPattern(), "gopher:*"},
			{"Gemini", GeminiPattern(), "gemini:*"},
			{"Finger", FingerPattern(), "finger:*"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.pattern != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, tt.pattern)
				}
			})
		}
	})

	t.Run("Event pattern", func(t *testing.T) {
		pattern := EventPattern("event123")
		expected := "event:event123:*"
		if pattern != expected {
			t.Errorf("expected %s, got %s", expected, pattern)
		}
	})

	t.Run("Event pattern all", func(t *testing.T) {
		pattern := EventPattern("")
		expected := "event:*"
		if pattern != expected {
			t.Errorf("expected %s, got %s", expected, pattern)
		}
	})

	t.Run("Profile pattern", func(t *testing.T) {
		pattern := ProfilePattern("pubkey123")
		expected := "profile:pubkey123:*"
		if pattern != expected {
			t.Errorf("expected %s, got %s", expected, pattern)
		}
	})
}

func TestInvalidationPatterns(t *testing.T) {
	t.Run("Kind 0 event", func(t *testing.T) {
		patterns := InvalidationPatterns("event123", 0, "pubkey123")

		// Should include event, aggregate, kind0, and profile patterns
		expectedPatterns := []string{
			"event:event123:*",
			"aggregate:event123",
			"kind0:pubkey123",
			"profile:pubkey123:*",
		}

		for _, expected := range expectedPatterns {
			found := false
			for _, pattern := range patterns {
				if pattern == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected pattern %s not found in %v", expected, patterns)
			}
		}
	})

	t.Run("Kind 1 event", func(t *testing.T) {
		patterns := InvalidationPatterns("event123", 1, "pubkey123")

		// Should include section invalidation
		found := false
		for _, pattern := range patterns {
			if strings.Contains(pattern, "section:notes") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected notes section invalidation pattern")
		}
	})

	t.Run("Kind 3 event", func(t *testing.T) {
		patterns := InvalidationPatterns("event123", 3, "pubkey123")

		// Should include kind3 pattern
		found := false
		for _, pattern := range patterns {
			if pattern == "kind3:pubkey123" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected kind3 pattern")
		}
	})
}

func TestHashKey(t *testing.T) {
	t.Run("Consistent hashing", func(t *testing.T) {
		key := "test-key"
		hash1 := HashKey(key)
		hash2 := HashKey(key)

		if hash1 != hash2 {
			t.Error("expected consistent hash for same key")
		}
	})

	t.Run("Different keys produce different hashes", func(t *testing.T) {
		hash1 := HashKey("key1")
		hash2 := HashKey("key2")

		if hash1 == hash2 {
			t.Error("expected different hashes for different keys")
		}
	})

	t.Run("Hash length", func(t *testing.T) {
		hash := HashKey("test")

		// SHA256 hex encoding is 64 characters
		if len(hash) != 64 {
			t.Errorf("expected hash length 64, got %d", len(hash))
		}
	})
}
