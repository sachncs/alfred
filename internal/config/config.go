// Package config loads Alfred configuration from flags and environment
// variables. Environment variables override flags.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the runtime configuration for Alfred.
type Config struct {
	Port          int
	Version       string
	WorkspaceRoot string
	BearerToken   string
}

// Defaults returns a Config with production defaults.
func Defaults() Config {
	return Config{
		Port:          8899,
		Version:       "0.2.0-phase2",
		WorkspaceRoot: ".",
	}
}

// ParseFlags populates a Config from command-line flags.
// Call after flag.Parse() or use flag.CommandLine.
func ParseFlags() Config {
	cfg := Defaults()
	flag.IntVar(&cfg.Port, "port", cfg.Port, "HTTP server port")
	flag.StringVar(&cfg.WorkspaceRoot, "workspace-root", cfg.WorkspaceRoot, "workspace root directory")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		cfg.Port = -1 // sentinel: caller should print version and exit
	}
	return cfg
}

// ApplyEnv overrides Config fields with environment variables.
func (c *Config) ApplyEnv() {
	if v := os.Getenv("ALFRED_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Port = n
		}
	}
	if v := os.Getenv("ALFRED_WORKSPACE_ROOT"); v != "" {
		c.WorkspaceRoot = v
	}
	if v := os.Getenv("ALFRED_BEARER_TOKEN"); v != "" {
		c.BearerToken = v
	}
}
