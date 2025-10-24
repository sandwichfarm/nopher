package config

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed example.yaml
var exampleConfig embed.FS

// Config represents the complete Nopher configuration
type Config struct {
	Site       Site       `yaml:"site"`
	Identity   Identity   `yaml:"identity"`
	Protocols  Protocols  `yaml:"protocols"`
	Relays     Relays     `yaml:"relays"`
	Discovery  Discovery  `yaml:"discovery"`
	Sync       Sync       `yaml:"sync"`
	Inbox      Inbox      `yaml:"inbox"`
	Outbox     Outbox     `yaml:"outbox"`
	Storage    Storage    `yaml:"storage"`
	Rendering  Rendering  `yaml:"rendering"`
	Caching    Caching    `yaml:"caching"`
	Logging    Logging    `yaml:"logging"`
	Layout     Layout     `yaml:"layout"`
}

// Site contains site metadata
type Site struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Operator    string `yaml:"operator"`
}

// Identity contains Nostr identity information
type Identity struct {
	Npub string `yaml:"npub"` // Public key from file
	Nsec string `yaml:"-"`    // Private key from env only, never serialized
}

// Protocols contains protocol server configurations
type Protocols struct {
	Gopher GopherProtocol `yaml:"gopher"`
	Gemini GeminiProtocol `yaml:"gemini"`
	Finger FingerProtocol `yaml:"finger"`
}

// GopherProtocol contains Gopher server settings
type GopherProtocol struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Bind    string `yaml:"bind"`
}

// GeminiProtocol contains Gemini server settings
type GeminiProtocol struct {
	Enabled bool      `yaml:"enabled"`
	Host    string    `yaml:"host"`
	Port    int       `yaml:"port"`
	Bind    string    `yaml:"bind"`
	TLS     GeminiTLS `yaml:"tls"`
}

// GeminiTLS contains TLS configuration for Gemini
type GeminiTLS struct {
	CertPath     string `yaml:"cert_path"`
	KeyPath      string `yaml:"key_path"`
	AutoGenerate bool   `yaml:"auto_generate"`
}

// FingerProtocol contains Finger server settings
type FingerProtocol struct {
	Enabled  bool   `yaml:"enabled"`
	Port     int    `yaml:"port"`
	Bind     string `yaml:"bind"`
	MaxUsers int    `yaml:"max_users"`
}

// Relays contains relay configuration
type Relays struct {
	Seeds  []string    `yaml:"seeds"`
	Policy RelayPolicy `yaml:"policy"`
}

// RelayPolicy contains relay connection policies
type RelayPolicy struct {
	ConnectTimeoutMs   int   `yaml:"connect_timeout_ms"`
	MaxConcurrentSubs  int   `yaml:"max_concurrent_subs"`
	BackoffMs          []int `yaml:"backoff_ms"`
}

// Discovery contains relay discovery settings
type Discovery struct {
	RefreshSeconds      int  `yaml:"refresh_seconds"`
	UseOwnerHints       bool `yaml:"use_owner_hints"`
	UseAuthorHints      bool `yaml:"use_author_hints"`
	FallbackToSeeds     bool `yaml:"fallback_to_seeds"`
	MaxRelaysPerAuthor  int  `yaml:"max_relays_per_author"`
}

// Sync contains synchronization settings
type Sync struct {
	Enabled   bool      `yaml:"enabled"`
	Kinds     []int     `yaml:"kinds"`
	Scope     SyncScope `yaml:"scope"`
	Retention Retention `yaml:"retention"`
}

// SyncScope defines synchronization scope
type SyncScope struct {
	Mode                  string   `yaml:"mode"` // self|following|mutual|foaf
	Depth                 int      `yaml:"depth"`
	IncludeDirectMentions bool     `yaml:"include_direct_mentions"`
	IncludeThreadsOfMine  bool     `yaml:"include_threads_of_mine"`
	MaxAuthors            int      `yaml:"max_authors"`
	AllowlistPubkeys      []string `yaml:"allowlist_pubkeys"`
	DenylistPubkeys       []string `yaml:"denylist_pubkeys"`
}

