package security

import (
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func TestDenyList(t *testing.T) {
	t.Run("Basic deny list", func(t *testing.T) {
		dl := NewDenyList([]string{"pubkey1", "pubkey2"})

		if !dl.IsPubkeyDenied("pubkey1") {
			t.Error("expected pubkey1 to be denied")
		}

		if dl.IsPubkeyDenied("pubkey3") {
			t.Error("expected pubkey3 to not be denied")
		}
	})

	t.Run("Add and remove", func(t *testing.T) {
		dl := NewDenyList([]string{})

		dl.AddPubkey("pubkey1")
		if !dl.IsPubkeyDenied("pubkey1") {
			t.Error("expected pubkey1 to be denied after adding")
		}

		dl.RemovePubkey("pubkey1")
		if dl.IsPubkeyDenied("pubkey1") {
			t.Error("expected pubkey1 to not be denied after removing")
		}
	})

	t.Run("Filter events", func(t *testing.T) {
		dl := NewDenyList([]string{"denied"})

		events := []*nostr.Event{
			{PubKey: "allowed1"},
			{PubKey: "denied"},
			{PubKey: "allowed2"},
		}

		filtered := dl.FilterEvents(events)

		if len(filtered) != 2 {
			t.Errorf("expected 2 events, got %d", len(filtered))
		}

		for _, event := range filtered {
			if event.PubKey == "denied" {
				t.Error("denied pubkey should be filtered out")
			}
		}
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("Basic rate limiting", func(t *testing.T) {
		rl := NewRateLimiter(5, time.Second)
		defer rl.Close()

		clientID := "test-client"

		// Should allow first 5 requests
		for i := 0; i < 5; i++ {
			if !rl.Allow(clientID) {
				t.Errorf("request %d should be allowed", i+1)
			}
		}

		// Should deny 6th request
		if rl.Allow(clientID) {
			t.Error("6th request should be denied")
		}
	})

	t.Run("Token refill", func(t *testing.T) {
		rl := NewRateLimiter(2, 100*time.Millisecond)
		defer rl.Close()

		clientID := "test-client"

		// Use up tokens
		rl.Allow(clientID)
		rl.Allow(clientID)

		// Should be denied
		if rl.Allow(clientID) {
			t.Error("should be denied")
		}

		// Wait for refill
		time.Sleep(150 * time.Millisecond)

		// Should be allowed again
		if !rl.Allow(clientID) {
			t.Error("should be allowed after refill")
		}
	})

	t.Run("Multiple clients", func(t *testing.T) {
		rl := NewRateLimiter(2, time.Second)
		defer rl.Close()

		// Client 1 uses tokens
		rl.Allow("client1")
		rl.Allow("client1")

		// Client 2 should have own tokens
		if !rl.Allow("client2") {
			t.Error("client2 should have own token bucket")
		}
	})
}

func TestValidator(t *testing.T) {
	v := NewValidator()

	t.Run("Gopher selector validation", func(t *testing.T) {
		tests := []struct {
			selector string
			valid    bool
		}{
			{"/valid/selector", true},
			{"/selector\r\n", false},      // CRLF injection
			{"/../etc/passwd", false},     // Directory traversal
			{"/selector\x00", false},      // Null byte
			{"/normal", true},
		}

		for _, tt := range tests {
			err := v.ValidateGopherSelector(tt.selector)
			if tt.valid && err != nil {
				t.Errorf("selector '%s' should be valid, got error: %v", tt.selector, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("selector '%s' should be invalid", tt.selector)
			}
		}
	})

	t.Run("Pubkey validation", func(t *testing.T) {
		tests := []struct {
			pubkey string
			valid  bool
		}{
			{"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", true},
			{"invalid", false},
			{"0123456789abcdef", false}, // Too short
			{"0123456789abcdefg123456789abcdef0123456789abcdef0123456789abcdef", false}, // Invalid hex
		}

		for _, tt := range tests {
			err := v.ValidatePubkey(tt.pubkey)
			if tt.valid && err != nil {
				t.Errorf("pubkey should be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("pubkey should be invalid")
			}
		}
	})

	t.Run("Npub validation", func(t *testing.T) {
		tests := []struct {
			npub  string
			valid bool
		}{
			{"npub1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8t3jg9", true}, // Valid length
			{"invalid", false},
			{"nsec1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8t3jg9", false}, // Wrong prefix
		}

		for _, tt := range tests {
			err := v.ValidateNpub(tt.npub)
			if tt.valid && err != nil {
				t.Errorf("npub should be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("npub '%s' should be invalid", tt.npub)
			}
		}
	})

	t.Run("Integer validation", func(t *testing.T) {
		err := v.ValidateInteger(50, 0, 100)
		if err != nil {
			t.Errorf("50 should be valid in range 0-100: %v", err)
		}

		err = v.ValidateInteger(150, 0, 100)
		if err == nil {
			t.Error("150 should be invalid in range 0-100")
		}
	})

	t.Run("Sanitization", func(t *testing.T) {
		input := "test\r\n\x00"
		sanitized := v.SanitizeInput(input)

		if sanitized != "test" {
			t.Errorf("expected 'test', got '%s'", sanitized)
		}
	})
}

func TestSecretManager(t *testing.T) {
	t.Run("Basic secret management", func(t *testing.T) {
		sm := NewSecretManager()

		sm.Set("test-key", "test-value")

		value, exists := sm.Get("test-key")
		if !exists {
			t.Error("expected key to exist")
		}

		if value != "test-value" {
			t.Errorf("expected 'test-value', got '%s'", value)
		}
	})

	t.Run("Redaction", func(t *testing.T) {
		sm := NewSecretManager()

		redacted := sm.Redact("secret1234567890")
		if redacted == "secret1234567890" {
			t.Error("expected secret to be redacted")
		}

		if !contains(redacted, "...") {
			t.Error("expected redacted format with '...'")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		sm := NewSecretManager()
		sm.Set("key1", "value1")
		sm.Set("key2", "value2")

		sm.Clear()

		_, exists := sm.Get("key1")
		if exists {
			t.Error("expected secrets to be cleared")
		}
	})
}

func TestSecureString(t *testing.T) {
	t.Run("Secure string redaction", func(t *testing.T) {
		ss := NewSecureString("secret1234567890")

		// String() should return redacted version
		str := ss.String()
		if str == "secret1234567890" {
			t.Error("expected string to be redacted")
		}

		// Get() should return actual value
		if ss.Get() != "secret1234567890" {
			t.Error("Get() should return actual value")
		}
	})
}

func TestSecretValidator(t *testing.T) {
	sv := NewSecretValidator()

	t.Run("Nsec validation", func(t *testing.T) {
		valid := "nsec1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8t3jg9"
		invalid := "npub1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8t3jg9"

		if err := sv.ValidateNsec(valid); err != nil {
			t.Errorf("valid nsec should pass: %v", err)
		}

		if err := sv.ValidateNsec(invalid); err == nil {
			t.Error("invalid nsec should fail")
		}
	})

	t.Run("Secret leak detection", func(t *testing.T) {
		leaks := sv.CheckForLeakedSecrets("This contains nsec1234567890")

		if len(leaks) == 0 {
			t.Error("expected leak detection")
		}
	})
}

func TestContentFilter(t *testing.T) {
	cf := NewContentFilter([]string{"spam", "bad"})

	t.Run("Banned content detection", func(t *testing.T) {
		if !cf.ContainsBannedContent("this is spam") {
			t.Error("should detect spam")
		}

		if cf.ContainsBannedContent("this is good") {
			t.Error("should not detect good content as bad")
		}
	})

	t.Run("Event filtering", func(t *testing.T) {
		event := &nostr.Event{
			Content: "this is spam",
		}

		if !cf.IsEventFiltered(event) {
			t.Error("event with banned content should be filtered")
		}
	})
}

func TestCombinedFilter(t *testing.T) {
	dl := NewDenyList([]string{"badpubkey"})
	cf := NewContentFilter([]string{"spam"})
	combined := NewCombinedFilter(dl, cf)

	t.Run("Combined filtering", func(t *testing.T) {
		// Event with denied pubkey
		event1 := &nostr.Event{
			PubKey:  "badpubkey",
			Content: "good content",
		}

		if combined.IsEventAllowed(event1) {
			t.Error("event from denied pubkey should not be allowed")
		}

		// Event with banned content
		event2 := &nostr.Event{
			PubKey:  "goodpubkey",
			Content: "this is spam",
		}

		if combined.IsEventAllowed(event2) {
			t.Error("event with banned content should not be allowed")
		}

		// Good event
		event3 := &nostr.Event{
			PubKey:  "goodpubkey",
			Content: "good content",
		}

		if !combined.IsEventAllowed(event3) {
			t.Error("good event should be allowed")
		}
	})
}
