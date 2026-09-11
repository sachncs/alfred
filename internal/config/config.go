// Package config loads Alfred configuration from flags and environment
// variables. Environment variables override flags.
package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds the runtime configuration for Alfred.
type Config struct {
	Port          int
	Version       string
	WorkspaceRoot string
	BearerToken   string
	DBPath        string
}

// Defaults returns a Config with production defaults. Version is left
// empty so callers can populate it from internal/buildinfo.Version.
func Defaults() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Port:          8899,
		Version:       "",
		WorkspaceRoot: ".",
		DBPath:        filepath.Join(home, ".alfred", "alfred.db"),
	}
}

// RegisterFlags binds CLI flags to cfg and returns the --version pointer.
// Callers should flag.Parse() then check *showVersion.
func RegisterFlags(cfg *Config) *bool {
	flag.IntVar(&cfg.Port, "port", cfg.Port, "HTTP server port")
	flag.StringVar(&cfg.WorkspaceRoot, "workspace-root", cfg.WorkspaceRoot, "workspace root directory")
	flag.StringVar(&cfg.DBPath, "db-path", cfg.DBPath, "SQLite database path")
	return flag.Bool("version", false, "print version and exit")
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
	if v := os.Getenv("ALFRED_DB_PATH"); v != "" {
		c.DBPath = v
	}
}
