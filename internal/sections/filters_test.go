package sections

import (
	"testing"
	"time"
)

func TestFilterBuilder(t *testing.T) {
	t.Run("Basic filter", func(t *testing.T) {
		filter := NewFilterBuilder().
			Kinds(1, 30023).
			Authors("pubkey1", "pubkey2").
			Limit(20).
			Build()

		if len(filter.Kinds) != 2 {
			t.Errorf("expected 2 kinds, got %d", len(filter.Kinds))
		}

		if len(filter.Authors) != 2 {
			t.Errorf("expected 2 authors, got %d", len(filter.Authors))
		}

		if filter.Limit != 20 {
			t.Errorf("expected limit 20, got %d", filter.Limit)
		}
	})

	t.Run("Tag filter", func(t *testing.T) {
		filter := NewFilterBuilder().
			Tag("e", "event123").
			Tag("p", "pubkey1", "pubkey2").
			Build()

		if len(filter.Tags) != 2 {
			t.Errorf("expected 2 tag keys, got %d", len(filter.Tags))
		}

		eTags := filter.Tags["e"]
		if len(eTags) != 1 || eTags[0] != "event123" {
			t.Error("unexpected e tag values")
		}

		pTags := filter.Tags["p"]
		if len(pTags) != 2 {
			t.Errorf("expected 2 p tag values, got %d", len(pTags))
		}
	})

	t.Run("Time range filter", func(t *testing.T) {
		since := time.Now().Add(-24 * time.Hour)
		until := time.Now()

		filter := NewFilterBuilder().
			Since(since).
			Until(until).
			Build()

		if filter.Since == nil {
			t.Error("expected since to be set")
		}

		if filter.Until == nil {
			t.Error("expected until to be set")
		}
	})
}

func TestTimeRangeFilters(t *testing.T) {
	t.Run("Today", func(t *testing.T) {
		trf := Today()

		if trf.Start.IsZero() || trf.End.IsZero() {
			t.Error("expected start and end to be set")
		}

		duration := trf.End.Sub(trf.Start)
		if duration != 24*time.Hour {
			t.Errorf("expected 24 hour duration, got %v", duration)
		}
	})

	t.Run("This week", func(t *testing.T) {
		trf := ThisWeek()

		duration := trf.End.Sub(trf.Start)
		if duration != 7*24*time.Hour {
			t.Errorf("expected 7 day duration, got %v", duration)
		}
	})

	t.Run("This month", func(t *testing.T) {
		trf := ThisMonth()

		if trf.Start.Day() != 1 {
			t.Errorf("expected month start on day 1, got %d", trf.Start.Day())
		}
	})

	t.Run("Last N days", func(t *testing.T) {
		trf := LastNDays(7)

		duration := trf.End.Sub(trf.Start)
		expectedDuration := 7 * 24 * time.Hour

		// Allow small variance for test execution time
		diff := duration - expectedDuration
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Minute {
			t.Errorf("expected ~7 day duration, got %v", duration)
		}
	})

	t.Run("Last N hours", func(t *testing.T) {
		trf := LastNHours(12)

		duration := trf.End.Sub(trf.Start)
		expectedDuration := 12 * time.Hour

		diff := duration - expectedDuration
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Minute {
			t.Errorf("expected ~12 hour duration, got %v", duration)
		}
	})
}

func TestInteractionFilters(t *testing.T) {
	t.Run("Interaction filter", func(t *testing.T) {
		filter := InteractionFilter("event123").Build()

		if len(filter.Kinds) != 3 {
			t.Errorf("expected 3 kinds (reactions, zaps, replies), got %d", len(filter.Kinds))
		}

		eTags := filter.Tags["e"]
		if len(eTags) != 1 || eTags[0] != "event123" {
			t.Error("expected e tag with event123")
		}
	})

	t.Run("Thread filter", func(t *testing.T) {
		filter := ThreadFilter("root123").Build()

		if len(filter.Kinds) != 1 || filter.Kinds[0] != 1 {
			t.Error("expected kind 1 for thread filter")
		}

		eTags := filter.Tags["e"]
		if len(eTags) != 1 || eTags[0] != "root123" {
			t.Error("expected e tag with root123")
		}
	})

	t.Run("Mention filter", func(t *testing.T) {
		filter := MentionFilter("pubkey123").Build()

		pTags := filter.Tags["p"]
		if len(pTags) != 1 || pTags[0] != "pubkey123" {
			t.Error("expected p tag with pubkey123")
		}
	})
}

func TestKindFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   KindFilter
		expected int
	}{
		{"Notes", KindFilterNotes, 1},
		{"Articles", KindFilterArticles, 30023},
		{"Reactions", KindFilterReactions, 7},
		{"Zaps", KindFilterZaps, 9735},
		{"Profiles", KindFilterProfiles, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.filter.Apply(NewFilterBuilder()).Build()

			if len(filter.Kinds) != 1 {
				t.Errorf("expected 1 kind, got %d", len(filter.Kinds))
			}

			if filter.Kinds[0] != tt.expected {
				t.Errorf("expected kind %d, got %d", tt.expected, filter.Kinds[0])
			}
		})
	}
}

func TestScopeFilterBuilder(t *testing.T) {
	pubkey := "testpubkey"

	tests := []struct {
		name          string
		scope         Scope
		expectedCount int
	}{
		{"Self", ScopeSelf, 1},
		{"Following", ScopeFollowing, 1}, // Simplified - would be more in real implementation
		{"All", ScopeAll, 0},              // No author filter for "all"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfb := NewScopeFilterBuilder(pubkey, tt.scope, 2)
			authors, err := sfb.BuildAuthors()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(authors) != tt.expectedCount {
				t.Errorf("expected %d authors, got %d", tt.expectedCount, len(authors))
			}
		})
	}
}
