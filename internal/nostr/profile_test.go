package nostr

import (
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

func TestParseProfile(t *testing.T) {
	tests := []struct {
		name     string
		event    *nostr.Event
		expected *ProfileMetadata
	}{
		{
			name: "complete profile",
			event: &nostr.Event{
				Kind: 0,
				Content: `{
					"name": "Alice",
					"display_name": "Alice Wonderland",
					"about": "I love cryptography",
					"picture": "https://example.com/alice.jpg",
					"banner": "https://example.com/banner.jpg",
					"website": "https://alice.example.com",
					"nip05": "alice@example.com",
					"lud16": "alice@getalby.com"
				}`,
			},
			expected: &ProfileMetadata{
				Name:        "Alice",
				DisplayName: "Alice Wonderland",
				About:       "I love cryptography",
				Picture:     "https://example.com/alice.jpg",
				Banner:      "https://example.com/banner.jpg",
				Website:     "https://alice.example.com",
				NIP05:       "alice@example.com",
				LUD16:       "alice@getalby.com",
			},
		},
		{
			name: "minimal profile with only name",
			event: &nostr.Event{
				Kind:    0,
				Content: `{"name": "Bob"}`,
			},
			expected: &ProfileMetadata{
				Name: "Bob",
			},
		},
		{
			name: "empty profile",
			event: &nostr.Event{
				Kind:    0,
				Content: `{}`,
			},
			expected: &ProfileMetadata{},
		},
		{
			name:     "nil event",
			event:    nil,
			expected: nil,
		},
		{
			name: "wrong kind",
			event: &nostr.Event{
				Kind:    1,
				Content: `{"name": "Alice"}`,
			},
			expected: nil,
		},
		{
			name: "invalid JSON",
			event: &nostr.Event{
				Kind:    0,
				Content: `{invalid json`,
			},
			expected: &ProfileMetadata{}, // Returns empty on parse error
		},
		{
			name: "profile with lud06",
			event: &nostr.Event{
				Kind: 0,
				Content: `{
					"name": "Charlie",
					"lud06": "LNURL1234567890"
				}`,
			},
			expected: &ProfileMetadata{
				Name:  "Charlie",
				LUD06: "LNURL1234567890",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseProfile(tt.event)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Name != tt.expected.Name {
				t.Errorf("Name: expected %q, got %q", tt.expected.Name, result.Name)
			}
			if result.DisplayName != tt.expected.DisplayName {
				t.Errorf("DisplayName: expected %q, got %q", tt.expected.DisplayName, result.DisplayName)
			}
			if result.About != tt.expected.About {
				t.Errorf("About: expected %q, got %q", tt.expected.About, result.About)
			}
			if result.Picture != tt.expected.Picture {
				t.Errorf("Picture: expected %q, got %q", tt.expected.Picture, result.Picture)
			}
			if result.Banner != tt.expected.Banner {
				t.Errorf("Banner: expected %q, got %q", tt.expected.Banner, result.Banner)
			}
			if result.Website != tt.expected.Website {
				t.Errorf("Website: expected %q, got %q", tt.expected.Website, result.Website)
			}
			if result.NIP05 != tt.expected.NIP05 {
				t.Errorf("NIP05: expected %q, got %q", tt.expected.NIP05, result.NIP05)
			}
			if result.LUD16 != tt.expected.LUD16 {
				t.Errorf("LUD16: expected %q, got %q", tt.expected.LUD16, result.LUD16)
			}
			if result.LUD06 != tt.expected.LUD06 {
				t.Errorf("LUD06: expected %q, got %q", tt.expected.LUD06, result.LUD06)
			}
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		profile  *ProfileMetadata
		expected string
	}{
		{
			name: "has display_name",
			profile: &ProfileMetadata{
				Name:        "alice",
				DisplayName: "Alice Wonderland",
			},
			expected: "Alice Wonderland",
		},
		{
			name: "only has name",
			profile: &ProfileMetadata{
				Name: "bob",
			},
			expected: "bob",
		},
		{
			name:     "empty profile",
			profile:  &ProfileMetadata{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.GetDisplayName()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetLightningAddress(t *testing.T) {
	tests := []struct {
		name     string
		profile  *ProfileMetadata
		expected string
	}{
		{
			name: "has lud16",
			profile: &ProfileMetadata{
				LUD16: "alice@getalby.com",
				LUD06: "LNURL1234",
			},
			expected: "alice@getalby.com",
		},
		{
			name: "only has lud06",
			profile: &ProfileMetadata{
				LUD06: "LNURL1234",
			},
			expected: "LNURL1234",
		},
		{
			name:     "empty profile",
			profile:  &ProfileMetadata{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.GetLightningAddress()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHasAnyField(t *testing.T) {
	tests := []struct {
		name     string
		profile  *ProfileMetadata
		expected bool
	}{
		{
			name: "has name",
			profile: &ProfileMetadata{
				Name: "Alice",
			},
			expected: true,
		},
		{
			name: "has about",
			profile: &ProfileMetadata{
				About: "I love cryptography",
			},
			expected: true,
		},
		{
			name:     "empty profile",
			profile:  &ProfileMetadata{},
			expected: false,
		},
		{
			name: "has multiple fields",
			profile: &ProfileMetadata{
				Name:    "Bob",
				Website: "https://bob.com",
				NIP05:   "bob@example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.HasAnyField()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
