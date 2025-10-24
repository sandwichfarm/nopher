package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "manual"
)

func main() {
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
		fmt.Println("For more information:")
		fmt.Println("  nopher --version")
		fmt.Println("  nopher --help")
		os.Exit(1)
	}

	fmt.Printf("Starting nopher with config: %s\n", *configPath)
	fmt.Println("(Server implementation pending - Phase 0 bootstrap)")
	os.Exit(0)
}
