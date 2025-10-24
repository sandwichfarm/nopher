package gopher

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/sandwich/nopher/internal/aggregates"
	"github.com/sandwich/nopher/internal/config"
	"github.com/sandwich/nopher/internal/storage"
)

func TestGopherProtocol(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Identity: config.Identity{
			Npub: "test-pubkey",
		},
		Storage: config.Storage{
			Driver:     "sqlite",
			SQLitePath: ":memory:",
		},
	}

	gopherCfg := &config.GopherProtocol{
		Enabled: true,
		Host:    "localhost",
		Port:    17070, // Use non-standard port for testing
	}

	// Create storage
	ctx := context.Background()
	st, err := storage.New(ctx, &cfg.Storage)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer st.Close()

	// Create aggregates manager
	aggMgr := aggregates.NewManager(st, cfg)

	// Create server
	server := New(gopherCfg, cfg, st, "localhost", aggMgr)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test 1: Empty selector (root)
	t.Run("RootSelector", func(t *testing.T) {
		response := sendGopherRequest(t, gopherCfg.Port, "")
		if !strings.Contains(response, "Nopher") {
			t.Errorf("Root response should contain 'Nopher', got: %s", response)
		}
		if !strings.HasSuffix(response, ".\r\n") {
			t.Errorf("Response should end with gopher terminator '.\\r\\n'")
		}
	})

	// Test 2: Outbox selector
	t.Run("OutboxSelector", func(t *testing.T) {
		response := sendGopherRequest(t, gopherCfg.Port, "/outbox")
		if !strings.Contains(response, "Outbox") {
			t.Errorf("Outbox response should contain 'Outbox', got: %s", response)
		}
	})

	// Test 3: Inbox selector
	t.Run("InboxSelector", func(t *testing.T) {
		response := sendGopherRequest(t, gopherCfg.Port, "/inbox")
		if !strings.Contains(response, "Inbox") {
			t.Errorf("Inbox response should contain 'Inbox', got: %s", response)
		}
	})

	// Test 4: Diagnostics selector
	t.Run("DiagnosticsSelector", func(t *testing.T) {
		response := sendGopherRequest(t, gopherCfg.Port, "/diagnostics")
		if !strings.Contains(response, "Diagnostics") {
			t.Errorf("Diagnostics response should contain 'Diagnostics', got: %s", response)
		}
		if !strings.Contains(response, "localhost") {
			t.Errorf("Diagnostics should contain hostname")
		}
	})

	// Test 5: Invalid selector
	t.Run("InvalidSelector", func(t *testing.T) {
		response := sendGopherRequest(t, gopherCfg.Port, "/invalid")
		if !strings.Contains(response, "3") || !strings.Contains(response, "Unknown") {
			t.Errorf("Invalid selector should return error type (3), got: %s", response)
		}
	})
}

func TestGophermapFormat(t *testing.T) {
	gmap := NewGophermap("localhost", 70)

	// Test adding different item types
	gmap.AddInfo("Test info")
	gmap.AddDirectory("Test directory", "/test")
	gmap.AddTextFile("Test file", "/test.txt")
	gmap.AddError("Test error")

	result := gmap.String()

	// Check basic structure
	if !strings.HasSuffix(result, ".\r\n") {
		t.Errorf("Gophermap should end with '.\\r\\n'")
	}

	// Check item types are present
	if !strings.Contains(result, "iTest info") {
		t.Errorf("Should contain info item (type 'i')")
	}
	if !strings.Contains(result, "1Test directory") {
		t.Errorf("Should contain directory item (type '1')")
	}
	if !strings.Contains(result, "0Test file") {
		t.Errorf("Should contain text file item (type '0')")
	}
	if !strings.Contains(result, "3Test error") {
		t.Errorf("Should contain error item (type '3')")
	}

	// Check TAB separators
	lines := strings.Split(result, "\r\n")
	for _, line := range lines {
		if len(line) > 0 && line != "." {
			// Each line should have exactly 4 TAB-separated parts (after type char)
			tabCount := strings.Count(line, "\t")
			if tabCount != 3 {
				t.Errorf("Line should have 3 TABs, got %d: %s", tabCount, line)
			}
		}
	}
}

func TestRendererOutput(t *testing.T) {
	renderer := NewRenderer()

	// Test note list rendering
	notes := []*aggregates.EnrichedEvent{}
	lines := renderer.RenderNoteList(notes, "Test List")

	if len(lines) == 0 {
		t.Errorf("RenderNoteList should return lines")
	}
	if lines[0] != "Test List" {
		t.Errorf("First line should be title, got: %s", lines[0])
	}
}

// Helper function to send a Gopher request
func sendGopherRequest(t *testing.T, port int, selector string) string {
	// Connect to server
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)), 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send selector
	_, err = conn.Write([]byte(selector + "\r\n"))
	if err != nil {
		t.Fatalf("Failed to send selector: %v", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	var response strings.Builder
	for {
		line, err := reader.ReadString('\n')
		response.WriteString(line)
		if err != nil || strings.HasSuffix(response.String(), ".\r\n") {
			break
		}
	}

	return response.String()
}
