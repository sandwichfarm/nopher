package gemini

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/sandwich/nopher/internal/aggregates"
	"github.com/sandwich/nopher/internal/config"
	"github.com/sandwich/nopher/internal/storage"
)

func TestGeminiProtocol(t *testing.T) {
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

	geminiCfg := &config.GeminiProtocol{
		Enabled: true,
		Host:    "localhost",
		Port:    11965, // Use non-standard port for testing
		TLS: config.GeminiTLS{
			AutoGenerate: true,
		},
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
	server, err := New(geminiCfg, cfg, st, "localhost", aggMgr)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Test 1: Root path
	t.Run("RootPath", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "gemini://localhost/")
		if !strings.Contains(response, "20 ") {
			t.Errorf("Root response should have status 20, got: %s", response[:20])
		}
		if !strings.Contains(response, "Nopher") {
			t.Errorf("Root response should contain 'Nopher'")
		}
	})

	// Test 2: Outbox path
	t.Run("OutboxPath", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "gemini://localhost/outbox")
		if !strings.Contains(response, "20 ") {
			t.Errorf("Outbox response should have status 20")
		}
		if !strings.Contains(response, "Outbox") {
			t.Errorf("Outbox response should contain 'Outbox'")
		}
	})

	// Test 3: Inbox path
	t.Run("InboxPath", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "gemini://localhost/inbox")
		if !strings.Contains(response, "20 ") {
			t.Errorf("Inbox response should have status 20")
		}
		if !strings.Contains(response, "Inbox") {
			t.Errorf("Inbox response should contain 'Inbox'")
		}
	})

	// Test 4: Diagnostics path
	t.Run("DiagnosticsPath", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "gemini://localhost/diagnostics")
		if !strings.Contains(response, "20 ") {
			t.Errorf("Diagnostics response should have status 20")
		}
		if !strings.Contains(response, "Diagnostics") {
			t.Errorf("Diagnostics response should contain 'Diagnostics'")
		}
	})

	// Test 5: Search path (should request input)
	t.Run("SearchPath", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "gemini://localhost/search")
		if !strings.Contains(response, "10 ") {
			t.Errorf("Search without query should request input (status 10), got: %s", response[:20])
		}
	})

	// Test 6: Invalid path
	t.Run("InvalidPath", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "gemini://localhost/invalid")
		if !strings.Contains(response, "51 ") {
			t.Errorf("Invalid path should return status 51 (not found), got: %s", response[:20])
		}
	})

	// Test 7: Invalid URL
	t.Run("InvalidURL", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "not-a-url")
		if !strings.Contains(response, "59 ") {
			t.Errorf("Invalid URL should return status 59 (bad request), got: %s", response[:20])
		}
	})

	// Test 8: Non-gemini scheme
	t.Run("NonGeminiScheme", func(t *testing.T) {
		response := sendGeminiRequest(t, geminiCfg.Port, "http://localhost/")
		if !strings.Contains(response, "53 ") {
			t.Errorf("Non-gemini scheme should return status 53 (proxy refused), got: %s", response[:20])
		}
	})
}

func TestGeminiResponseFormat(t *testing.T) {
	// Test success response
	t.Run("SuccessResponse", func(t *testing.T) {
		response := FormatSuccessResponse("# Hello\n\nTest content")
		responseStr := string(response)

		if !strings.HasPrefix(responseStr, "20 ") {
			t.Errorf("Success response should start with '20 '")
		}
		if !strings.Contains(responseStr, "text/gemini") {
			t.Errorf("Success response should contain 'text/gemini'")
		}
		if !strings.Contains(responseStr, "\r\n") {
			t.Errorf("Response should contain CRLF")
		}
	})

	// Test error response
	t.Run("ErrorResponse", func(t *testing.T) {
		response := FormatErrorResponse(StatusNotFound, "Not found")
		responseStr := string(response)

		if !strings.HasPrefix(responseStr, "51 ") {
			t.Errorf("Error response should start with '51 '")
		}
		if !strings.Contains(responseStr, "Not found") {
			t.Errorf("Error response should contain error message")
		}
	})

	// Test input response
	t.Run("InputResponse", func(t *testing.T) {
		response := FormatInputResponse("Enter query:", false)
		responseStr := string(response)

		if !strings.HasPrefix(responseStr, "10 ") {
			t.Errorf("Input response should start with '10 '")
		}
		if !strings.Contains(responseStr, "Enter query:") {
			t.Errorf("Input response should contain prompt")
		}
	})

	// Test sensitive input response
	t.Run("SensitiveInputResponse", func(t *testing.T) {
		response := FormatInputResponse("Enter password:", true)
		responseStr := string(response)

		if !strings.HasPrefix(responseStr, "11 ") {
			t.Errorf("Sensitive input response should start with '11 '")
		}
	})

	// Test redirect response
	t.Run("RedirectResponse", func(t *testing.T) {
		response := FormatRedirectResponse("gemini://new.host/path", false)
		responseStr := string(response)

		if !strings.HasPrefix(responseStr, "30 ") {
			t.Errorf("Temporary redirect should start with '30 '")
		}

		response = FormatRedirectResponse("gemini://new.host/path", true)
		responseStr = string(response)

		if !strings.HasPrefix(responseStr, "31 ") {
			t.Errorf("Permanent redirect should start with '31 '")
		}
	})
}

func TestRendererOutput(t *testing.T) {
	renderer := NewRenderer()

	// Test home rendering
	t.Run("HomeRendering", func(t *testing.T) {
		home := renderer.RenderHome()

		if !strings.Contains(home, "# Nopher") {
			t.Errorf("Home should contain title")
		}
		if !strings.Contains(home, "=> /outbox") {
			t.Errorf("Home should contain outbox link")
		}
		if !strings.Contains(home, "=> /inbox") {
			t.Errorf("Home should contain inbox link")
		}
	})

	// Test note list rendering
	t.Run("NoteListRendering", func(t *testing.T) {
		notes := []*aggregates.EnrichedEvent{}
		gemtext := renderer.RenderNoteList(notes, "Test List", "gemini://localhost/")

		if !strings.Contains(gemtext, "# Test List") {
			t.Errorf("Note list should contain title")
		}
		if !strings.Contains(gemtext, "No notes yet") {
			t.Errorf("Empty note list should say 'No notes yet'")
		}
	})
}

// Helper function to send a Gemini request
func sendGeminiRequest(t *testing.T, port int, url string) string {
	// Create TLS config that accepts self-signed certs
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Connect to server
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 5 * time.Second},
		"tcp",
		net.JoinHostPort("localhost", fmt.Sprintf("%d", port)),
		tlsConfig,
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send URL
	_, err = conn.Write([]byte(url + "\r\n"))
	if err != nil {
		t.Fatalf("Failed to send URL: %v", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	var response strings.Builder

	// Read until connection closes or timeout
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		line, err := reader.ReadString('\n')
		response.WriteString(line)
		if err != nil {
			break
		}
	}

	return response.String()
}
