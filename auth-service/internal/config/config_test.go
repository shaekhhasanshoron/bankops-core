package config

import (
	"os"
	"testing"
)

// unset helper to redo all envs
func unset(keys ...string) {
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}

// TestLoad_Defaults_AreDevEnv checks default configuration are prod configurations
func TestLoad_Defaults_AreDevEnv(t *testing.T) {
	unset("AUTH_ENV", "AUTH_HTTP__ADDR")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	// defaults env=prod
	if cfg.Env != EnvProd {
		t.Fatalf("env=%q, want %q", cfg.Env, EnvProd)
	}

	// defaults addr=:8080
	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("addr=%q, want :8080", cfg.HTTP.Addr)
	}
	if cfg.Observability.MetricsConfig.Enabled {
		t.Fatalf("metrics should default to false")
	}
}

// TestLoad_EnvOverride checks overrides env with default
func TestLoad_EnvOverride(t *testing.T) {
	_ = os.Setenv("AUTH_HTTP__ADDR", ":9090")
	defer unset("AUTH_HTTP__ADDR")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("addr=%q, want :9090", cfg.HTTP.Addr)
	}
}
