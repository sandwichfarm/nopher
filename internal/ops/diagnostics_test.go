package ops

import (
	"strings"
	"testing"
	"time"
)

func TestSystemStats(t *testing.T) {
	collector := &DiagnosticsCollector{
		version:   "test-version",
		commit:    "test-commit",
		startTime: time.Now().Add(-1 * time.Hour),
	}

	stats := collector.CollectSystemStats()

	if stats.Version != "test-version" {
		t.Errorf("expected version 'test-version', got '%s'", stats.Version)
	}

	if stats.Commit != "test-commit" {
		t.Errorf("expected commit 'test-commit', got '%s'", stats.Commit)
	}

	if stats.Uptime < time.Hour {
		t.Errorf("expected uptime >= 1h, got %v", stats.Uptime)
	}

	if stats.NumGoroutines == 0 {
		t.Error("expected goroutines > 0")
	}

	if stats.MemAllocMB == 0 {
		t.Error("expected memory allocated > 0")
	}
}

func TestDiagnosticsFormatAsText(t *testing.T) {
	diag := &Diagnostics{
		CollectedAt: time.Now(),
		System: &SystemStats{
			Version:       "v1.0.0",
			Commit:        "abc123",
			Uptime:        1 * time.Hour,
			GoVersion:     "go1.21",
			NumGoroutines: 10,
			MemAllocMB:    100.5,
			MemSysMB:      200.0,
			NumGC:          5,
		},
		Storage: &StorageStats{
			Driver:         "sqlite",
			TotalEvents:    1000,
			EventsByKind:   map[int]int64{1: 500, 3: 300, 7: 200},
			DatabaseSizeMB: 50.5,
		},
		Sync: &SyncStats{
			Enabled:         true,
			RelayCount:      3,
			ConnectedRelays: 2,
			TotalSynced:     1000,
		},
		Relays: []RelayHealth{
			{
				URL:          "wss://relay.test",
				Connected:    true,
				EventsSynced: 500,
			},
		},
		Aggregates: &AggregateStats{
			TotalAggregates: 800,
			ByKind:          map[int]int64{1: 500, 3: 300},
		},
	}

	text := diag.FormatAsText()

	// Check that all sections are present
	expectedSections := []string{
		"=== nophr Diagnostics ===",
		"--- System ---",
		"--- Storage ---",
		"--- Sync ---",
		"--- Relay Health ---",
		"--- Aggregates ---",
		"v1.0.0",
		"abc123",
		"sqlite",
		"1000",
		"wss://relay.test",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(text, expected) {
			t.Errorf("expected text to contain '%s', got: %s", expected, text)
		}
	}
}

func TestDiagnosticsFormatAsGemtext(t *testing.T) {
	diag := &Diagnostics{
		CollectedAt: time.Now(),
		System: &SystemStats{
			Version:       "v1.0.0",
			Commit:        "abc123",
			Uptime:        1 * time.Hour,
			GoVersion:     "go1.21",
			NumGoroutines: 10,
			MemAllocMB:    100.5,
		},
		Storage: &StorageStats{
			Driver:         "sqlite",
			TotalEvents:    1000,
			DatabaseSizeMB: 50.5,
		},
		Sync: &SyncStats{
			Enabled:         true,
			RelayCount:      3,
			ConnectedRelays: 2,
			TotalSynced:     1000,
		},
		Aggregates: &AggregateStats{
			TotalAggregates: 800,
		},
	}

	gemtext := diag.FormatAsGemtext()

	// Check for gemtext headings
	expectedHeadings := []string{
		"# nophr Diagnostics",
		"## System",
		"## Storage",
		"## Sync",
	}

	for _, expected := range expectedHeadings {
		if !strings.Contains(gemtext, expected) {
			t.Errorf("expected gemtext to contain '%s'", expected)
		}
	}

	// Check for bullet points
	if !strings.Contains(gemtext, "* Version:") {
		t.Error("expected gemtext to contain bullet points")
	}
}

func TestDiagnosticsFormatAsGophermap(t *testing.T) {
	diag := &Diagnostics{
		CollectedAt: time.Now(),
		System: &SystemStats{
			Version:    "v1.0.0",
			Uptime:     1 * time.Hour,
			MemAllocMB: 100.5,
		},
		Storage: &StorageStats{
			Driver:         "sqlite",
			TotalEvents:    1000,
			DatabaseSizeMB: 50.5,
		},
	}

	gophermap := diag.FormatAsGophermap("localhost", 70)

	// Check for gophermap lines
	if !strings.Contains(gophermap, "\t\tlocalhost\t70\r\n") {
		t.Error("expected gophermap format with tabs and CRLF")
	}

	if !strings.Contains(gophermap, "nophr Diagnostics") {
		t.Error("expected gophermap to contain title")
	}

	if !strings.Contains(gophermap, "v1.0.0") {
		t.Error("expected gophermap to contain version")
	}
}