// Retention defines data retention policies
type Retention struct {
	KeepDays      int  `yaml:"keep_days"`
	PruneOnStart  bool `yaml:"prune_on_start"`
}

// Inbox contains inbox aggregation settings
type Inbox struct {
	IncludeReplies    bool          `yaml:"include_replies"`
	IncludeReactions  bool          `yaml:"include_reactions"`
	IncludeZaps       bool          `yaml:"include_zaps"`
	GroupByThread     bool          `yaml:"group_by_thread"`
	CollapseReposts   bool          `yaml:"collapse_reposts"`
	NoiseFilters      NoiseFilters  `yaml:"noise_filters"`
}

// NoiseFilters defines filtering rules for inbox
type NoiseFilters struct {
	MinZapSats            int      `yaml:"min_zap_sats"`
	AllowedReactionChars  []string `yaml:"allowed_reaction_chars"`
}

// Outbox contains outbox/publishing settings
type Outbox struct {
	Publish   PublishSettings `yaml:"publish"`
	DraftDir  string          `yaml:"draft_dir"`
	AutoSign  bool            `yaml:"auto_sign"`
}

// PublishSettings defines what to publish
type PublishSettings struct {
	Notes     bool `yaml:"notes"`
	Reactions bool `yaml:"reactions"`
	Zaps      bool `yaml:"zaps"`
}

// Storage contains storage backend settings
type Storage struct {
	Driver        string `yaml:"driver"` // sqlite|lmdb
	SQLitePath    string `yaml:"sqlite_path"`
	LMDBPath      string `yaml:"lmdb_path"`
	LMDBMaxSizeMB int    `yaml:"lmdb_max_size_mb"`
}

// Rendering contains protocol-specific rendering options
type Rendering struct {
	Gopher GopherRendering `yaml:"gopher"`
	Gemini GeminiRendering `yaml:"gemini"`
	Finger FingerRendering `yaml:"finger"`
}

// GopherRendering contains Gopher rendering options
type GopherRendering struct {
	MaxLineLength  int    `yaml:"max_line_length"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	DateFormat     string `yaml:"date_format"`
	ThreadIndent   string `yaml:"thread_indent"`
}

// GeminiRendering contains Gemini rendering options
type GeminiRendering struct {
	MaxLineLength  int    `yaml:"max_line_length"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	Emoji          bool   `yaml:"emoji"`
}

// FingerRendering contains Finger rendering options
type FingerRendering struct {
	PlanSource       string `yaml:"plan_source"`
	RecentNotesCount int    `yaml:"recent_notes_count"`
}

// Caching contains caching configuration
type Caching struct {
	Enabled    bool              `yaml:"enabled"`
	Engine     string            `yaml:"engine"` // memory|redis
	RedisURL   string            `yaml:"redis_url"`
	TTL        CacheTTL          `yaml:"ttl"`
	Aggregates AggregatesCaching `yaml:"aggregates"`
	Overrides  map[string]interface{} `yaml:"overrides,omitempty"`
}

// CacheTTL contains TTL settings for different cache types
type CacheTTL struct {
	Sections map[string]int `yaml:"sections"`
	Render   map[string]int `yaml:"render"`
}

// AggregatesCaching contains aggregate caching settings
type AggregatesCaching struct {
	Enabled                    bool `yaml:"enabled"`
	UpdateOnIngest             bool `yaml:"update_on_ingest"`
	ReconcilerIntervalSeconds  int  `yaml:"reconciler_interval_seconds"`
}

// Logging contains logging configuration
type Logging struct {
	Level string `yaml:"level"` // debug|info|warn|error
}

