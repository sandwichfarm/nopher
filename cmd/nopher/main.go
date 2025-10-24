package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sandwich/nopher/internal/config"
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
	fmt.Println("Protocol servers:")
	if cfg.Protocols.Gopher.Enabled {
		fmt.Printf("  - Gopher: %s:%d\n", cfg.Protocols.Gopher.Host, cfg.Protocols.Gopher.Port)
	}
	if cfg.Protocols.Gemini.Enabled {
		fmt.Printf("  - Gemini: %s:%d\n", cfg.Protocols.Gemini.Host, cfg.Protocols.Gemini.Port)
	}
	if cfg.Protocols.Finger.Enabled {
		fmt.Printf("  - Finger: port %d\n", cfg.Protocols.Finger.Port)
	}
	fmt.Println()
	fmt.Println("Configuration loaded successfully!")
	fmt.Println("(Server implementation pending - Phase 1 complete)")
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
