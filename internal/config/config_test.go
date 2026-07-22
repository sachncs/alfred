package config

import (
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Port != 8899 {
		t.Errorf("default port = %d, want 8899", cfg.Port)
	}
	if cfg.WorkspaceRoot != "." {
		t.Errorf("default workspace root = %q, want \".\"", cfg.WorkspaceRoot)
	}
	if cfg.Version == "" {
		t.Error("default version should not be empty")
	}
}

func TestApplyEnvPort(t *testing.T) {
	cfg := Defaults()
	_ = os.Setenv("ALFRED_PORT", "9999")
	defer func() { _ = os.Unsetenv("ALFRED_PORT") }()

	cfg.ApplyEnv()
	if cfg.Port != 9999 {
		t.Errorf("port after env = %d, want 9999", cfg.Port)
	}
}

func TestApplyEnvWorkspaceRoot(t *testing.T) {
	cfg := Defaults()
	_ = os.Setenv("ALFRED_WORKSPACE_ROOT", "/tmp/test-workspace")
	defer func() { _ = os.Unsetenv("ALFRED_WORKSPACE_ROOT") }()

	cfg.ApplyEnv()
	if cfg.WorkspaceRoot != "/tmp/test-workspace" {
		t.Errorf("workspace root after env = %q, want /tmp/test-workspace", cfg.WorkspaceRoot)
	}
}

func TestApplyEnvBearerToken(t *testing.T) {
	cfg := Defaults()
	_ = os.Setenv("ALFRED_BEARER_TOKEN", "secret-token")
	defer func() { _ = os.Unsetenv("ALFRED_BEARER_TOKEN") }()

	cfg.ApplyEnv()
	if cfg.BearerToken != "secret-token" {
		t.Errorf("bearer token after env = %q, want secret-token", cfg.BearerToken)
	}
}

func TestApplyEnvInvalidPortIgnored(t *testing.T) {
	cfg := Defaults()
	cfg.Port = 8899
	_ = os.Setenv("ALFRED_PORT", "not-a-number")
	defer func() { _ = os.Unsetenv("ALFRED_PORT") }()

	cfg.ApplyEnv()
	if cfg.Port != 8899 {
		t.Errorf("invalid port env should be ignored, got %d", cfg.Port)
	}
}

func TestApplyEnvEmptyVarsNoop(t *testing.T) {
	cfg := Defaults()
	cfg.Port = 8899
	cfg.WorkspaceRoot = "."
	cfg.ApplyEnv()
	if cfg.Port != 8899 {
		t.Errorf("port changed without env var: %d", cfg.Port)
	}
	if cfg.WorkspaceRoot != "." {
		t.Errorf("workspace root changed without env var: %q", cfg.WorkspaceRoot)
	}
}