// Layout contains layout and section definitions
type Layout struct {
	Sections map[string]interface{} `yaml:"sections,omitempty"`
	Pages    map[string]interface{} `yaml:"pages,omitempty"`
}

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	if err := applyEnvOverrides(&cfg); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Validate configuration
	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *Config) error {
	// NOPHER_NSEC is the most important one - never in file
	if nsec := os.Getenv("NOPHER_NSEC"); nsec != "" {
		cfg.Identity.Nsec = nsec
	}

	// Redis URL from env if using redis
	if redisURL := os.Getenv("NOPHER_REDIS_URL"); redisURL != "" {
		cfg.Caching.RedisURL = redisURL
	}

	// Allow overriding any config via NOPHER_ prefix
	// This is a simplified implementation - full version would use reflection
	// to handle all nested fields automatically

	return nil
}

// GetExampleConfig returns the embedded example configuration
func GetExampleConfig() ([]byte, error) {
	return exampleConfig.ReadFile("example.yaml")
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		Site: Site{
			Title:       "My Nostr Site",
			Description: "Personal Nostr gateway",
			Operator:    "Anonymous",
		},
		Identity: Identity{
			Npub: "",
		},
		Protocols: Protocols{
			Gopher: GopherProtocol{
				Enabled: true,
				Host:    "localhost",
				Port:    70,
				Bind:    "0.0.0.0",
			},
			Gemini: GeminiProtocol{
				Enabled: true,
				Host:    "localhost",
				Port:    1965,
				Bind:    "0.0.0.0",
				TLS: GeminiTLS{
					CertPath:     "./certs/cert.pem",
					KeyPath:      "./certs/key.pem",
					AutoGenerate: true,
				},
			},
			Finger: FingerProtocol{
				Enabled:  true,
				Port:     79,
				Bind:     "0.0.0.0",
				MaxUsers: 100,
			},
		},
		Relays: Relays{
			Seeds: []string{
				"wss://relay.damus.io",
				"wss://relay.nostr.band",
				"wss://nos.lol",
			},
			Policy: RelayPolicy{
				ConnectTimeoutMs:  5000,
				MaxConcurrentSubs: 8,
				BackoffMs:         []int{500, 1500, 5000},
			},
		},
		Discovery: Discovery{
			RefreshSeconds:     900,
			UseOwnerHints:      true,
			UseAuthorHints:     true,
			FallbackToSeeds:    true,
			MaxRelaysPerAuthor: 8,
		},
		Sync: Sync{
			Kinds: []int{0, 1, 3, 6, 7, 9735, 30023, 10002},
			Scope: SyncScope{
				Mode:                  "foaf",
				Depth:                 2,
				IncludeDirectMentions: true,
				IncludeThreadsOfMine:  true,
				MaxAuthors:            5000,
				AllowlistPubkeys:      []string{},
				DenylistPubkeys:       []string{},
			},
			Retention: Retention{
				KeepDays:     365,
				PruneOnStart: true,
			},
		},
		Inbox: Inbox{
			IncludeReplies:   true,
			IncludeReactions: true,
			IncludeZaps:      true,
			GroupByThread:    true,
			CollapseReposts:  true,
			NoiseFilters: NoiseFilters{
				MinZapSats:           1,
				AllowedReactionChars: []string{"+"},
			},
		},
		Outbox: Outbox{
			Publish: PublishSettings{
				Notes:     true,
				Reactions: false,
				Zaps:      false,
			},
			DraftDir: "./content",
			AutoSign: false,
		},
		Storage: Storage{
			Driver:        "sqlite",
			SQLitePath:    "./data/nopher.db",
			LMDBPath:      "./data/nopher.lmdb",
			LMDBMaxSizeMB: 10240,
		},
		Rendering: Rendering{
			Gopher: GopherRendering{
				MaxLineLength:  70,
				ShowTimestamps: true,
				DateFormat:     "2006-01-02 15:04 MST",
				ThreadIndent:   "  ",
			},
			Gemini: GeminiRendering{
				MaxLineLength:  80,
				ShowTimestamps: true,
				Emoji:          true,
			},
			Finger: FingerRendering{
				PlanSource:       "kind_0",
				RecentNotesCount: 5,
			},
		},
		Caching: Caching{
			Enabled:  true,
			Engine:   "memory",
			RedisURL: "",
			TTL: CacheTTL{
				Sections: map[string]int{
					"notes":        60,
					"comments":     30,
					"articles":     300,
					"interactions": 10,
				},
				Render: map[string]int{
					"gopher_menu":     300,
					"gemini_page":     300,
					"finger_response": 60,
					"kind_1":          86400,
					"kind_30023":      604800,
					"kind_0":          3600,
					"kind_3":          600,
				},
			},
			Aggregates: AggregatesCaching{
				Enabled:                   true,
				UpdateOnIngest:            true,
				ReconcilerIntervalSeconds: 900,
			},
		},
		Logging: Logging{
			Level: "info",
		},
		Layout: Layout{
			Sections: make(map[string]interface{}),
			Pages:    make(map[string]interface{}),
		},
	}
}

