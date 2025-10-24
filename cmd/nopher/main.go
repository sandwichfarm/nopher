package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sandwich/nopher/internal/aggregates"
	"github.com/sandwich/nopher/internal/config"
	"github.com/sandwich/nopher/internal/finger"
	"github.com/sandwich/nopher/internal/gemini"
	"github.com/sandwich/nopher/internal/gopher"
	"github.com/sandwich/nopher/internal/storage"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "manual"
)

func main() {
	// Define subcommands
	if len(os.Args) > 1 && os.Args[1] == "init" {
		handleInit()
		return
	}

	var (
		showVersion = flag.Bool("version", false, "Show version information")
		configPath  = flag.String("config", "", "Path to configuration file")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("nopher %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		fmt.Printf("  by:     %s\n", builtBy)
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Println("Nopher - Nostr to Gopher/Gemini/Finger Gateway")
		fmt.Println()
		fmt.Println("No configuration file specified. Use --config <path> to specify config.")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  nopher init              Generate example configuration")
		fmt.Println("  nopher --version         Show version information")
		fmt.Println("  nopher --config <path>   Start with configuration file")
		os.Exit(1)
	}

	// Load and validate configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting nopher %s\n", version)
	fmt.Printf("  Site: %s\n", cfg.Site.Title)
	fmt.Printf("  Operator: %s\n", cfg.Site.Operator)
	fmt.Printf("  Identity: %s\n", cfg.Identity.Npub)
	fmt.Println()

	// Run the application
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	fmt.Println("Initializing storage...")
	st, err := storage.New(ctx, &cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer st.Close()
	fmt.Printf("  Storage: %s initialized\n", cfg.Storage.Driver)

	// Initialize aggregates manager
	fmt.Println("Initializing aggregates manager...")
	aggMgr := aggregates.NewManager(st, cfg)
	fmt.Println("  Aggregates manager ready")

	// Initialize protocol servers
	var servers []interface{ Stop() error }

	// Gopher server
	if cfg.Protocols.Gopher.Enabled {
		fmt.Printf("Starting Gopher server on %s:%d...\n", cfg.Protocols.Gopher.Host, cfg.Protocols.Gopher.Port)
		gopherServer := gopher.New(&cfg.Protocols.Gopher, cfg, st, cfg.Protocols.Gopher.Host, aggMgr)
		if err := gopherServer.Start(); err != nil {
			return fmt.Errorf("failed to start Gopher server: %w", err)
		}
		servers = append(servers, gopherServer)
		fmt.Println("  Gopher server ready")
	}

	// Gemini server
	if cfg.Protocols.Gemini.Enabled {
		fmt.Printf("Starting Gemini server on %s:%d...\n", cfg.Protocols.Gemini.Host, cfg.Protocols.Gemini.Port)
		geminiServer, err := gemini.New(&cfg.Protocols.Gemini, cfg, st, cfg.Protocols.Gemini.Host, aggMgr)
		if err != nil {
			return fmt.Errorf("failed to create Gemini server: %w", err)
		}
		if err := geminiServer.Start(); err != nil {
			return fmt.Errorf("failed to start Gemini server: %w", err)
		}
		servers = append(servers, geminiServer)
		fmt.Println("  Gemini server ready")
	}

	// Finger server
	if cfg.Protocols.Finger.Enabled {
		fmt.Printf("Starting Finger server on port %d...\n", cfg.Protocols.Finger.Port)
		fingerServer := finger.New(&cfg.Protocols.Finger, cfg, st, aggMgr)
		if err := fingerServer.Start(); err != nil {
			return fmt.Errorf("failed to start Finger server: %w", err)
		}
		servers = append(servers, fingerServer)
		fmt.Println("  Finger server ready")
	}

	if len(servers) == 0 {
		return fmt.Errorf("no protocol servers enabled")
	}

	fmt.Println()
	fmt.Println("✓ All services started successfully!")
	fmt.Println()
	fmt.Println("Press Ctrl+C to shutdown gracefully...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("Shutting down gracefully...")

	// Stop all servers
	for _, server := range servers {
		if err := server.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping server: %v\n", err)
		}
	}

	fmt.Println("✓ Shutdown complete")
	return nil
}

func handleInit() {
	exampleConfig, err := config.GetExampleConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading example config: %v\n", err)
		os.Exit(1)
	}

	// Write to stdout
	fmt.Print(string(exampleConfig))
}
