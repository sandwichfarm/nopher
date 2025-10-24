package sections

import (
	"testing"
)

func TestDefaultSections(t *testing.T) {
	sections := DefaultSections()

	if len(sections) == 0 {
		t.Fatal("expected default sections, got none")
	}

	expectedSections := map[string]bool{
		"notes":     false,
		"articles":  false,
		"reactions": false,
		"zaps":      false,
	}

	for _, section := range sections {
		if _, exists := expectedSections[section.Name]; exists {
			expectedSections[section.Name] = true
		}
	}

	for name, found := range expectedSections {
		if !found {
			t.Errorf("expected default section %s not found", name)
		}
	}
}

func TestInboxSection(t *testing.T) {
	pubkey := "testpubkey123"
	section := InboxSection(pubkey)

	if section.Name != "inbox" {
		t.Errorf("expected name 'inbox', got '%s'", section.Name)
	}

	if len(section.Filters.Tags) == 0 {
		t.Error("expected tag filters for inbox")
	}

	pTags, exists := section.Filters.Tags["p"]
	if !exists {
		t.Error("expected 'p' tag filter")
	}

	if len(pTags) != 1 || pTags[0] != pubkey {
		t.Errorf("expected p tag with pubkey %s", pubkey)
	}
}

func TestOutboxSection(t *testing.T) {
	pubkey := "testpubkey123"
	section := OutboxSection(pubkey)

	if section.Name != "outbox" {
		t.Errorf("expected name 'outbox', got '%s'", section.Name)
	}

	if len(section.Filters.Authors) != 1 || section.Filters.Authors[0] != pubkey {
		t.Errorf("expected author filter with pubkey %s", pubkey)
	}

	if section.ShowAuthors {
		t.Error("expected ShowAuthors to be false for outbox")
	}
}

func TestSectionManager(t *testing.T) {
	manager := NewManager(nil)

	t.Run("Register section", func(t *testing.T) {
		section := &Section{
			Name:  "test-section",
			Title: "Test Section",
			Limit: 10,
		}

		err := manager.RegisterSection(section)
		if err != nil {
			t.Fatalf("failed to register section: %v", err)
		}

		retrieved, err := manager.GetSection("test-section")
		if err != nil {
			t.Fatalf("failed to get section: %v", err)
		}

		if retrieved.Name != "test-section" {
			t.Errorf("expected name 'test-section', got '%s'", retrieved.Name)
		}
	})

	t.Run("Get nonexistent section", func(t *testing.T) {
		_, err := manager.GetSection("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent section")
		}
	})

	t.Run("List sections", func(t *testing.T) {
		sections := manager.ListSections()
		if len(sections) == 0 {
			t.Error("expected at least one section")
		}
	})

	t.Run("Register section without name", func(t *testing.T) {
		section := &Section{
			Title: "No Name",
		}

		err := manager.RegisterSection(section)
		if err == nil {
			t.Error("expected error when registering section without name")
		}
	})

	t.Run("Default limit", func(t *testing.T) {
		section := &Section{
			Name:  "no-limit",
			Title: "No Limit Section",
		}

		manager.RegisterSection(section)

		retrieved, _ := manager.GetSection("no-limit")
		if retrieved.Limit != 20 {
			t.Errorf("expected default limit 20, got %d", retrieved.Limit)
		}
	})
}

func TestArchiveFormatting(t *testing.T) {
	t.Run("Day archive", func(t *testing.T) {
		archive := &Archive{
			Period: ArchiveByDay,
			Year:   2025,
			Month:  10,
			Day:    24,
		}

		title := archive.FormatTitle()
		if title != "October 24, 2025" {
			t.Errorf("unexpected day archive title: %s", title)
		}

		selector := archive.FormatArchiveSelector("notes")
		expected := "/archive/notes/2025/10/24"
		if selector != expected {
			t.Errorf("expected selector %s, got %s", expected, selector)
		}
	})

	t.Run("Month archive", func(t *testing.T) {
		archive := &Archive{
			Period: ArchiveByMonth,
			Year:   2025,
			Month:  10,
		}

		title := archive.FormatTitle()
		if title != "October 2025" {
			t.Errorf("unexpected month archive title: %s", title)
		}

		selector := archive.FormatArchiveSelector("notes")
		expected := "/archive/notes/2025/10"
		if selector != expected {
			t.Errorf("expected selector %s, got %s", expected, selector)
		}
	})

	t.Run("Year archive", func(t *testing.T) {
		archive := &Archive{
			Period: ArchiveByYear,
			Year:   2025,
		}

		title := archive.FormatTitle()
		if title != "2025" {
			t.Errorf("unexpected year archive title: %s", title)
		}

		selector := archive.FormatArchiveSelector("notes")
		expected := "/archive/notes/2025"
		if selector != expected {
			t.Errorf("expected selector %s, got %s", expected, selector)
		}
	})
}