// validLogLevels defines allowed log levels
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// validSyncModes defines allowed sync modes
var validSyncModes = map[string]bool{
	"self":      true,
	"following": true,
	"mutual":    true,
	"foaf":      true,
}

// validStorageDrivers defines allowed storage drivers
var validStorageDrivers = map[string]bool{
	"sqlite": true,
	"lmdb":   true,
}

// validCacheEngines defines allowed cache engines
var validCacheEngines = map[string]bool{
	"memory": true,
	"redis":  true,
}

// Validate checks if a configuration is valid
func Validate(cfg *Config) error {
	// Validate identity
	if cfg.Identity.Npub == "" {
		return fmt.Errorf("identity.npub is required")
	}
	if !strings.HasPrefix(cfg.Identity.Npub, "npub1") {
		return fmt.Errorf("identity.npub must start with 'npub1'")
	}

	// Validate at least one protocol is enabled
	if !cfg.Protocols.Gopher.Enabled && !cfg.Protocols.Gemini.Enabled && !cfg.Protocols.Finger.Enabled {
		return fmt.Errorf("at least one protocol must be enabled")
	}

	// Validate ports
	if cfg.Protocols.Gopher.Enabled && (cfg.Protocols.Gopher.Port < 1 || cfg.Protocols.Gopher.Port > 65535) {
		return fmt.Errorf("gopher port must be between 1 and 65535")
	}
	if cfg.Protocols.Gemini.Enabled && (cfg.Protocols.Gemini.Port < 1 || cfg.Protocols.Gemini.Port > 65535) {
		return fmt.Errorf("gemini port must be between 1 and 65535")
	}
	if cfg.Protocols.Finger.Enabled && (cfg.Protocols.Finger.Port < 1 || cfg.Protocols.Finger.Port > 65535) {
		return fmt.Errorf("finger port must be between 1 and 65535")
	}

	// Validate relay seeds
	if len(cfg.Relays.Seeds) == 0 {
		return fmt.Errorf("at least one relay seed is required")
	}
	for _, seed := range cfg.Relays.Seeds {
		if !strings.HasPrefix(seed, "wss://") && !strings.HasPrefix(seed, "ws://") {
			return fmt.Errorf("relay seed must start with ws:// or wss://: %s", seed)
		}
	}

	// Validate sync mode
	if !validSyncModes[cfg.Sync.Scope.Mode] {
		return fmt.Errorf("invalid sync mode: %s (must be one of: self, following, mutual, foaf)", cfg.Sync.Scope.Mode)
	}

	// Validate storage driver
	if !validStorageDrivers[cfg.Storage.Driver] {
		return fmt.Errorf("invalid storage driver: %s (must be one of: sqlite, lmdb)", cfg.Storage.Driver)
	}

	// Validate cache engine
	if cfg.Caching.Enabled && !validCacheEngines[cfg.Caching.Engine] {
		return fmt.Errorf("invalid cache engine: %s (must be one of: memory, redis)", cfg.Caching.Engine)
	}

	// Validate log level
	if !validLogLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", cfg.Logging.Level)
	}

	return nil
}
