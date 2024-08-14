package config

import (
	"flag"
	"runtime"
)

// Flags represents the command-line flags
type Flags struct {
	DebugLevel     int
	Concurrency    int
	ConfigFile     string
	SecretsFile    string
	HostnameFilter string
	Verbose        bool
	NoPanorama     bool
	ReportOnly     bool
}

// setupFlags sets up the flags without parsing them
func setupFlags(fs *flag.FlagSet, cfg *Flags) {
	fs.IntVar(&cfg.DebugLevel, "debug", 0, "Debug level: 0=INFO, 1=DEBUG")
	fs.IntVar(&cfg.Concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent operations")
	fs.StringVar(&cfg.ConfigFile, "config", "panorama.yaml", "Path to the Panorama configuration file")
	fs.StringVar(&cfg.SecretsFile, "secrets", ".secrets.yaml", "Path to the secrets file")
	fs.StringVar(&cfg.HostnameFilter, "filter", "", "Comma-separated list of hostname patterns to filter devices")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	fs.BoolVar(&cfg.NoPanorama, "nopanorama", false, "Use inventory.yaml instead of querying Panorama")
	fs.BoolVar(&cfg.ReportOnly, "reportonly", false, "Run in report-only mode without connecting to devices")
}

// ParseFlags parses command-line flags and returns a configuration object.
func ParseFlags() (*Flags, *Config) {
	cfg := &Flags{}
	setupFlags(flag.CommandLine, cfg)
	flag.Parse()

	config := &Config{
		HostnameFilter: cfg.HostnameFilter,
		ReportOnly:     cfg.ReportOnly,
	}

	return cfg, config
}
