package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	// Verify some key defaults
	if cfg.Protocols.Gopher.Port != 70 {
		t.Errorf("Expected Gopher port 70, got %d", cfg.Protocols.Gopher.Port)
	}

	if cfg.Protocols.Gemini.Port != 1965 {
		t.Errorf("Expected Gemini port 1965, got %d", cfg.Protocols.Gemini.Port)
	}

	if cfg.Storage.Driver != "sqlite" {
		t.Errorf("Expected default storage driver 'sqlite', got %s", cfg.Storage.Driver)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Logging.Level)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing npub",
			cfg:     &Config{Identity: Identity{Npub: ""}},
			wantErr: true,
			errMsg:  "npub is required",
		},
		{
			name:    "invalid npub prefix",
			cfg:     &Config{Identity: Identity{Npub: "nsec1test"}},
			wantErr: true,
			errMsg:  "must start with 'npub1'",
		},
		{
			name: "no protocols enabled",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: false},
					Gemini: GeminiProtocol{Enabled: false},
					Finger: FingerProtocol{Enabled: false},
				},
				Relays: Relays{Seeds: []string{"wss://relay.test"}},
			},
			wantErr: true,
			errMsg:  "at least one protocol must be enabled",
		},
		{
			name: "invalid port range",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 99999},
				},
				Relays: Relays{Seeds: []string{"wss://relay.test"}},
			},
			wantErr: true,
			errMsg:  "port must be between",
		},
		{
			name: "no relay seeds",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays: Relays{Seeds: []string{}},
			},
			wantErr: true,
			errMsg:  "at least one relay seed is required",
		},
		{
			name: "invalid relay seed protocol",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays: Relays{Seeds: []string{"https://relay.test"}},
			},
			wantErr: true,
			errMsg:  "must start with ws://",
		},
		{
			name: "invalid sync mode",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays: Relays{Seeds: []string{"wss://relay.test"}},
				Sync: Sync{
					Scope: SyncScope{Mode: "invalid"},
				},
			},
			wantErr: true,
			errMsg:  "invalid sync mode",
		},
		{
			name: "invalid storage driver",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays:  Relays{Seeds: []string{"wss://relay.test"}},
				Sync:    Sync{Scope: SyncScope{Mode: "self"}},
				Storage: Storage{Driver: "postgres"},
			},
			wantErr: true,
			errMsg:  "invalid storage driver",
		},
		{
			name: "invalid cache engine",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays:  Relays{Seeds: []string{"wss://relay.test"}},
				Sync:    Sync{Scope: SyncScope{Mode: "self"}},
				Storage: Storage{Driver: "sqlite"},
				Caching: Caching{Enabled: true, Engine: "invalid"},
			},
			wantErr: true,
			errMsg:  "invalid cache engine",
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays:  Relays{Seeds: []string{"wss://relay.test"}},
				Sync:    Sync{Scope: SyncScope{Mode: "self"}},
				Storage: Storage{Driver: "sqlite"},
				Caching: Caching{Enabled: true, Engine: "memory"},
				Logging: Logging{Level: "invalid"},
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "valid minimal config",
			cfg: &Config{
				Identity: Identity{Npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"},
				Protocols: Protocols{
					Gopher: GopherProtocol{Enabled: true, Port: 70},
				},
				Relays:  Relays{Seeds: []string{"wss://relay.test"}},
				Sync:    Sync{Scope: SyncScope{Mode: "self"}},
				Storage: Storage{Driver: "sqlite"},
				Caching: Caching{Enabled: false},
				Logging: Logging{Level: "info"},
				Display: Display{
					Limits: DisplayLimits{
						SummaryLength:     100,
						MaxContentLength:  5000,
						MaxThreadDepth:    10,
						MaxRepliesInFeed:  3,
						TruncateIndicator: "...",
					},
				},
				Behavior: Behavior{
					SortPreferences: SortPreferences{
						Notes:    "chronological",
						Articles: "chronological",
						Replies:  "chronological",
						Mentions: "chronological",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "valid config",
			content: `
site:
  title: "Test Site"
  description: "Test Description"
  operator: "Test Operator"

identity:
  npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"

protocols:
  gopher:
    enabled: true
    host: "localhost"
    port: 70
    bind: "0.0.0.0"

relays:
  seeds:
    - "wss://relay.test"
  policy:
    connect_timeout_ms: 5000
    max_concurrent_subs: 8
    backoff_ms: [500, 1500, 5000]

sync:
  scope:
    mode: "self"

storage:
  driver: "sqlite"

logging:
  level: "info"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Site.Title != "Test Site" {
					t.Errorf("Expected title 'Test Site', got %s", cfg.Site.Title)
				}
				if cfg.Identity.Npub != "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq" {
					t.Errorf("Expected npub 'npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq', got %s", cfg.Identity.Npub)
				}
				if !cfg.Protocols.Gopher.Enabled {
					t.Error("Expected Gopher to be enabled")
				}
			},
		},
		{
			name: "invalid yaml",
			content: `
site:
  title: "Test"
  invalid: [unclosed
`,
			wantErr: true,
		},
		{
			name: "missing required field",
			content: `
site:
  title: "Test"

protocols:
  gopher:
    enabled: true
    port: 70

relays:
  seeds:
    - "wss://relay.test"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test config file
			configPath := filepath.Join(tmpDir, tt.name+".yaml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected 'failed to read config file' error, got %v", err)
	}
}

func TestEnvOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("NOPHR_NSEC", "nsec1test")
	os.Setenv("NOPHR_REDIS_URL", "redis://localhost:6379")
	defer func() {
		os.Unsetenv("NOPHR_NSEC")
		os.Unsetenv("NOPHR_REDIS_URL")
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")
	content := `
site:
  title: "Test"

identity:
  npub: "npub1nq3zgtqruwhnz0xx40gh4a4fkamlr2sc7ke5wqs2s3nyv2fpy9esg4hdwq"

protocols:
  gopher:
    enabled: true
    port: 70

relays:
  seeds:
    - "wss://relay.test"

sync:
  scope:
    mode: "self"

storage:
  driver: "sqlite"

caching:
  enabled: true
  engine: "redis"

logging:
  level: "info"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Identity.Nsec != "nsec1test" {
		t.Errorf("Expected nsec from env 'nsec1test', got %s", cfg.Identity.Nsec)
	}

	if cfg.Caching.RedisURL != "redis://localhost:6379" {
		t.Errorf("Expected redis URL from env, got %s", cfg.Caching.RedisURL)
	}
}

func TestGetExampleConfig(t *testing.T) {
	content, err := GetExampleConfig()
	if err != nil {
		t.Fatalf("GetExampleConfig() failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("GetExampleConfig() returned empty content")
	}

	// Verify it's valid YAML
	if !strings.Contains(string(content), "site:") {
		t.Error("Example config doesn't contain expected YAML structure")
	}
}
