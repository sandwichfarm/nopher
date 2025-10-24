package config

import (
	"testing"
)

func TestSyncKindsToIntSlice(t *testing.T) {
	tests := []struct {
		name     string
		kinds    SyncKinds
		expected []int
	}{
		{
			name:     "empty sync kinds",
			kinds:    SyncKinds{},
			expected: []int{},
		},
		{
			name: "all kinds enabled",
			kinds: SyncKinds{
				Profiles:    true,
				Notes:       true,
				ContactList: true,
				Reposts:     true,
				Reactions:   true,
				Zaps:        true,
				Articles:    true,
				RelayList:   true,
			},
			expected: []int{0, 1, 3, 6, 7, 9735, 30023, 10002},
		},
		{
			name: "only notes and reactions",
			kinds: SyncKinds{
				Notes:     true,
				Reactions: true,
			},
			expected: []int{1, 7},
		},
		{
			name: "only profiles",
			kinds: SyncKinds{
				Profiles: true,
			},
			expected: []int{0},
		},
		{
			name: "with allowlist",
			kinds: SyncKinds{
				Notes:     true,
				Allowlist: []int{100, 200},
			},
			expected: []int{1, 100, 200},
		},
		{
			name: "only allowlist",
			kinds: SyncKinds{
				Allowlist: []int{42, 1337},
			},
			expected: []int{42, 1337},
		},
		{
			name: "profiles and contact list (common setup)",
			kinds: SyncKinds{
				Profiles:    true,
				ContactList: true,
			},
			expected: []int{0, 3},
		},
		{
			name: "social kinds (notes, reactions, zaps)",
			kinds: SyncKinds{
				Notes:     true,
				Reactions: true,
				Zaps:      true,
			},
			expected: []int{1, 7, 9735},
		},
		{
			name: "long-form content",
			kinds: SyncKinds{
				Articles: true,
			},
			expected: []int{30023},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.kinds.ToIntSlice()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d kinds, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Check each expected kind is in result
			for _, expectedKind := range tt.expected {
				found := false
				for _, resultKind := range result {
					if resultKind == expectedKind {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected kind %d not found in result: %v", expectedKind, result)
				}
			}
		})
	}
}

func TestSyncKindsOrder(t *testing.T) {
	// Test that kinds are returned in expected order
	kinds := SyncKinds{
		Articles:    true,
		Notes:       true,
		Profiles:    true,
		Reactions:   true,
		ContactList: true,
		Reposts:     true,
		Zaps:        true,
		RelayList:   true,
	}

	result := kinds.ToIntSlice()
	expected := []int{0, 1, 3, 6, 7, 9735, 30023, 10002}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d kinds, got %d", len(expected), len(result))
	}

	for i, kind := range expected {
		if result[i] != kind {
			t.Errorf("Expected kind at index %d to be %d, got %d", i, kind, result[i])
		}
	}
}

func TestSyncKindsAllowlistOrder(t *testing.T) {
	// Test that allowlist kinds appear after standard kinds
	kinds := SyncKinds{
		Notes:     true,
		Reactions: true,
		Allowlist: []int{1000, 2000, 3000},
	}

	result := kinds.ToIntSlice()

	// First two should be standard kinds
	if result[0] != 1 || result[1] != 7 {
		t.Errorf("Expected standard kinds first, got: %v", result[:2])
	}

	// Last three should be allowlist kinds
	allowlistStart := len(result) - 3
	expectedAllowlist := []int{1000, 2000, 3000}
	for i, kind := range expectedAllowlist {
		if result[allowlistStart+i] != kind {
			t.Errorf("Expected allowlist kind at index %d to be %d, got %d",
				allowlistStart+i, kind, result[allowlistStart+i])
		}
	}
}

func TestSyncKindsIndividualFlags(t *testing.T) {
	// Test each flag individually
	tests := []struct {
		name         string
		setup        func(*SyncKinds)
		expectedKind int
	}{
		{
			name:         "Profiles",
			setup:        func(sk *SyncKinds) { sk.Profiles = true },
			expectedKind: 0,
		},
		{
			name:         "Notes",
			setup:        func(sk *SyncKinds) { sk.Notes = true },
			expectedKind: 1,
		},
		{
			name:         "ContactList",
			setup:        func(sk *SyncKinds) { sk.ContactList = true },
			expectedKind: 3,
		},
		{
			name:         "Reposts",
			setup:        func(sk *SyncKinds) { sk.Reposts = true },
			expectedKind: 6,
		},
		{
			name:         "Reactions",
			setup:        func(sk *SyncKinds) { sk.Reactions = true },
			expectedKind: 7,
		},
		{
			name:         "Zaps",
			setup:        func(sk *SyncKinds) { sk.Zaps = true },
			expectedKind: 9735,
		},
		{
			name:         "Articles",
			setup:        func(sk *SyncKinds) { sk.Articles = true },
			expectedKind: 30023,
		},
		{
			name:         "RelayList",
			setup:        func(sk *SyncKinds) { sk.RelayList = true },
			expectedKind: 10002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kinds := SyncKinds{}
			tt.setup(&kinds)

			result := kinds.ToIntSlice()

			if len(result) != 1 {
				t.Errorf("Expected exactly 1 kind, got %d: %v", len(result), result)
				return
			}

			if result[0] != tt.expectedKind {
				t.Errorf("Expected kind %d, got %d", tt.expectedKind, result[0])
			}
		})
	}
}

func TestSyncKindsNoDuplicates(t *testing.T) {
	// Ensure no duplicates even with allowlist
	kinds := SyncKinds{
		Notes:     true,
		Allowlist: []int{1, 7, 100}, // 1 is duplicate with Notes
	}

	result := kinds.ToIntSlice()

	// Check for duplicates
	seen := make(map[int]bool)
	for _, kind := range result {
		if seen[kind] {
			t.Errorf("Duplicate kind found: %d", kind)
		}
		seen[kind] = true
	}

	// Should have 4 unique kinds: 1 (Notes), 1 (duplicate from allowlist), 7, 100
	// Actually, based on implementation, duplicates might be allowed
	// Let's just verify the function works consistently
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}
